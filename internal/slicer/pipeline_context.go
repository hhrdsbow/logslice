package slicer

import "context"

// contextStageOptions holds configuration for the context-lines pipeline stage.
type contextStageOptions struct {
	before  int
	after   int
	matcher func(string) bool
}

// ContextStageOption configures a context-lines stage.
type ContextStageOption func(*contextStageOptions)

// WithContextBefore sets the number of lines to emit before each match.
func WithContextBefore(n int) ContextStageOption {
	return func(o *contextStageOptions) { o.before = n }
}

// WithContextAfter sets the number of lines to emit after each match.
func WithContextAfter(n int) ContextStageOption {
	return func(o *contextStageOptions) { o.after = n }
}

// WithContextMatcher sets the predicate used to identify matching lines.
// If not provided, no lines will match.
func WithContextMatcher(fn func(string) bool) ContextStageOption {
	return func(o *contextStageOptions) { o.matcher = fn }
}

// NewContextStage constructs a pipeline stage that wraps ContextFilter.
// It reads from in and returns a channel of context-filtered lines.
func NewContextStage(ctx context.Context, in <-chan string, opts ...ContextStageOption) <-chan string {
	o := &contextStageOptions{
		matcher: func(string) bool { return false },
	}
	for _, opt := range opts {
		opt(o)
	}

	out := make(chan string, 64)
	cf := NewContextFilter(o.matcher, o.before, o.after)
	go cf.Run(ctx, in, out)
	return out
}
