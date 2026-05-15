package slicer

import (
	"fmt"
	"regexp"
)

// GrepFilter retains only lines matching (or not matching) a regular expression,
// optionally capturing a sub-expression for output.
type GrepFilter struct {
	re      *regexp.Regexp
	invert  bool
	group   int // 0 = whole match, >0 = capture group index
}

// GrepOption configures a GrepFilter.
type GrepOption func(*GrepFilter)

// WithGrepInvert causes the filter to pass lines that do NOT match.
func WithGrepInvert() GrepOption {
	return func(g *GrepFilter) { g.invert = true }
}

// WithGrepGroup selects a capture group index whose text replaces the line.
// Index 0 means the whole match.
func WithGrepGroup(n int) GrepOption {
	return func(g *GrepFilter) {
		if n >= 0 {
			g.group = n
		}
	}
}

// NewGrepFilter creates a GrepFilter for the given regular expression.
func NewGrepFilter(pattern string, opts ...GrepOption) (*GrepFilter, error) {
	if pattern == "" {
		return nil, fmt.Errorf("grep: pattern must not be empty")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("grep: invalid pattern: %w", err)
	}
	g := &GrepFilter{re: re}
	for _, o := range opts {
		o(g)
	}
	return g, nil
}

// Apply returns the (possibly transformed) line and whether it should be kept.
// If the filter has a capture group configured the returned string contains
// only that group's text when there is a match.
func (g *GrepFilter) Apply(line string) (string, bool) {
	matches := g.re.FindStringSubmatch(line)
	matched := len(matches) > 0
	if g.invert {
		return line, !matched
	}
	if !matched {
		return line, false
	}
	if g.group > 0 && g.group < len(matches) {
		return matches[g.group], true
	}
	return line, true
}
