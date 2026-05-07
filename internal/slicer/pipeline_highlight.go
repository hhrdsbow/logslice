package slicer

import "fmt"

// HighlightOption configures a highlight pipeline stage.
type HighlightOption func(*highlightConfig)

type highlightConfig struct {
	pattern string
	prefix  string
	suffix  string
	strip   bool
}

// WithHighlightPattern sets the regex pattern to match for highlighting.
func WithHighlightPattern(pattern string) HighlightOption {
	return func(c *highlightConfig) { c.pattern = pattern }
}

// WithHighlightColor sets the ANSI prefix and suffix used to wrap matches.
func WithHighlightColor(prefix, suffix string) HighlightOption {
	return func(c *highlightConfig) {
		c.prefix = prefix
		c.suffix = suffix
	}
}

// WithANSIStrip configures the stage to strip ANSI codes instead of adding them.
func WithANSIStrip() HighlightOption {
	return func(c *highlightConfig) { c.strip = true }
}

// NewHighlightPipelineStage builds a pipeline stage that highlights or strips
// ANSI codes from log lines. At least WithHighlightPattern must be provided
// unless WithANSIStrip is used alone.
func NewHighlightPipelineStage(in <-chan string, opts ...HighlightOption) (<-chan string, error) {
	cfg := &highlightConfig{
		prefix: "\033[33m", // yellow by default
		suffix: "\033[0m",
	}
	for _, o := range opts {
		o(cfg)
	}

	if cfg.strip {
		out := make(chan string)
		go func() {
			defer close(out)
			for line := range in {
				out <- ANSIStrip(line)
			}
		}()
		return out, nil
	}

	if cfg.pattern == "" {
		return nil, fmt.Errorf("highlight stage: pattern required unless WithANSIStrip is set")
	}

	h, err := NewHighlighter(cfg.pattern, cfg.prefix, cfg.suffix)
	if err != nil {
		return nil, fmt.Errorf("highlight stage: %w", err)
	}

	return NewHighlightStage(in, h), nil
}
