package slicer

import (
	"context"
	"fmt"
	"regexp"
	"sort"
)

// CountByExtractor extracts a key from a log line for frequency counting.
type CountByExtractor func(line string) (string, bool)

// CountByResult holds a key and its occurrence count.
type CountByResult struct {
	Key   string
	Count int
}

// CountBy accumulates line counts grouped by a key extracted from each line.
type CountBy struct {
	extract CountByExtractor
	counts  map[string]int
}

// NewCountBy creates a CountBy using the provided extractor function.
func NewCountBy(extract CountByExtractor) *CountBy {
	if extract == nil {
		extract = func(line string) (string, bool) { return line, true }
	}
	return &CountBy{
		extract: extract,
		counts:  make(map[string]int),
	}
}

// NewRegexCountBy creates a CountBy that groups lines by the first capture
// group of re. Lines that do not match are silently skipped.
func NewRegexCountBy(pattern string) (*CountBy, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("countby: invalid pattern: %w", err)
	}
	extract := func(line string) (string, bool) {
		m := re.FindStringSubmatch(line)
		if m == nil {
			return "", false
		}
		if len(m) > 1 {
			return m[1], true
		}
		return m[0], true
	}
	return NewCountBy(extract), nil
}

// Add records a line into the appropriate bucket.
func (c *CountBy) Add(line string) {
	key, ok := c.extract(line)
	if !ok {
		return
	}
	c.counts[key]++
}

// Results returns counts sorted descending by frequency.
func (c *CountBy) Results() []CountByResult {
	out := make([]CountByResult, 0, len(c.counts))
	for k, v := range c.counts {
		out = append(out, CountByResult{Key: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// Reset clears all accumulated counts.
func (c *CountBy) Reset() {
	c.counts = make(map[string]int)
}

// NewCountByStage reads from in, feeds every line into cb, and forwards each
// line unchanged to the returned channel. When in closes (or ctx is done) the
// output channel is closed.
func NewCountByStage(ctx context.Context, in <-chan string, cb *CountBy) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				cb.Add(line)
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
