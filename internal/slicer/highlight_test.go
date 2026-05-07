package slicer

import (
	"strings"
	"testing"
)

func TestNewHighlighter_EmptyPattern(t *testing.T) {
	_, err := NewHighlighter("", "[", "]")
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewHighlighter_InvalidRegex(t *testing.T) {
	_, err := NewHighlighter("[invalid", "[", "]")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestHighlighter_Apply_NoMatch(t *testing.T) {
	h, err := NewHighlighter(`ERROR`, "<<", ">>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := h.Apply("INFO everything is fine")
	if got != "INFO everything is fine" {
		t.Errorf("expected unchanged line, got %q", got)
	}
}

func TestHighlighter_Apply_Match(t *testing.T) {
	h, err := NewHighlighter(`ERROR`, "<<", ">>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := h.Apply("2024/01/01 ERROR something failed")
	if !strings.Contains(got, "<<ERROR>>") {
		t.Errorf("expected <<ERROR>> in output, got %q", got)
	}
}

func TestHighlighter_Apply_MultipleMatches(t *testing.T) {
	h, err := NewHighlighter(`\d+`, "[", "]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := h.Apply("line 42 and 99")
	if got != "line [42] and [99]" {
		t.Errorf("expected 'line [42] and [99]', got %q", got)
	}
}

func TestHighlighter_HighlightTransform(t *testing.T) {
	h, err := NewHighlighter(`WARN`, ">", "<")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn := h.HighlightTransform()
	got := fn("WARN: disk low")
	if !strings.Contains(got, ">WARN<") {
		t.Errorf("transform did not highlight, got %q", got)
	}
}

func TestANSIStrip(t *testing.T) {
	input := "\033[31mERROR\033[0m: something bad"
	got := ANSIStrip(input)
	expected := "ERROR: something bad"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestANSIStrip_NoEscapes(t *testing.T) {
	input := "plain text line"
	got := ANSIStrip(input)
	if got != input {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestNewHighlightStage(t *testing.T) {
	h, err := NewHighlighter(`ERROR`, "[", "]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	in := make(chan string, 3)
	in <- "INFO ok"
	in <- "ERROR bad"
	in <- "DEBUG fine"
	close(in)

	out := NewHighlightStage(in, h)
	var results []string
	for line := range out {
		results = append(results, line)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(results))
	}
	if results[0] != "INFO ok" {
		t.Errorf("line 0: expected 'INFO ok', got %q", results[0])
	}
	if !strings.Contains(results[1], "[ERROR]") {
		t.Errorf("line 1: expected [ERROR] highlight, got %q", results[1])
	}
	if results[2] != "DEBUG fine" {
		t.Errorf("line 2: expected 'DEBUG fine', got %q", results[2])
	}
}
