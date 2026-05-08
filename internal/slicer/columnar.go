package slicer

import (
	"fmt"
	"regexp"
	"strings"
)

// ColumnExtractor extracts named columns from log lines using a delimiter or regex.
type ColumnExtractor struct {
	headers []string
	splitRe *regexp.Regexp
	delim   string
}

// ColumnExtractorOption configures a ColumnExtractor.
type ColumnExtractorOption func(*ColumnExtractor)

// WithDelimiter sets a literal string delimiter for splitting columns.
func WithDelimiter(delim string) ColumnExtractorOption {
	return func(c *ColumnExtractor) {
		c.delim = delim
		c.splitRe = nil
	}
}

// WithColumnRegex sets a regex whose capture groups become column values.
func WithColumnRegex(pattern string) (ColumnExtractorOption, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("columnar: invalid regex: %w", err)
	}
	return func(c *ColumnExtractor) {
		c.splitRe = re
		c.delim = ""
	}, nil
}

// NewColumnExtractor creates a ColumnExtractor with the given column headers.
// By default it splits on whitespace.
func NewColumnExtractor(headers []string, opts ...ColumnExtractorOption) *ColumnExtractor {
	c := &ColumnExtractor{
		headers: headers,
		delim:   "",
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Extract parses a line and returns a map of header -> value.
// Missing columns are set to empty string.
func (c *ColumnExtractor) Extract(line string) map[string]string {
	var parts []string
	switch {
	case c.splitRe != nil:
		m := c.splitRe.FindStringSubmatch(line)
		if len(m) > 1 {
			parts = m[1:]
		}
	case c.delim != "":
		parts = strings.Split(line, c.delim)
	default:
		parts = strings.Fields(line)
	}

	out := make(map[string]string, len(c.headers))
	for i, h := range c.headers {
		if i < len(parts) {
			out[h] = parts[i]
		} else {
			out[h] = ""
		}
	}
	return out
}

// Format reassembles extracted columns into a line using the given template.
// Template tokens are {{column_name}}.
func (c *ColumnExtractor) Format(cols map[string]string, tmpl string) string {
	result := tmpl
	for k, v := range cols {
		result = strings.ReplaceAll(result, "{"+"{"+k+"}}", v)
	}
	return result
}
