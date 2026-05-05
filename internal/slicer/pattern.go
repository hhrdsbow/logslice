package slicer

import (
	"bufio"
	"io"
	"regexp"
)

// PatternSlicer segments log lines based on a regular expression pattern.
// Lines matching the pattern are written to the output writer.
type PatternSlicer struct {
	pattern  *regexp.Regexp
	invert   bool
}

// NewPatternSlicer creates a PatternSlicer that matches lines against the given pattern.
// If invert is true, lines that do NOT match are included instead.
func NewPatternSlicer(pattern string, invert bool) (*PatternSlicer, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &PatternSlicer{pattern: re, invert: invert}, nil
}

// Matches reports whether the given line should be included in the output.
func (p *PatternSlicer) Matches(line string) bool {
	matched := p.pattern.MatchString(line)
	if p.invert {
		return !matched
	}
	return matched
}

// Slice reads lines from r, writing matching lines to w.
// Returns the number of lines written and any error encountered.
func (p *PatternSlicer) Slice(r io.Reader, w io.Writer) (int, error) {
	scanner := bufio.NewScanner(r)
	bw := bufio.NewWriter(w)
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if p.Matches(line) {
			if _, err := bw.WriteString(line + "\n"); err != nil {
				return count, err
			}
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return count, err
	}
	return count, bw.Flush()
}
