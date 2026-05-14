package slicer

import "context"

// severityStageOptions holds configuration for NewSeverityStage.
type severityStageOptions struct {
	min SeverityLevel
}

// SeverityStageOption is a functional option for NewSeverityStage.
type SeverityStageOption func(*severityStageOptions)

// WithMinSeverity sets the minimum severity level a line must meet to pass.
func WithMinSeverity(level SeverityLevel) SeverityStageOption {
	return func(o *severityStageOptions) {
		o.min = level
	}
}

// NewSeverityStage returns a pipeline stage that filters lines by severity.
// Lines whose detected severity is below the minimum are dropped.
// If no options are provided the stage defaults to SeverityDebug (passes all
// recognised severity lines and drops UNKNOWN lines).
func NewSeverityStage(ctx context.Context, in <-chan string, opts ...SeverityStageOption) <-chan string {
	cfg := &severityStageOptions{min: SeverityDebug}
	for _, o := range opts {
		o(cfg)
	}

	f := NewSeverityFilter(cfg.min)
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
				if f.Accept(line) {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out
}
