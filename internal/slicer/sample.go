package slicer

import (
	"sync/atomic"
)

// Sampler performs reservoir-style line sampling, keeping every Nth line.
// It is safe for concurrent use.
type Sampler struct {
	n       uint64
	counter atomic.Uint64
}

// NewSampler creates a Sampler that passes through every nth line (1-based).
// n=1 means every line is kept; n=0 is treated as n=1.
func NewSampler(n uint64) *Sampler {
	if n == 0 {
		n = 1
	}
	return &Sampler{n: n}
}

// Keep returns true if the current line should be kept according to the
// sampling rate. The counter increments on every call.
func (s *Sampler) Keep(line string) bool {
	_ = line // sampling is position-based, not content-based
	count := s.counter.Add(1)
	return count%s.n == 0
}

// Reset resets the internal counter back to zero.
func (s *Sampler) Reset() {
	s.counter.Store(0)
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() uint64 {
	return s.n
}

// SampleStage wraps a Sampler as a pipeline-compatible filter function.
// It returns a LineFilter that can be composed into a pipeline.
func SampleStage(n uint64) func(in <-chan string, done <-chan struct{}) <-chan string {
	sampler := NewSampler(n)
	return func(in <-chan string, done <-chan struct{}) <-chan string {
		out := make(chan string)
		go func() {
			defer close(out)
			for {
				select {
				case <-done:
					return
				case line, ok := <-in:
					if !ok {
						return
					}
					if sampler.Keep(line) {
						select {
						case out <- line:
						case <-done:
							return
						}
					}
				}
			}
		}()
		return out
	}
}
