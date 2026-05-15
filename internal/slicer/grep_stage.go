package slicer

import "context"

// GrepStageOption configures the pipeline grep stage.
type GrepStageOption func(*grepStageConfig)

type grepStageConfig struct {
	pattern string
	invert  bool
	group   int
}

// WithGrepPattern sets the regular expression pattern.
func WithGrepPattern(p string) GrepStageOption {
	return func(c *grepStageConfig) { c.pattern = p }
}

// WithGrepStageInvert inverts the match logic.
func WithGrepStageInvert() GrepStageOption {
	return func(c *grepStageConfig) { c.invert = true }
}

// WithGrepCaptureGroup selects a capture group index to extract.
func WithGrepCaptureGroup(n int) GrepStageOption {
	return func(c *grepStageConfig) { c.group = n }
}

// NewGrepStage creates a pipeline stage that filters (and optionally
// transforms) lines using a regular expression.
// It returns an error if no pattern is provided or the pattern is invalid.
func NewGrepStage(ctx context.Context, in <-chan string, opts ...GrepStageOption) (<-chan string, error) {
	cfg := &grepStageConfig{}
	for _, o := range opts {
		o(cfg)
	}

	grepOpts := []GrepOption{WithGrepGroup(cfg.group)}
	if cfg.invert {
		grepOpts = append(grepOpts, WithGrepInvert())
	}

	f, err := NewGrepFilter(cfg.pattern, grepOpts...)
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
				if result, keep := f.Apply(line); keep {
					select {
					case out <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out, nil
}
