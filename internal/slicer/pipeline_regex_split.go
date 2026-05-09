package slicer

import (
	"context"
	"fmt"
)

// RegexSplitStageOption configures NewRegexSplitStage.
type RegexSplitStageOption func(*regexSplitStageConfig)

type regexSplitStageConfig struct {
	pattern  string
	invert   bool
	maxSegs  int
	sep      string
}

// WithSplitPattern sets the boundary regex pattern (required).
func WithSplitPattern(pattern string) RegexSplitStageOption {
	return func(c *regexSplitStageConfig) { c.pattern = pattern }
}

// WithSplitInvertStage inverts the boundary condition.
func WithSplitInvertStage(invert bool) RegexSplitStageOption {
	return func(c *regexSplitStageConfig) { c.invert = invert }
}

// WithSplitMaxSegments caps the number of emitted segments.
func WithSplitMaxSegments(n int) RegexSplitStageOption {
	return func(c *regexSplitStageConfig) { c.maxSegs = n }
}

// WithSegmentSeparator sets an optional separator line emitted between
// segments in the output channel (empty string = no separator).
func WithSegmentSeparator(sep string) RegexSplitStageOption {
	return func(c *regexSplitStageConfig) { c.sep = sep }
}

// NewRegexSplitStage creates a pipeline stage that splits the input stream
// into segments using a RegexSplitter and flattens them back into a line
// channel, optionally inserting a separator between segments.
//
// Returns an error if the pattern is missing or invalid.
func NewRegexSplitStage(ctx context.Context, in <-chan string, opts ...RegexSplitStageOption) (<-chan string, error) {
	cfg := &regexSplitStageConfig{}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.pattern == "" {
		return nil, fmt.Errorf("regex_split_stage: pattern is required")
	}
	rs, err := NewRegexSplitter(cfg.pattern,
		WithSplitInvert(cfg.invert),
		WithMaxSegments(cfg.maxSegs),
	)
	if err != nil {
		return nil, err
	}
	out := make(chan string)
	go func() {
		defer close(out)
		segCh := rs.Split(ctx, in)
		first := true
		for {
			select {
			case <-ctx.Done():
				return
			case seg, ok := <-segCh:
				if !ok {
					return
				}
				if !first && cfg.sep != "" {
					select {
					case out <- cfg.sep:
					case <-ctx.Done():
						return
					}
				}
				first = false
				for _, line := range seg {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out, nil
}
