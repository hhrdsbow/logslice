package slicer

import (
	"context"
)

// headStageOptions holds configuration for the head pipeline stage.
type headStageOptions struct {
	max int
}

// HeadStageOption configures a head pipeline stage.
type HeadStageOption func(*headStageOptions)

// WithHeadMax sets the maximum number of lines the head stage will forward.
func WithHeadMax(n int) HeadStageOption {
	return func(o *headStageOptions) {
		o.max = n
	}
}

// NewHeadStage returns a pipeline stage function that limits output to the
// first N lines. It wraps HeadReader so it fits the standard stage signature
// func(context.Context, <-chan string) <-chan string.
func NewHeadStage(opts ...HeadStageOption) func(context.Context, <-chan string) <-chan string {
	cfg := &headStageOptions{max: 10}
	for _, o := range opts {
		o(cfg)
	}
	r := NewHeadReader(cfg.max)
	return func(ctx context.Context, in <-chan string) <-chan string {
		return r.Read(ctx, in)
	}
}
