package slicer

import (
	"context"
	"fmt"
)

// HeadReader emits only the first N lines from an input channel.
type HeadReader struct {
	max int
}

// NewHeadReader creates a HeadReader that forwards at most max lines.
// If max <= 0 it is clamped to 1.
func NewHeadReader(max int) *HeadReader {
	if max <= 0 {
		max = 1
	}
	return &HeadReader{max: max}
}

// Read reads from in and writes at most h.max lines to the returned channel.
// The output channel is closed when the limit is reached or in is exhausted.
func (h *HeadReader) Read(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
				count++
				if count >= h.max {
					return
				}
			}
		}
	}()
	return out
}

// Summary returns a human-readable description of the head limit.
func (h *HeadReader) Summary() string {
	return fmt.Sprintf("head: first %d line(s)", h.max)
}
