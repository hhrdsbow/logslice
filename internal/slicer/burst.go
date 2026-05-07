package slicer

import (
	"context"
	"time"
)

// BurstDetector detects bursts of log activity — periods where the line rate
// exceeds a given threshold within a sliding window.
type BurstDetector struct {
	windowSize time.Duration
	threshold  int
	timestamps []time.Time
}

// NewBurstDetector creates a BurstDetector that fires when more than threshold
// lines arrive within windowSize.
func NewBurstDetector(windowSize time.Duration, threshold int) *BurstDetector {
	if threshold < 1 {
		threshold = 1
	}
	return &BurstDetector{
		windowSize: windowSize,
		threshold:  threshold,
		timestamps: make([]time.Time, 0, threshold*2),
	}
}

// Record records a line arrival at now and returns true if a burst is detected.
func (b *BurstDetector) Record(now time.Time) bool {
	cutoff := now.Add(-b.windowSize)
	// evict old entries
	i := 0
	for i < len(b.timestamps) && b.timestamps[i].Before(cutoff) {
		i++
	}
	b.timestamps = append(b.timestamps[i:], now)
	return len(b.timestamps) > b.threshold
}

// Reset clears all recorded timestamps.
func (b *BurstDetector) Reset() {
	b.timestamps = b.timestamps[:0]
}

// BurstStage is a pipeline stage that annotates or filters lines based on
// burst detection. Lines during a burst are tagged with a prefix.
type BurstStage struct {
	detector *BurstDetector
	prefix   string
}

// NewBurstStage creates a pipeline stage using the provided BurstDetector.
// During a burst, each line is prefixed with prefix.
func NewBurstStage(detector *BurstDetector, prefix string) *BurstStage {
	if prefix == "" {
		prefix = "[BURST] "
	}
	return &BurstStage{detector: detector, prefix: prefix}
}

// Run reads from in, tags burst lines, and writes to the returned channel.
func (s *BurstStage) Run(ctx context.Context, in <-chan string) <-chan string {
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
				if s.detector.Record(time.Now()) {
					line = s.prefix + line
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
