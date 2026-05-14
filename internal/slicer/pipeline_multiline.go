package slicer

// multilineStageConfig holds options for NewMultilineStage.
type multilineStageConfig struct {
	opts []MultilineOption
}

// MultilineStageOption configures the pipeline multiline stage.
type MultilineStageOption func(*multilineStageConfig)

// WithMultilineStart sets the start-of-block regex pattern.
func WithMultilineStart(pat string) MultilineStageOption {
	return func(c *multilineStageConfig) {
		c.opts = append(c.opts, WithStartPattern(pat))
	}
}

// WithMultilineContinue sets the continuation-line regex pattern.
func WithMultilineContinue(pat string) MultilineStageOption {
	return func(c *multilineStageConfig) {
		c.opts = append(c.opts, WithContinuePattern(pat))
	}
}

// WithMultilineJoinSep overrides the join separator (default space).
func WithMultilineJoinSep(sep string) MultilineStageOption {
	return func(c *multilineStageConfig) {
		c.opts = append(c.opts, WithJoinSep(sep))
	}
}

// NewMultilineStage returns a pipeline stage that folds multiline blocks into
// single logical lines. It returns an error if no start or continue pattern is
// provided, or if either pattern is an invalid regex.
func NewMultilineStage(in <-chan string, done <-chan struct{}, options ...MultilineStageOption) (<-chan string, error) {
	cfg := &multilineStageConfig{}
	for _, o := range options {
		o(cfg)
	}
	f, err := NewMultilineFolder(cfg.opts...)
	if err != nil {
		return nil, err
	}
	return f.Fold(in, done), nil
}
