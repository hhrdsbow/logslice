package slicer

import (
	"fmt"
	"regexp"
)

// Filter defines criteria for including or excluding log lines.
type Filter struct {
	include []*regexp.Regexp
	exclude []*regexp.Regexp
}

// FilterOption configures a Filter.
type FilterOption func(*Filter) error

// WithInclude adds a pattern that a line must match to be kept.
func WithInclude(pattern string) FilterOption {
	return func(f *Filter) error {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid include pattern %q: %w", pattern, err)
		}
		f.include = append(f.include, re)
		return nil
	}
}

// WithExclude adds a pattern that, if matched, causes the line to be dropped.
func WithExclude(pattern string) FilterOption {
	return func(f *Filter) error {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
		f.exclude = append(f.exclude, re)
		return nil
	}
}

// NewFilter constructs a Filter from the provided options.
func NewFilter(opts ...FilterOption) (*Filter, error) {
	f := &Filter{}
	for _, opt := range opts {
		if err := opt(f); err != nil {
			return nil, err
		}
	}
	return f, nil
}

// Accept returns true when the line passes all include and exclude rules.
// If no include patterns are set every line is accepted unless excluded.
func (f *Filter) Accept(line string) bool {
	for _, re := range f.exclude {
		if re.MatchString(line) {
			return false
		}
	}
	if len(f.include) == 0 {
		return true
	}
	for _, re := range f.include {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// Apply filters a slice of lines, returning only those accepted.
func (f *Filter) Apply(lines []string) []string {
	out := lines[:0:0]
	for _, l := range lines {
		if f.Accept(l) {
			out = append(out, l)
		}
	}
	return out
}
