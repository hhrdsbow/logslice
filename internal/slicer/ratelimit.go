package slicer

import (
	"context"
	"time"
)

// RateLimiter controls the throughput of lines emitted from a channel,
// enforcing a maximum number of lines per second. A zero or negative
// rate means no limiting is applied.
type RateLimiter struct {
	rate     int           // max lines per second
	interval time.Duration // derived from rate
}

// NewRateLimiter creates a RateLimiter that allows at most linesPerSec
// lines per second. If linesPerSec <= 0 the limiter is a no-op.
func NewRateLimiter(linesPerSec int) *RateLimiter {
	rl := &RateLimiter{rate: linesPerSec}
	if linesPerSec > 0 {
		rl.interval = time.Second / time.Duration(linesPerSec)
	}
	return rl
}

// Apply reads lines from in and forwards them to the returned channel,
// inserting a delay between lines when a rate limit is configured.
// The output channel is closed when in is closed or ctx is cancelled.
func (rl *RateLimiter) Apply(ctx context.Context, in <-chan string) <-chan string {
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
				if rl.interval > 0 {
					select {
					case <-time.After(rl.interval):
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

// Rate returns the configured lines-per-second limit (0 means unlimited).
func (rl *RateLimiter) Rate() int {
	return rl.rate
}
