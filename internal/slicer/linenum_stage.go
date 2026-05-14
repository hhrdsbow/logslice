package slicer

import "context"

// LineNumStageOption configures a NewLineNumStage.
type LineNumStageOption func(*lineNumStageConfig)

type lineNumStageConfig struct {
	start int64
	end   int64
}

// WithLineStart sets the first line number to include (1-based).
func WithLineStart(n int64) LineNumStageOption {
	return func(c *lineNumStageConfig) {
		c.start = n
	}
}

// WithLineEnd sets the last line number to include (0 = no limit).
func WithLineEnd(n int64) LineNumStageOption {
	return func(c *lineNumStageConfig) {
		c.end = n
	}
}

// NewLineNumStage returns a pipeline stage that passes only lines within the
// configured line number range. Lines outside the range are silently dropped.
func NewLineNumStage(in <-chan string, opts ...LineNumStageOption) (<-chan string, error) {
	cfg := &lineNumStageConfig{start: 1}
	for _, o := range opts {
		o(cfg)
	}

	f := NewLineNumFilter(cfg.start, cfg.end)
	out := make(chan string)

	go func() {
		defer close(out)
		for line := range in {
			if f.Accept(line) {
				out <- line
			}
			// once we've passed the end we can stop early
			if cfg.end > 0 && f.Current() > cfg.end {
				// drain remaining input to avoid blocking the producer
				go func() { for range in {} }()
				return
			}
		}
	}()

	return out, nil
}

// NewLineNumStageCtx is a context-aware variant of NewLineNumStage.
func NewLineNumStageCtx(ctx context.Context, in <-chan string, opts ...LineNumStageOption) (<-chan string, error) {
	cfg := &lineNumStageConfig{start: 1}
	for _, o := range opts {
		o(cfg)
	}

	f := NewLineNumFilter(cfg.start, cfg.end)
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
				if cfg.end > 0 && f.Current() > cfg.end {
					return
				}
			}
		}
	}()

	return out, nil
}
