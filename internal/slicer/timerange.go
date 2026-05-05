package slicer

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

// TimeRange defines the start and end boundaries for log slicing.
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// TimeParser is a function that extracts a time.Time from a log line.
type TimeParser func(line string) (time.Time, bool)

// TimeRangeSlicer streams log lines that fall within the given TimeRange.
type TimeRangeSlicer struct {
	Range  TimeRange
	Parser TimeParser
}

// Slice reads from r, writing lines within the time range to w.
// It returns the number of lines written and any error encountered.
func (s *TimeRangeSlicer) Slice(r io.Reader, w io.Writer) (int, error) {
	if s.Parser == nil {
		return 0, fmt.Errorf("slicer: TimeParser must not be nil")
	}

	scanner := bufio.NewScanner(r)
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		t, ok := s.Parser(line)
		if !ok {
			// Lines without a parseable timestamp are skipped.
			continue
		}
		if (t.Equal(s.Range.Start) || t.After(s.Range.Start)) &&
			(t.Equal(s.Range.End) || t.Before(s.Range.End)) {
			if _, err := fmt.Fprintln(bw, line); err != nil {
				return count, fmt.Errorf("slicer: write error: %w", err)
			}
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("slicer: scan error: %w", err)
	}
	return count, nil
}
