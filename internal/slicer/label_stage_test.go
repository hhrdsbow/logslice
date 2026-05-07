package slicer

import (
	"testing"
)

func sendLabelLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectLabelLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewLabelStage_InvalidRegex(t *testing.T) {
	in := sendLabelLines([]string{"hello"})
	_, err := NewLabelStage(in, WithLabelRule("[", "bad"))
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNewLabelStage_NoRules_PassThrough(t *testing.T) {
	lines := []string{"line one", "line two"}
	in := sendLabelLines(lines)
	out, err := NewLabelStage(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectLabelLines(out)
	if len(got) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(got))
	}
	for i, g := range got {
		if g != lines[i] {
			t.Errorf("line %d: got %q, want %q", i, g, lines[i])
		}
	}
}

func TestNewLabelStage_AppliesLabels(t *testing.T) {
	lines := []string{"ERROR: boom", "INFO: ok", "WARN: careful"}
	in := sendLabelLines(lines)
	out, err := NewLabelStage(in,
		WithLabelRule(`ERROR`, "error"),
		WithLabelRule(`WARN`, "warning"),
		WithDefaultLabel("info"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectLabelLines(out)
	want := []string{
		"[error] ERROR: boom",
		"[info] INFO: ok",
		"[warning] WARN: careful",
	}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("line %d: got %q, want %q", i, g, want[i])
		}
	}
}

func TestNewLabelStage_CustomFormat(t *testing.T) {
	in := sendLabelLines([]string{"ERROR: oops"})
	out, err := NewLabelStage(in,
		WithLabelFormat("%s|%s"),
		WithLabelRule(`ERROR`, "err"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectLabelLines(out)
	want := "err|ERROR: oops"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %v, want [%q]", got, want)
	}
}
