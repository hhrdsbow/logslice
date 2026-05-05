package slicer

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewPatternSlicer_InvalidRegex(t *testing.T) {
	_, err := NewPatternSlicer("[invalid", false)
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestPatternSlicer_Matches(t *testing.T) {
	ps, err := NewPatternSlicer(`ERROR`, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ps.Matches("2024-01-01 ERROR something went wrong") {
		t.Error("expected match for line containing ERROR")
	}
	if ps.Matches("2024-01-01 INFO all good") {
		t.Error("expected no match for line without ERROR")
	}
}

func TestPatternSlicer_MatchesInvert(t *testing.T) {
	ps, err := NewPatternSlicer(`DEBUG`, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ps.Matches("2024-01-01 DEBUG verbose output") {
		t.Error("expected no match (inverted) for DEBUG line")
	}
	if !ps.Matches("2024-01-01 INFO normal line") {
		t.Error("expected match (inverted) for non-DEBUG line")
	}
}

func TestPatternSlicer_Slice(t *testing.T) {
	input := strings.Join([]string{
		"2024-01-01 INFO startup",
		"2024-01-01 ERROR disk full",
		"2024-01-01 WARN low memory",
		"2024-01-01 ERROR connection refused",
		"2024-01-01 INFO shutdown",
	}, "\n")

	ps, err := NewPatternSlicer(`ERROR`, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	count, err := ps.Slice(strings.NewReader(input), &buf)
	if err != nil {
		t.Fatalf("Slice returned error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 lines written, got %d", count)
	}

	output := buf.String()
	if !strings.Contains(output, "disk full") {
		t.Error("expected output to contain 'disk full'")
	}
	if !strings.Contains(output, "connection refused") {
		t.Error("expected output to contain 'connection refused'")
	}
	if strings.Contains(output, "INFO") {
		t.Error("expected output to not contain INFO lines")
	}
}

func TestPatternSlicer_SliceEmpty(t *testing.T) {
	ps, err := NewPatternSlicer(`ERROR`, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	count, err := ps.Slice(strings.NewReader(""), &buf)
	if err != nil {
		t.Fatalf("Slice returned error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 lines written, got %d", count)
	}
}
