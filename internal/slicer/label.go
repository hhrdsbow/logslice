package slicer

import (
	"fmt"
	"regexp"
	"strings"
)

// LabelRule maps a compiled regex to a label string.
type LabelRule struct {
	pattern *regexp.Regexp
	label   string
}

// Labeler assigns labels to lines based on matching rules.
type Labeler struct {
	rules  []LabelRule
	format string // e.g. "[%s] %s"
	defaultLabel string
}

// NewLabeler creates a Labeler. format must contain two %s verbs (label, line).
// An empty defaultLabel means unmatched lines pass through unchanged.
func NewLabeler(format, defaultLabel string) *Labeler {
	if format == "" {
		format = "[%s] %s"
	}
	return &Labeler{format: format, defaultLabel: defaultLabel}
}

// AddRule registers a pattern→label rule. Returns error on bad regex.
func (l *Labeler) AddRule(pattern, label string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("label rule: invalid pattern %q: %w", pattern, err)
	}
	l.rules = append(l.rules, LabelRule{pattern: re, label: label})
	return nil
}

// Apply returns the line with a label prefix if any rule matches.
// Rules are evaluated in order; the first match wins.
func (l *Labeler) Apply(line string) string {
	for _, r := range l.rules {
		if r.pattern.MatchString(line) {
			return fmt.Sprintf(l.format, r.label, line)
		}
	}
	if l.defaultLabel != "" {
		return fmt.Sprintf(l.format, l.defaultLabel, line)
	}
	return line
}

// LabelTransform returns a TransformFunc backed by the Labeler.
func LabelTransform(l *Labeler) TransformFunc {
	return func(line string) string {
		return l.Apply(line)
	}
}

// StripLabel removes a label prefix added by the default format "[<label>] ".
func StripLabel(line string) string {
	if len(line) > 0 && line[0] == '[' {
		if idx := strings.Index(line, "] "); idx != -1 {
			return line[idx+2:]
		}
	}
	return line
}
