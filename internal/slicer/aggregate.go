package slicer

import (
	"fmt"
	"strings"
	"sync"
)

// AggregateFunc reduces a slice of lines into a single summary line.
type AggregateFunc func(lines []string) string

// Aggregator collects lines into fixed-size buckets and emits one summary
// line per bucket using the provided AggregateFunc.
type Aggregator struct {
	mu      sync.Mutex
	bucket  []string
	size    int
	aggrFn  AggregateFunc
}

// NewAggregator creates an Aggregator that groups every n lines and reduces
// them with fn. n is clamped to a minimum of 1.
func NewAggregator(n int, fn AggregateFunc) *Aggregator {
	if n < 1 {
		n = 1
	}
	if fn == nil {
		fn = ConcatAggregate
	}
	return &Aggregator{size: n, aggrFn: fn}
}

// Add appends a line to the current bucket. When the bucket is full it is
// flushed and the summary line is returned together with true.
func (a *Aggregator) Add(line string) (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bucket = append(a.bucket, line)
	if len(a.bucket) >= a.size {
		summary := a.aggrFn(a.bucket)
		a.bucket = a.bucket[:0]
		return summary, true
	}
	return "", false
}

// Flush emits a summary for any remaining lines in the bucket. Returns empty
// string and false when the bucket is already empty.
func (a *Aggregator) Flush() (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.bucket) == 0 {
		return "", false
	}
	summary := a.aggrFn(a.bucket)
	a.bucket = a.bucket[:0]
	return summary, true
}

// ConcatAggregate joins lines with " | ".
func ConcatAggregate(lines []string) string {
	return strings.Join(lines, " | ")
}

// CountAggregate returns a line of the form "[N lines]".
func CountAggregate(lines []string) string {
	return fmt.Sprintf("[%d lines]", len(lines))
}

// FirstAggregate returns only the first line of the bucket.
func FirstAggregate(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}
