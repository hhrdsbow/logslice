package slicer

import (
	"math"
	"strings"
)

// EntropyFilter drops or keeps lines based on Shannon entropy of their content.
// High-entropy lines often indicate noise (e.g. base64, binary garbage);
// low-entropy lines may indicate repetitive/boring output.
type EntropyFilter struct {
	minEntropy float64
	maxEntropy float64
	invert      bool
}

// EntropyOption configures an EntropyFilter.
type EntropyOption func(*EntropyFilter)

// WithMinEntropy sets the minimum acceptable entropy (inclusive).
func WithMinEntropy(min float64) EntropyOption {
	return func(e *EntropyFilter) { e.minEntropy = min }
}

// WithMaxEntropy sets the maximum acceptable entropy (inclusive).
func WithMaxEntropy(max float64) EntropyOption {
	return func(e *EntropyFilter) { e.maxEntropy = max }
}

// WithEntropyInvert inverts the filter: keep lines outside the range.
func WithEntropyInvert() EntropyOption {
	return func(e *EntropyFilter) { e.invert = true }
}

// NewEntropyFilter constructs an EntropyFilter with the given options.
// Default range is [0, 8] (pass everything).
func NewEntropyFilter(opts ...EntropyOption) *EntropyFilter {
	f := &EntropyFilter{
		minEntropy: 0.0,
		maxEntropy: 8.0,
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Accept returns true if the line's entropy is within [min, max] (or outside if inverted).
func (f *EntropyFilter) Accept(line string) bool {
	e := shannonEntropy(line)
	inRange := e >= f.minEntropy && e <= f.maxEntropy
	if f.invert {
		return !inRange
	}
	return inRange
}

// Entropy returns the Shannon entropy of the given string.
func (f *EntropyFilter) Entropy(line string) float64 {
	return shannonEntropy(line)
}

// shannonEntropy computes the Shannon entropy (bits) of a string.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, r := range s {
		freq[r]++
	}
	n := float64(strings.Count(s, "") - 1) // rune count
	if n == 0 {
		return 0
	}
	var h float64
	for _, c := range freq {
		p := float64(c) / n
		h -= p * math.Log2(p)
	}
	return h
}

// NewEntropyStage returns a pipeline stage that filters lines by entropy.
func NewEntropyStage(in <-chan string, opts ...EntropyOption) <-chan string {
	f := NewEntropyFilter(opts...)
	out := make(chan string)
	go func() {
		defer close(out)
		for line := range in {
			if f.Accept(line) {
				out <- line
			}
		}
	}()
	return out
}
