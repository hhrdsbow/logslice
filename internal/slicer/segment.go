package slicer

import (
	"fmt"
	"time"
)

// Segment represents a named slice of log lines.
type Segment struct {
	Name  string
	Lines []string
	Start time.Time
	End   time.Time
}

// NewSegment creates a new named Segment.
func NewSegment(name string) *Segment {
	return &Segment{Name: name}
}

// Add appends a line to the segment.
func (s *Segment) Add(line string) {
	s.Lines = append(s.Lines, line)
}

// Len returns the number of lines in the segment.
func (s *Segment) Len() int {
	return len(s.Lines)
}

// IsEmpty reports whether the segment contains no lines.
func (s *Segment) IsEmpty() bool {
	return len(s.Lines) == 0
}

// WriteTo writes all lines in the segment to the given OutputWriter.
func (s *Segment) WriteTo(ow *OutputWriter) error {
	for _, line := range s.Lines {
		if err := ow.WriteLine(line); err != nil {
			return fmt.Errorf("segment %q: write line: %w", s.Name, err)
		}
	}
	return nil
}

// Summary returns a human-readable summary of the segment.
func (s *Segment) Summary() string {
	if s.Start.IsZero() || s.End.IsZero() {
		return fmt.Sprintf("segment=%q lines=%d", s.Name, s.Len())
	}
	return fmt.Sprintf("segment=%q lines=%d start=%s end=%s",
		s.Name, s.Len(),
		s.Start.Format(time.RFC3339),
		s.End.Format(time.RFC3339),
	)
}
