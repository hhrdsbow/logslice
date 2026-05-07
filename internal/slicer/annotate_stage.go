package slicer

import "context"

// AnnotateStage is a pipeline stage that annotates every line passing through
// it using an Annotator.
type AnnotateStage struct {
	annotator *Annotator
}

// NewAnnotateStage creates a pipeline stage wrapping the given Annotator.
func NewAnnotateStage(a *Annotator) *AnnotateStage {
	return &AnnotateStage{annotator: a}
}

// Run reads lines from in, annotates each one, and forwards the result to the
// returned channel. The output channel is closed when in is closed or ctx is
// cancelled.
func (s *AnnotateStage) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		var lineNum int64
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				lineNum++
				annotated := s.annotator.Apply(line, lineNum)
				select {
				case out <- annotated:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
