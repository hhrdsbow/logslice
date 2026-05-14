package slicer

import "context"

// JSONPathStageOption configures a JSONPath pipeline stage.
type JSONPathStageOption func(*jsonPathStageConfig)

type jsonPathStageConfig struct {
	dotPath  string
	fallback string
	passthrough bool
}

// WithJSONPath sets the dot-separated JSON key path to extract.
func WithJSONPath(dotPath string) JSONPathStageOption {
	return func(c *jsonPathStageConfig) { c.dotPath = dotPath }
}

// WithJSONStageFallback sets the fallback value for missing keys.
func WithJSONStageFallback(s string) JSONPathStageOption {
	return func(c *jsonPathStageConfig) { c.fallback = s }
}

// WithJSONPassthrough keeps lines that are not valid JSON unchanged instead of
// emitting the fallback.
func WithJSONPassthrough() JSONPathStageOption {
	return func(c *jsonPathStageConfig) { c.passthrough = true }
}

// NewJSONPathStage creates a pipeline stage that replaces each line with the
// value found at the configured JSON path. Lines that cannot be parsed as JSON
// are replaced by the fallback (or kept unchanged when WithJSONPassthrough is set).
// Returns an error if no path is configured or the path is invalid.
func NewJSONPathStage(in <-chan string, opts ...JSONPathStageOption) (<-chan string, error) {
	cfg := &jsonPathStageConfig{}
	for _, o := range opts {
		o(cfg)
	}

	extractorOpts := []JSONPathOption{}
	if cfg.fallback != "" {
		extractorOpts = append(extractorOpts, WithJSONFallback(cfg.fallback))
	}

	ext, err := NewJSONPathExtractor(cfg.dotPath, extractorOpts...)
	if err != nil {
		return nil, err
	}

	out := make(chan string)
	go func() {
		defer close(out)
		for line := range in {
			result := ext.Extract(line)
			if cfg.passthrough && result == cfg.fallback && result == "" {
				// heuristic: if fallback is empty and result is empty, pass original
				out <- line
				continue
			}
			out <- result
		}
	}()
	return out, nil
}

// NewJSONPathStageCtx is like NewJSONPathStage but respects context cancellation.
func NewJSONPathStageCtx(ctx context.Context, in <-chan string, opts ...JSONPathStageOption) (<-chan string, error) {
	cfg := &jsonPathStageConfig{}
	for _, o := range opts {
		o(cfg)
	}

	extractorOpts := []JSONPathOption{}
	if cfg.fallback != "" {
		extractorOpts = append(extractorOpts, WithJSONFallback(cfg.fallback))
	}

	ext, err := NewJSONPathExtractor(cfg.dotPath, extractorOpts...)
	if err != nil {
		return nil, err
	}

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
				out <- ext.Extract(line)
			}
		}
	}()
	return out, nil
}
