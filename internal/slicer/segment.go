package slicer

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Segment holds a named collection of log lines with optional time metadata.
type Segment struct {
	Name      string
	lines     []string
	bytes     int64
	FirstTime *time.Time
	LastTime  *time.Time
}

// NewSegment creates an empty Segment with the given name.
func NewSegment(name string) *Segment {
	return &Segment{Name: name}
}

// Add appends a line to the segment.
func (s *Segment) Add(line string) {
	s.lines = append(s.lines, line)
	s.bytes += int64(len(line)) + 1 // +1 for newline
}

// Len returns the number of lines in the segment.
func (s *Segment) Len() int {
	return len(s.lines)
}

// Bytes returns the total byte size of the segment content.
func (s *Segment) Bytes() int64 {
	return s.bytes
}

// IsEmpty reports whether the segment contains no lines.
func (s *Segment) IsEmpty() bool {
	return len(s.lines) == 0
}

// WriteTo writes all lines to w, each terminated by a newline.
func (s *Segment) WriteTo(w io.Writer) (int64, error) {
	var total int64
	for _, line := range s.lines {
		n, err := fmt.Fprintln(w, line)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// Summary returns a human-readable description of the segment.
func (s *Segment) Summary() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "segment %q: %d lines, %d bytes", s.Name, s.Len(), s.Bytes())
	if s.FirstTime != nil && s.LastTime != nil {
		fmt.Fprintf(&sb, " [%s – %s]", s.FirstTime.Format(time.RFC3339), s.LastTime.Format(time.RFC3339))
	}
	return sb.String()
}
