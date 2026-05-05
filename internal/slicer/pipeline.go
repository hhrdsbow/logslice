package slicer

import (
	"io"
)

// Pipeline ties together a LineReader and a Slicer, writing matching lines
// to an io.Writer. It returns the number of lines written and any error.
type Pipeline struct {
	Reader  *LineReader
	Slicer  Slicer
	Writer  io.Writer
}

// Run executes the pipeline: reads lines, tests each against the Slicer,
// and writes matches to Writer. Cancellation is handled via done.
func (p *Pipeline) Run(done <-chan struct{}) (int64, error) {
	lines, errCh := p.Reader.Lines(done)

	var written int64
	for line := range lines {
		if p.Slicer.Match(line) {
			if _, err := io.WriteString(p.Writer, line+"\n"); err != nil {
				return written, err
			}
			written++
		}
	}

	if err := <-errCh; err != nil {
		return written, err
	}
	return written, nil
}

// NewPipeline is a convenience constructor.
func NewPipeline(r io.Reader, s Slicer, w io.Writer) *Pipeline {
	return &Pipeline{
		Reader: NewLineReader(r, 0),
		Slicer: s,
		Writer: w,
	}
}
