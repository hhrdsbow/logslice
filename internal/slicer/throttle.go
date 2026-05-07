package slicer

import (
	"context"
	"time"
)

// ThrottleConfig controls how the throttle stage behaves.
type ThrottleConfig struct {
	// MaxLines is the maximum number of lines allowed per Interval.
	MaxLines int
	// Interval is the sliding window over which MaxLines is enforced.
	Interval time.Duration
}

// Throttle drops lines that exceed the configured rate within a sliding
// time window, forwarding all others to out.
type Throttle struct {
	cfg       ThrottleConfig
	timestamps []time.Time
}

// NewThrottle creates a Throttle. MaxLines must be >= 1; Interval must be > 0.
func NewThrottle(cfg ThrottleConfig) (*Throttle, error) {
	if cfg.MaxLines < 1 {
		cfg.MaxLines = 1
	}
	if cfg.Interval <= 0 {
		cfg.Interval = time.Second
	}
	return &Throttle{cfg: cfg}, nil
}

// Allow returns true if the line should be forwarded.
func (t *Throttle) Allow() bool {
	now := time.Now()
	cutoff := now.Add(-t.cfg.Interval)

	// evict old timestamps
	valid := t.timestamps[:0]
	for _, ts := range t.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	t.timestamps = valid

	if len(t.timestamps) >= t.cfg.MaxLines {
		return false
	}
	t.timestamps = append(t.timestamps, now)
	return true
}

// Reset clears the internal window state.
func (t *Throttle) Reset() {
	t.timestamps = t.timestamps[:0]
}

// ThrottleStage wires a Throttle into a pipeline channel.
func ThrottleStage(ctx context.Context, in <-chan string, cfg ThrottleConfig) (<-chan string, error) {
	th, err := NewThrottle(cfg)
	if err != nil {
		return nil, err
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
				if th.Allow() {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out, nil
}
