package slicer

import (
	"context"
	"fmt"
	"time"
)

// ThrottleStageOption configures a throttle pipeline stage.
type ThrottleStageOption func(*ThrottleConfig)

// WithThrottleMaxLines sets the maximum lines per interval.
func WithThrottleMaxLines(n int) ThrottleStageOption {
	return func(c *ThrottleConfig) { c.MaxLines = n }
}

// WithThrottleInterval sets the sliding window interval.
func WithThrottleInterval(d time.Duration) ThrottleStageOption {
	return func(c *ThrottleConfig) { c.Interval = d }
}

// NewThrottleStage constructs a pipeline stage that throttles line throughput.
// It reads from in, drops lines exceeding the configured rate, and forwards
// the rest to the returned channel.
func NewThrottleStage(ctx context.Context, in <-chan string, opts ...ThrottleStageOption) (<-chan string, error) {
	cfg := ThrottleConfig{
		MaxLines: 100,
		Interval: time.Second,
	}
	for _, o := range opts {
		o(&cfg)
	}
	out, err := ThrottleStage(ctx, in, cfg)
	if err != nil {
		return nil, fmt.Errorf("throttle stage: %w", err)
	}
	return out, nil
}
