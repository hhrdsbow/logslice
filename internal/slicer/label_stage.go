package slicer

import "context"

// LabelStageOption configures a label pipeline stage.
type LabelStageOption func(*labelStageConfig)

type labelStageConfig struct {
	rules         []labelRuleSpec
	format        string
	defaultLabel  string
}

type labelRuleSpec struct {
	pattern string
	label   string
}

// WithLabelRule adds a pattern→label rule to the stage.
func WithLabelRule(pattern, label string) LabelStageOption {
	return func(c *labelStageConfig) {
		c.rules = append(c.rules, labelRuleSpec{pattern: pattern, label: label})
	}
}

// WithLabelFormat sets a custom printf format string (must have two %s: label, line).
func WithLabelFormat(format string) LabelStageOption {
	return func(c *labelStageConfig) { c.format = format }
}

// WithDefaultLabel sets the label applied to lines that match no rule.
func WithDefaultLabel(label string) LabelStageOption {
	return func(c *labelStageConfig) { c.defaultLabel = label }
}

// NewLabelStage returns a pipeline stage that labels each line.
// It returns (nil, err) if any rule pattern is invalid.
func NewLabelStage(in <-chan string, opts ...LabelStageOption) (<-chan string, error) {
	cfg := &labelStageConfig{}
	for _, o := range opts {
		o(cfg)
	}

	labeler := NewLabeler(cfg.format, cfg.defaultLabel)
	for _, r := range cfg.rules {
		if err := labeler.AddRule(r.pattern, r.label); err != nil {
			return nil, err
		}
	}

	out := make(chan string)
	go func() {
		defer close(out)
		ctx := context.Background()
		_ = ctx
		for line := range in {
			out <- labeler.Apply(line)
		}
	}()
	return out, nil
}
