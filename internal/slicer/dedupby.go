package slicer

import (
	"fmt"
	"regexp"
	"sync"
)

// DedupByExtractor extracts a key from a line for deduplication purposes.
type DedupByExtractor func(line string) string

// DedupByFilter deduplicates lines based on a extracted key rather than the
// full line content. This allows, for example, deduplicating log lines that
// share the same message but differ in timestamp.
type DedupByFilter struct {
	extract DedupByExtractor
	seen   map[uint64]struct{}
	mu     sync.Mutex
}

// NewDedupByFilter creates a DedupByFilter using the provided key extractor.
// If extract is nil, the whole line is used as the key (equivalent to DedupFilter).
func NewDedupByFilter(extract DedupByExtractor) *DedupByFilter {
	if extract == nil {
		extract = func(line string) string { return line }
	}
	return &DedupByFilter{
		extract: extract,
		seen:    make(map[uint64]struct{}),
	}
}

// NewRegexDedupByFilter creates a DedupByFilter that extracts the first
// capture group of pattern as the dedup key. If there is no capture group,
// the full match is used. Returns an error for invalid patterns.
func NewRegexDedupByFilter(pattern string) (*DedupByFilter, error) {
	if pattern == "" {
		return nil, fmt.Errorf("dedupby: pattern must not be empty")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("dedupby: invalid pattern: %w", err)
	}
	extract := func(line string) string {
		m := re.FindStringSubmatch(line)
		if m == nil {
			return line
		}
		if len(m) > 1 {
			return m[1]
		}
		return m[0]
	}
	return NewDedupByFilter(extract), nil
}

// Accept returns true if the line's key has not been seen before.
func (d *DedupByFilter) Accept(line string) bool {
	key := d.extract(line)
	h := hash(key)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.seen[h]; exists {
		return false
	}
	d.seen[h] = struct{}{}
	return true
}

// Reset clears the seen-key set so the filter starts fresh.
func (d *DedupByFilter) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[uint64]struct{})
}

// NewDedupByStage returns a pipeline stage that deduplicates lines by key.
func NewDedupByStage(f *DedupByFilter) func(in <-chan string) <-chan string {
	return func(in <-chan string) <-chan string {
		out := make(chan string)
		go func() {
			defer close(out)
			for line := range in {
				if f.Accept(line) {
					out <- line
				}
			}
		}()
		return out
	}
}
