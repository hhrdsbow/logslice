package slicer

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSegment_AddAndLen(t *testing.T) {
	s := NewSegment("test")
	if s.Len() != 0 {
		t.Errorf("expected 0, got %d", s.Len())
	}
	s.Add("line 1")
	s.Add("line 2")
	if s.Len() != 2 {
		t.Errorf("expected 2, got %d", s.Len())
	}
}

func TestSegment_IsEmpty(t *testing.T) {
	s := NewSegment("empty")
	if !s.IsEmpty() {
		t.Error("expected empty segment")
	}
	s.Add("data")
	if s.IsEmpty() {
		t.Error("expected non-empty segment")
	}
}

func TestSegment_WriteTo(t *testing.T) {
	s := NewSegment("write-test")
	s.Add("alpha")
	s.Add("beta")
	s.Add("gamma")

	var buf bytes.Buffer
	ow, err := NewOutputWriter(OutputConfig{Mode: OutputStdout, Writer: &buf})
	if err != nil {
		t.Fatalf("NewOutputWriter error: %v", err)
	}
	if err := s.WriteTo(ow); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	_ = ow.Close()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "alpha" || lines[2] != "gamma" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestSegment_Summary_NoTime(t *testing.T) {
	s := NewSegment("no-time")
	s.Add("x")
	got := s.Summary()
	if !strings.Contains(got, "no-time") || !strings.Contains(got, "lines=1") {
		t.Errorf("unexpected summary: %s", got)
	}
}

func TestSegment_Summary_WithTime(t *testing.T) {
	s := NewSegment("timed")
	s.Start = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s.End = time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	s.Add("line")
	got := s.Summary()
	if !strings.Contains(got, "2024-01-01T00:00:00Z") {
		t.Errorf("unexpected summary: %s", got)
	}
	if !strings.Contains(got, "2024-01-01T01:00:00Z") {
		t.Errorf("unexpected summary: %s", got)
	}
}
