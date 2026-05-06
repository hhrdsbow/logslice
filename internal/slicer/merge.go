package slicer

import (
	"context"
	"sync"
)

// MergeReader merges multiple line channels into a single output channel,
// preserving order of arrival (fan-in). All input channels are drained
// concurrently; the output channel is closed once every input is exhausted.
type MergeReader struct {
	inputs []<-chan string
}

// NewMergeReader creates a MergeReader that fans in the provided channels.
func NewMergeReader(inputs ...<-chan string) *MergeReader {
	return &MergeReader{inputs: inputs}
}

// Read starts draining all input channels and emits lines to the returned
// channel. The returned channel is closed when all inputs are exhausted or
// ctx is cancelled.
func (m *MergeReader) Read(ctx context.Context) <-chan string {
	out := make(chan string, len(m.inputs)*8)

	var wg sync.WaitGroup
	for _, ch := range m.inputs {
		wg.Add(1)
		go func(src <-chan string) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case line, ok := <-src:
					if !ok {
						return
					}
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
