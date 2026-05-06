package slicer

import (
	"strings"
	"unicode"
)

// TransformFunc is a function that transforms a log line.
type TransformFunc func(line string) string

// Transformer applies a chain of TransformFuncs to each line.
type Transformer struct {
	fns []TransformFunc
}

// NewTransformer creates a Transformer with the given transform functions.
func NewTransformer(fns ...TransformFunc) *Transformer {
	return &Transformer{fns: fns}
}

// Apply runs all transform functions on the given line in order.
func (t *Transformer) Apply(line string) string {
	for _, fn := range t.fns {
		line = fn(line)
	}
	return line
}

// TrimSpaceTransform returns a TransformFunc that trims leading/trailing whitespace.
func TrimSpaceTransform() TransformFunc {
	return func(line string) string {
		return strings.TrimSpace(line)
	}
}

// LowercaseTransform returns a TransformFunc that lowercases the entire line.
func LowercaseTransform() TransformFunc {
	return func(line string) string {
		return strings.ToLower(line)
	}
}

// RedactTransform returns a TransformFunc that replaces occurrences of target
// with the given replacement string.
func RedactTransform(target, replacement string) TransformFunc {
	return func(line string) string {
		return strings.ReplaceAll(line, target, replacement)
	}
}

// TruncateTransform returns a TransformFunc that truncates lines to maxLen runes.
func TruncateTransform(maxLen int) TransformFunc {
	return func(line string) string {
		runes := []rune(line)
		if len(runes) <= maxLen {
			return line
		}
		return string(runes[:maxLen])
	}
}

// StripControlTransform returns a TransformFunc that removes non-printable control characters.
func StripControlTransform() TransformFunc {
	return func(line string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsControl(r) && r != '\t' {
				return -1
			}
			return r
		}, line)
	}
}
