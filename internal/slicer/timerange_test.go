package slicer

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// sampleParser parses lines prefixed with "2006-01-02T15:04:05 ".
func sampleParser(line string) (time.Time, bool) {
	if len(line) < 20 {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02T15:04:05", line[:19])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func mustTime(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		panic(err)
	}
	return t
}

const sampleLog = `2024-01-01T10:00:00 startup complete
2024-01-01T10:05:00 request received
2024-01-01T10:10:00 processing done
2024-01-01T10:15:00 error occurred
2024-01-01T10:20:00 shutdown initiated
`

func TestTimeRangeSlicer_MatchesRange(t *testing.T) {
	s := &TimeRangeSlicer{
		Range:  TimeRange{Start: mustTime("2024-01-01T10:05:00"), End: mustTime("2024-01-01T10:15:00")},
		Parser: sampleParser,
	}
	var out bytes.Buffer
	n, err := s.Slice(strings.NewReader(sampleLog), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 lines, got %d", n)
	}
	if !strings.Contains(out.String(), "request received") {
		t.Error("expected 'request received' in output")
	}
}

func TestTimeRangeSlicer_NoMatch(t *testing.T) {
	s := &TimeRangeSlicer{
		Range:  TimeRange{Start: mustTime("2024-01-01T11:00:00"), End: mustTime("2024-01-01T12:00:00")},
		Parser: sampleParser,
	}
	var out bytes.Buffer
	n, err := s.Slice(strings.NewReader(sampleLog), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 lines, got %d", n)
	}
}

func TestTimeRangeSlicer_NilParser(t *testing.T) {
	s := &TimeRangeSlicer{}
	var out bytes.Buffer
	_, err := s.Slice(strings.NewReader(sampleLog), &out)
	if err == nil {
		t.Error("expected error for nil parser")
	}
}
