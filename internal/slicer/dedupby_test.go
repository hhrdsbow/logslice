package slicer

import (
	"testing"
)

func TestDedupByFilter_WholeLine_Default(t *testing.T) {
	f := NewDedupByFilter(nil)
	if !f.Accept("hello") {
		t.Fatal("expected first occurrence to be accepted")
	}
	if f.Accept("hello") {
		t.Fatal("expected duplicate to be rejected")
	}
	if !f.Accept("world") {
		t.Fatal("expected distinct line to be accepted")
	}
}

func TestDedupByFilter_CustomExtractor(t *testing.T) {
	// Extract first word as key.
	extract := func(line string) string {
		for i, c := range line {
			if c == ' ' {
				return line[:i]
			}
		}
		return line
	}
	f := NewDedupByFilter(extract)
	if !f.Accept("ERROR something happened") {
		t.Fatal("first ERROR line should be accepted")
	}
	// Same key "ERROR", different suffix — should be deduped.
	if f.Accept("ERROR another message") {
		t.Fatal("second ERROR line should be rejected by key")
	}
	if !f.Accept("WARN low memory") {
		t.Fatal("WARN line should be accepted")
	}
}

func TestNewRegexDedupByFilter_InvalidRegex(t *testing.T) {
	_, err := NewRegexDedupByFilter("[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNewRegexDedupByFilter_EmptyPattern(t *testing.T) {
	_, err := NewRegexDedupByFilter("")
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewRegexDedupByFilter_CaptureGroup(t *testing.T) {
	f, err := NewRegexDedupByFilter(`level=(\w+)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.Accept("ts=1 level=error msg=foo") {
		t.Fatal("first error-level line should be accepted")
	}
	if f.Accept("ts=2 level=error msg=bar") {
		t.Fatal("second error-level line should be deduped by capture group")
	}
	if !f.Accept("ts=3 level=warn msg=baz") {
		t.Fatal("warn-level line should be accepted")
	}
}

func TestDedupByFilter_Reset(t *testing.T) {
	f := NewDedupByFilter(nil)
	f.Accept("line")
	f.Reset()
	if !f.Accept("line") {
		t.Fatal("after Reset, line should be accepted again")
	}
}

func TestNewDedupByStageFromOptions_NoOptions(t *testing.T) {
	stage, err := NewDedupByStageFromOptions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	in := make(chan string, 4)
	in <- "a"
	in <- "a"
	in <- "b"
	in <- "b"
	close(in)
	var got []string
	for line := range stage(in) {
		got = append(got, line)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected output: %v", got)
	}
}

func TestNewDedupByStageFromOptions_WithPattern(t *testing.T) {
	stage, err := NewDedupByStageFromOptions(WithDedupByPattern(`(\w+)`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	in := make(chan string, 3)
	in <- "hello world"
	in <- "hello again"
	in <- "bye"
	close(in)
	var got []string
	for line := range stage(in) {
		got = append(got, line)
	}
	// First word "hello" deduped; "bye" is new.
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
}

func TestNewDedupByStageFromOptions_InvalidPattern(t *testing.T) {
	_, err := NewDedupByStageFromOptions(WithDedupByPattern("[bad"))
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}
