package slicer

import (
	"context"
	"time"
)

// jitterStageConfig holds pipeline-level jitter options.
type jitterStageConfig struct {
	maxDelay time.Duration
}

// PipelineJitterOption configures the pipeline jitter stage.
type PipelineJitterOption func(*jitterStageConfig)

// WithJitterMaxDelay sets the upper bound of random delay per line.
func WithJitterMaxDelay(d time.Duration) PipelineJitterOption {
	return func(c *jitterStageConfig) {
		if d > 0 {
			c.maxDelay = d
		}
	}
}

// NewJitterPipelineStage wires a jitter stage into a pipeline.
// It reads from in, applies a random delay bounded by maxDelay, and
// writes to the returned channel. Cancellation via ctx stops the stage.
//
// Example:
//
//	out := NewJitterPipelineStage(ctx, in, WithJitterMaxDelay(5*time.Millisecond))
func NewJitterPipelineStage(ctx context.Context, in <-chan string, opts ...PipelineJitterOption) <-chan string {
	cfg := &jitterStageConfig{}
	for _, o := range opts {
		o(cfg)
	}

	var jOpts []JitterOption
	if cfg.maxDelay > 0 {
		jOpts = append(jOpts, WithMaxJitter(cfg.maxDelay))
	}
	return NewJitterStage(ctx, in, jOpts...)
}
