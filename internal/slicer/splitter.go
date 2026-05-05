package slicer

import (
	"context"
	"fmt"
	"io"
)

// SplitConfig holds configuration for splitting a log stream into segments.
type SplitConfig struct {
	// MaxLines is the maximum number of lines per segment (0 = unlimited).
	MaxLines int
	// MaxBytes is the maximum number of bytes per segment (0 = unlimited).
	MaxBytes int64
}

// Splitter reads lines from a LineReader and emits Segments based on size limits.
type Splitter struct {
	reader *LineReader
	cfg    SplitConfig
}

// NewSplitter creates a Splitter that reads from r and splits by cfg.
func NewSplitter(r io.Reader, cfg SplitConfig) *Splitter {
	return &Splitter{
		reader: NewLineReader(r),
		cfg:    cfg,
	}
}

// Split reads all lines and emits segments over the returned channel.
// The channel is closed when the reader is exhausted or ctx is cancelled.
func (s *Splitter) Split(ctx context.Context) (<-chan *Segment, <-chan error) {
	out := make(chan *Segment)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		segIdx := 0
		current := NewSegment(fmt.Sprintf("segment-%03d", segIdx))

		for line := range s.reader.Lines(ctx) {
			if s.shouldFlush(current, int64(len(line))) {
				if !current.IsEmpty() {
					select {
					case out <- current:
					case <-ctx.Done():
						return
					}
				}
				segIdx++
				current = NewSegment(fmt.Sprintf("segment-%03d", segIdx))
			}
			current.Add(line)
		}

		if !current.IsEmpty() {
			select {
			case out <- current:
			case <-ctx.Done():
			}
		}
	}()

	return out, errs
}

// shouldFlush returns true if adding a line of size lineBytes would exceed limits.
func (s *Splitter) shouldFlush(seg *Segment, lineBytes int64) bool {
	if s.cfg.MaxLines > 0 && seg.Len() >= s.cfg.MaxLines {
		return true
	}
	if s.cfg.MaxBytes > 0 && seg.Bytes()+lineBytes > s.cfg.MaxBytes {
		return true
	}
	return false
}
