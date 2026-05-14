package slicer

import (
	"fmt"
	"sync/atomic"
)

// LineNumFilter tracks line numbers and allows filtering by line number range.
type LineNumFilter struct {
	start   int64
	end     int64
	counter int64
}

// NewLineNumFilter creates a filter that passes only lines within [start, end] (1-based, inclusive).
// A zero end means no upper bound.
func NewLineNumFilter(start, end int64) *LineNumFilter {
	if start < 1 {
		start = 1
	}
	return &LineNumFilter{
		start: start,
		end:   end,
	}
}

// Accept increments the internal counter and returns true if the current line
// falls within the configured range.
func (f *LineNumFilter) Accept(_ string) bool {
	n := atomic.AddInt64(&f.counter, 1)
	if n < f.start {
		return false
	}
	if f.end > 0 && n > f.end {
		return false
	}
	return true
}

// Current returns the current line count.
func (f *LineNumFilter) Current() int64 {
	return atomic.LoadInt64(&f.counter)
}

// Reset resets the line counter to zero.
func (f *LineNumFilter) Reset() {
	atomic.StoreInt64(&f.counter, 0)
}

// Summary returns a human-readable description of the filter range.
func (f *LineNumFilter) Summary() string {
	if f.end == 0 {
		return fmt.Sprintf("lines %d+", f.start)
	}
	return fmt.Sprintf("lines %d-%d", f.start, f.end)
}
