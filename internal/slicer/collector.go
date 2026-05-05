package slicer

import (
	"context"
	"fmt"
)

// CollectorConfig controls how lines are gathered into segments.
type CollectorConfig struct {
	MaxLinesPerSegment int    // 0 means unlimited
	SegmentPrefix      string // prefix for auto-named segments
}

// Collector reads matched lines from a channel and groups them into Segments.
type Collector struct {
	cfg     CollectorConfig
	counter int
}

// NewCollector creates a Collector with the given config.
func NewCollector(cfg CollectorConfig) *Collector {
	if cfg.SegmentPrefix == "" {
		cfg.SegmentPrefix = "segment"
	}
	return &Collector{cfg: cfg}
}

// Collect reads lines from ch and emits complete Segments on out.
// It respects ctx cancellation and closes out when done.
func (c *Collector) Collect(ctx context.Context, ch <-chan string, out chan<- *Segment) error {
	defer close(out)

	current := c.newSegment()

	for {
		select {
		case <-ctx.Done():
			if !current.IsEmpty() {
				out <- current
			}
			return ctx.Err()
		case line, ok := <-ch:
			if !ok {
				if !current.IsEmpty() {
					out <- current
				}
				return nil
			}
			current.Add(line)
			if c.cfg.MaxLinesPerSegment > 0 && current.Len() >= c.cfg.MaxLinesPerSegment {
				out <- current
				current = c.newSegment()
			}
		}
	}
}

func (c *Collector) newSegment() *Segment {
	c.counter++
	name := fmt.Sprintf("%s-%04d", c.cfg.SegmentPrefix, c.counter)
	return NewSegment(name)
}
