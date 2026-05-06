package slicer

import (
	"context"
	"fmt"
)

// RotateStage is a pipeline stage that fans lines into a RotateWriter.
// It reads from in, writes each line to the RotateWriter, and forwards
// every line to the returned output channel so downstream stages still
// receive all data.
type RotateStage struct {
	writer *RotateWriter
}

// NewRotateStage creates a RotateStage backed by the given RotateWriter.
func NewRotateStage(rw *RotateWriter) *RotateStage {
	return &RotateStage{writer: rw}
}

// Run starts the stage. It returns a channel of lines and a channel that
// carries the first non-nil error (or is closed on success). The RotateWriter
// is NOT closed when the stage finishes; callers own its lifecycle.
func (rs *RotateStage) Run(ctx context.Context, in <-chan string) (<-chan string, <-chan error) {
	out := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				if err := rs.writer.WriteLine(line); err != nil {
					errCh <- fmt.Errorf("rotate stage: %w", err)
					return
				}
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, errCh
}
