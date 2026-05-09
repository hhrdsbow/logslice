package slicer

import (
	"context"
	"fmt"
	"regexp"
)

// RegexSplitter splits the input line stream into segments whenever a line
// matches the given boundary pattern. Each segment starts at the matching
// line and ends just before the next match (or at EOF).
type RegexSplitter struct {
	re      *regexp.Regexp
	invert  bool
	maxSegs int
}

// RegexSplitOption configures a RegexSplitter.
type RegexSplitOption func(*RegexSplitter)

// WithSplitInvert inverts the boundary match: a new segment starts when a
// line does NOT match the pattern.
func WithSplitInvert(invert bool) RegexSplitOption {
	return func(r *RegexSplitter) { r.invert = invert }
}

// WithMaxSegments caps the number of segments emitted (0 = unlimited).
func WithMaxSegments(n int) RegexSplitOption {
	return func(r *RegexSplitter) {
		if n < 0 {
			n = 0
		}
		r.maxSegs = n
	}
}

// NewRegexSplitter creates a RegexSplitter that uses pattern as the segment
// boundary detector. Returns an error if pattern is not a valid regexp.
func NewRegexSplitter(pattern string, opts ...RegexSplitOption) (*RegexSplitter, error) {
	if pattern == "" {
		return nil, fmt.Errorf("regex_split: pattern must not be empty")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("regex_split: invalid pattern: %w", err)
	}
	rs := &RegexSplitter{re: re}
	for _, o := range opts {
		o(rs)
	}
	return rs, nil
}

// isBoundary reports whether line triggers a new segment.
func (rs *RegexSplitter) isBoundary(line string) bool {
	matched := rs.re.MatchString(line)
	if rs.invert {
		return !matched
	}
	return matched
}

// Split reads lines from in, groups them into segments, and sends each
// completed segment on the returned channel. The channel is closed when in
// is exhausted or ctx is cancelled.
func (rs *RegexSplitter) Split(ctx context.Context, in <-chan string) <-chan []string {
	out := make(chan []string)
	go func() {
		defer close(out)
		var current []string
		segCount := 0
		flush := func() bool {
			if len(current) == 0 {
				return true
			}
			select {
			case out <- current:
			case <-ctx.Done():
				return false
			}
			segCount++
			current = nil
			return true
		}
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					flush()
					return
				}
				if rs.isBoundary(line) && len(current) > 0 {
					if !flush() {
						return
					}
					if rs.maxSegs > 0 && segCount >= rs.maxSegs {
						return
					}
				}
				current = append(current, line)
			}
		}
	}()
	return out
}
