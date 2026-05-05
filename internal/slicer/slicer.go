package slicer

import (
	"bufio"
	"io"
)

// LineMatcher is implemented by any slicer that can determine
// whether a single log line should be included in output.
type LineMatcher interface {
	Matches(line string) bool
}

// Slicer streams lines from a reader to a writer based on a LineMatcher.
type Slicer struct {
	matcher LineMatcher
}

// New creates a Slicer using the provided LineMatcher.
func New(matcher LineMatcher) *Slicer {
	return &Slicer{matcher: matcher}
}

// SliceResult holds statistics from a Slice operation.
type SliceResult struct {
	LinesRead    int
	LinesWritten int
}

// Slice reads all lines from r, writing lines accepted by the matcher to w.
// It returns a SliceResult with read/write counts and any error encountered.
func (s *Slicer) Slice(r io.Reader, w io.Writer) (SliceResult, error) {
	scanner := bufio.NewScanner(r)
	bw := bufio.NewWriter(w)
	result := SliceResult{}

	for scanner.Scan() {
		line := scanner.Text()
		result.LinesRead++

		if s.matcher.Matches(line) {
			if _, err := bw.WriteString(line + "\n"); err != nil {
				return result, err
			}
			result.LinesWritten++
		}
	}

	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, bw.Flush()
}
