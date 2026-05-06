package slicer

import (
	"context"
	"time"
)

// WindowPipelineOption configures NewWindowPipelineStage.
type WindowPipelineOption func(*windowPipelineConfig)

type windowPipelineConfig struct {
	parser func(string) (time.Time, bool)
}

// WithWindowParser sets a custom timestamp parser for the window pipeline stage.
func WithWindowParser(p func(string) (time.Time, bool)) WindowPipelineOption {
	return func(c *windowPipelineConfig) { c.parser = p }
}

// NewWindowPipelineStage wires a WindowStage into the pipeline and returns
// the output channel. duration controls how wide the sliding window is.
// If no parser option is provided, AutoParser is used.
func NewWindowPipelineStage(
	ctx context.Context,
	in <-chan string,
	duration time.Duration,
	opts ...WindowPipelineOption,
) <-chan string {
	cfg := &windowPipelineConfig{
		parser: func(line string) (time.Time, bool) {
			t, err := AutoParser(line)
			if err != nil {
				return time.Time{}, false
			}
			return t, true
		},
	}
	for _, o := range opts {
		o(cfg)
	}
	stage := NewWindowStage(duration, cfg.parser)
	return stage.Run(ctx, in)
}
