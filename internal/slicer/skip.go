package slicer

import "context"

// SkipReader wraps an input channel and skips the first N lines,
// forwarding all subsequent lines to the output channel.
type SkipReader struct {
	input <-chan string
	skip  int
}

// NewSkipReader creates a SkipReader that skips the first n lines.
// If n <= 0 it is clamped to 0 (no lines skipped).
func NewSkipReader(input <-chan string, n int) *SkipReader {
	if n < 0 {
		n = 0
	}
	return &SkipReader{input: input, skip: n}
}

// Lines returns a channel that emits every line after the first n have
// been discarded. The channel is closed once the input is exhausted or
// ctx is cancelled.
func (s *SkipReader) Lines(ctx context.Context) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		skipped := 0
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-s.input:
				if !ok {
					return
				}
				if skipped < s.skip {
					skipped++
					continue
				}
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
