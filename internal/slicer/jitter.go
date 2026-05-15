package slicer

import (
	"context"
	"math/rand"
	"time"
)

// JitterConfig holds configuration for the jitter stage.
type JitterConfig struct {
	MaxDelay time.Duration
	rng      *rand.Rand
}

// JitterOption configures a JitterConfig.
type JitterOption func(*JitterConfig)

// WithMaxJitter sets the maximum random delay applied per line.
func WithMaxJitter(d time.Duration) JitterOption {
	return func(c *JitterConfig) {
		if d > 0 {
			c.MaxDelay = d
		}
	}
}

// NewJitterStage inserts a random sub-millisecond to MaxDelay pause before
// forwarding each line. Useful for smoothing bursty replay pipelines.
// If MaxDelay is zero or unset, lines pass through without delay.
func NewJitterStage(ctx context.Context, in <-chan string, opts ...JitterOption) <-chan string {
	cfg := &JitterConfig{
		MaxDelay: 0,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, o := range opts {
		o(cfg)
	}

	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				if cfg.MaxDelay > 0 {
					delay := time.Duration(cfg.rng.Int63n(int64(cfg.MaxDelay)))
					select {
					case <-time.After(delay):
					case <-ctx.Done():
						return
					}
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
