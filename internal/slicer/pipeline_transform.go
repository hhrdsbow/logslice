package slicer

// TransformStage wraps a Transformer and implements a pipeline stage that
// applies transformations to each line before passing it downstream.
type TransformStage struct {
	transformer *Transformer
}

// NewTransformStage creates a TransformStage using the provided Transformer.
func NewTransformStage(tr *Transformer) *TransformStage {
	if tr == nil {
		tr = NewTransformer()
	}
	return &TransformStage{transformer: tr}
}

// Run reads lines from in, applies the transformer, and sends results to out.
// It closes out when in is closed or done is signalled.
func (ts *TransformStage) Run(done <-chan struct{}, in <-chan string) <-chan string {
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
				transformed := ts.transformer.Apply(line)
				if transformed == "" {
					continue
				}
				select {
				case out <- transformed:
				case <-done:
					return
				}
			}
		}
	}()
	return out
}
