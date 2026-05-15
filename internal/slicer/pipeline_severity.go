package slicer

import "context"

// SeverityStageOption configures a severity filter pipeline stage.
type SeverityStageOption func(*severityStageConfig)

type severityStageConfig struct {
	minLevel SeverityLevel
}

// WithStageSeverity sets the minimum severity level lines must meet to pass through.
func WithStageSeverity(level SeverityLevel) SeverityStageOption {
	return func(cfg *severityStageConfig) {
		cfg.minLevel = level
	}
}

// NewSeverityPipelineStage returns a pipeline stage that filters lines below
// the configured minimum severity level. Lines whose severity cannot be
// detected are forwarded unchanged when the minimum is SeverityUnknown,
// otherwise they are dropped.
//
// Usage:
//
//	stage := NewSeverityPipelineStage(in, WithStageSeverity(SeverityWarn))
func NewSeverityPipelineStage(in <-chan string, opts ...SeverityStageOption) <-chan string {
	cfg := &severityStageConfig{
		minLevel: SeverityUnknown,
	}
	for _, o := range opts {
		o(cfg)
	}
	return NewSeverityPipelineStageCtx(context.Background(), in, cfg.minLevel)
}

// NewSeverityPipelineStageCtx is like NewSeverityPipelineStage but accepts a
// context for cancellation.
func NewSeverityPipelineStageCtx(ctx context.Context, in <-chan string, minLevel SeverityLevel) <-chan string {
	out := make(chan string)
	f := NewSeverityFilter(minLevel)
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
