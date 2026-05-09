package slicer

import "context"

// skipStageOptions holds configuration for the skip pipeline stage.
type skipStageOptions struct {
	n int
}

// SkipOption is a functional option for NewSkipStage.
type SkipOption func(*skipStageOptions)

// WithSkipN sets the number of lines to skip at the start of the stream.
func WithSkipN(n int) SkipOption {
	return func(o *skipStageOptions) {
		o.n = n
	}
}

// NewSkipStage builds a pipeline stage that discards the first n lines of
// the input channel and forwards the rest. It mirrors the pattern used by
// NewHeadStage and other pipeline helpers in this package.
func NewSkipStage(ctx context.Context, input <-chan string, opts ...SkipOption) <-chan string {
	cfg := &skipStageOptions{n: 0}
	for _, o := range opts {
		o(cfg)
	}
	return NewSkipReader(input, cfg.n).Lines(ctx)
}
