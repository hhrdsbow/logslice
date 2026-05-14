package slicer

import (
	"context"
)

// ContextFilter emits matching lines plus N lines of surrounding context
// (before and after each match), similar to grep -C.
type ContextFilter struct {
	matcher  func(string) bool
	before   int
	after    int
	buf      []string // ring buffer for pre-match lines
	afterLeft int
}

// NewContextFilter creates a ContextFilter. before and after specify how
// many surrounding lines to include around each matched line. Values less
// than zero are clamped to zero.
func NewContextFilter(matcher func(string) bool, before, after int) *ContextFilter {
	if before < 0 {
		before = 0
	}
	if after < 0 {
		after = 0
	}
	return &ContextFilter{
		matcher: matcher,
		before:  before,
		after:   after,
		buf:     make([]string, 0, before),
	}
}

// Run reads lines from in, applies context filtering, and sends results to
// out. It closes out when in is drained or ctx is cancelled.
func (cf *ContextFilter) Run(ctx context.Context, in <-chan string, out chan<- string) {
	defer close(out)

	send := func(line string) bool {
		select {
		case out <- line:
			return true
		case <-ctx.Done():
			return false
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-in:
			if !ok {
				return
			}
			if cf.matcher(line) {
				// flush pre-match buffer
				for _, b := range cf.buf {
					if !send(b) {
						return
					}
				}
				cf.buf = cf.buf[:0]
				if !send(line) {
					return
				}
				cf.afterLeft = cf.after
			} else if cf.afterLeft > 0 {
				if !send(line) {
					return
				}
				cf.afterLeft--
			} else {
				if cf.before > 0 {
					if len(cf.buf) == cf.before {
						cf.buf = cf.buf[1:]
					}
					cf.buf = append(cf.buf, line)
				}
			}
		}
	}
}
