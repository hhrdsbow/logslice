package slicer

import (
	"testing"
)

func TestNewFilter_InvalidInclude(t *testing.T) {
	_, err := NewFilter(WithInclude("[invalid"))
	if err == nil {
		t.Fatal("expected error for invalid include pattern")
	}
}

func TestNewFilter_InvalidExclude(t *testing.T) {
	_, err := NewFilter(WithExclude("[invalid"))
	if err == nil {
		t.Fatal("expected error for invalid exclude pattern")
	}
}

func TestFilter_NoRules_AcceptsAll(t *testing.T) {
	f, err := NewFilter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := []string{"hello", "world", "foo bar"}
	for _, l := range lines {
		if !f.Accept(l) {
			t.Errorf("expected line %q to be accepted", l)
		}
	}
}

func TestFilter_Include(t *testing.T) {
	f, err := NewFilter(WithInclude(`ERROR`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.Accept("2024/01/01 ERROR something failed") {
		t.Error("expected ERROR line to be accepted")
	}
	if f.Accept("2024/01/01 INFO all good") {
		t.Error("expected INFO line to be rejected")
	}
}

func TestFilter_Exclude(t *testing.T) {
	f, err := NewFilter(WithExclude(`DEBUG`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Accept("2024/01/01 DEBUG verbose output") {
		t.Error("expected DEBUG line to be excluded")
	}
	if !f.Accept("2024/01/01 INFO important") {
		t.Error("expected INFO line to pass through")
	}
}

func TestFilter_IncludeAndExclude(t *testing.T) {
	// Include ERROR lines but exclude those containing "timeout".
	f, err := NewFilter(WithInclude(`ERROR`), WithExclude(`timeout`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.Accept("ERROR: disk full") {
		t.Error("expected 'ERROR: disk full' to be accepted")
	}
	if f.Accept("ERROR: timeout waiting for response") {
		t.Error("expected timeout ERROR to be excluded")
	}
	if f.Accept("INFO: all good") {
		t.Error("expected INFO line to be rejected (no include match)")
	}
}

func TestFilter_Apply(t *testing.T) {
	f, err := NewFilter(WithInclude(`WARN|ERROR`), WithExclude(`noisy`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	input := []string{
		"INFO: startup",
		"WARN: low memory",
		"ERROR: noisy subsystem crash",
		"ERROR: real problem",
	}
	got := f.Apply(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
	if got[0] != "WARN: low memory" {
		t.Errorf("unexpected first line: %q", got[0])
	}
	if got[1] != "ERROR: real problem" {
		t.Errorf("unexpected second line: %q", got[1])
	}
}
