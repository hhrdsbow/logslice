package slicer

import (
	"fmt"
	"io"
	"sync/atomic"
)

// Progress provides a lightweight, goroutine-safe progress reporter that
// periodically emits line-count updates to an io.Writer.
type Progress struct {
	w        io.Writer
	linesRead atomic.Int64
	bytes     atomic.Int64
	verbose  bool
}

// NewProgress creates a Progress that writes to w.
// When verbose is true, byte counts are also reported.
func NewProgress(w io.Writer, verbose bool) *Progress {
	return &Progress{w: w, verbose: verbose}
}

// Tick records one processed line and n matched bytes.
func (p *Progress) Tick(matched bool, n int) {
	p.linesRead.Add(1)
	if matched {
		p.bytes.Add(int64(n))
	}
}

// LinesRead returns the current line count.
func (p *Progress) LinesRead() int64 {
	return p.linesRead.Load()
}

// BytesMatched returns the total matched bytes recorded so far.
func (p *Progress) BytesMatched() int64 {
	return p.bytes.Load()
}

// Report writes the current progress snapshot to the underlying writer.
func (p *Progress) Report() {
	lines := p.linesRead.Load()
	if p.verbose {
		bytes := p.bytes.Load()
		fmt.Fprintf(p.w, "progress: %d lines read, %d bytes matched\n", lines, bytes)
	} else {
		fmt.Fprintf(p.w, "progress: %d lines read\n", lines)
	}
}

// Summary writes a final one-line summary derived from a Stats value.
func (p *Progress) Summary(s *Stats) {
	fmt.Fprintf(p.w, "done: %d/%d lines matched across %d segment(s) in %s\n",
		s.LinesMatched, s.LinesRead, s.SegmentsTotal, s.Duration().Round(1e6))
}
