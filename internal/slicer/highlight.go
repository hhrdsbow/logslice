package slicer

import (
	"fmt"
	"regexp"
	"strings"
)

// HighlightFunc wraps a matched portion of a line with a prefix and suffix.
type HighlightFunc func(line string) string

// Highlighter applies ANSI colour highlighting to matched patterns in log lines.
type Highlighter struct {
	pattern *regexp.Regexp
	prefix  string
	suffix  string
}

// NewHighlighter creates a Highlighter that wraps regex matches with prefix/suffix.
// Use ANSI escape codes for colour, e.g. prefix="\033[31m", suffix="\033[0m".
func NewHighlighter(pattern, prefix, suffix string) (*Highlighter, error) {
	if pattern == "" {
		return nil, fmt.Errorf("highlight: pattern must not be empty")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("highlight: invalid pattern: %w", err)
	}
	return &Highlighter{pattern: re, prefix: prefix, suffix: suffix}, nil
}

// Apply wraps every match in the line with the configured prefix and suffix.
func (h *Highlighter) Apply(line string) string {
	return h.pattern.ReplaceAllStringFunc(line, func(m string) string {
		return h.prefix + m + h.suffix
	})
}

// HighlightTransform returns a TransformFunc suitable for use with NewTransformer.
func (h *Highlighter) HighlightTransform() TransformFunc {
	return func(line string) string {
		return h.Apply(line)
	}
}

// ANSIStrip removes all ANSI escape sequences from a line.
func ANSIStrip(line string) string {
	const ansiEsc = "\033["
	var b strings.Builder
	for i := 0; i < len(line); {
		if i+1 < len(line) && line[i] == '\033' && line[i+1] == '[' {
			// skip until 'm'
			j := i + 2
			for j < len(line) && line[j] != 'm' {
				j++
			}
			if j < len(line) {
				j++ // skip 'm'
			}
			i = j
			_ = ansiEsc
		} else {
			b.WriteByte(line[i])
			i++
		}
	}
	return b.String()
}

// NewHighlightStage returns a pipeline stage that highlights matches in each line.
func NewHighlightStage(in <-chan string, h *Highlighter) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for line := range in {
			out <- h.Apply(line)
		}
	}()
	return out
}
