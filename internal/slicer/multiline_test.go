package slicer

import (
	"testing"
)

func sendMultilines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectMultilines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestMultilineFolder_NoPatterns_Error(t *testing.T) {
	_, err := NewMultilineFolder()
	if err == nil {
		t.Fatal("expected error with no patterns")
	}
}

func TestMultilineFolder_StartOnly(t *testing.T) {
	f, err := NewMultilineFolder(WithStartPattern(`^START`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := []string{"START a", "continuation", "START b", "more"}
	out := f.Fold(sendMultilines(lines), make(chan struct{}))
	got := collectMultilines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 folded lines, got %d: %v", len(got), got)
	}
	if got[0] != "START a continuation" {
		t.Errorf("unexpected first line: %q", got[0])
	}
	if got[1] != "START b more" {
		t.Errorf("unexpected second line: %q", got[1])
	}
}

func TestMultilineFolder_ContinueOnly(t *testing.T) {
	f, err := NewMultilineFolder(WithContinuePattern(`^\s+`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := []string{"first", "  cont1", "  cont2", "second"}
	out := f.Fold(sendMultilines(lines), make(chan struct{}))
	got := collectMultilines(out)
	// "first" is emitted as-is; "  cont1" starts a new block; "  cont2" appended;
	// "second" flushes block then emitted
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(got), got)
	}
}

func TestMultilineFolder_CustomSep(t *testing.T) {
	f, err := NewMultilineFolder(
		WithStartPattern(`^ERR`),
		WithJoinSep("|"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := []string{"ERR x", "detail"}
	out := f.Fold(sendMultilines(lines), make(chan struct{}))
	got := collectMultilines(out)
	if len(got) != 1 || got[0] != "ERR x|detail" {
		t.Errorf("unexpected output: %v", got)
	}
}

func TestNewMultilineStage_MissingPatterns(t *testing.T) {
	done := make(chan struct{})
	_, err := NewMultilineStage(make(chan string), done)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewMultilineStage_FoldsLines(t *testing.T) {
	done := make(chan struct{})
	in := sendMultilines([]string{"START line1", "cont", "START line2"})
	out, err := NewMultilineStage(in, done,
		WithMultilineStart(`^START`),
		WithMultilineJoinSep(" "),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectMultilines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(got), got)
	}
	if got[0] != "START line1 cont" {
		t.Errorf("unexpected: %q", got[0])
	}
}

func TestMultilineFolder_CancelViaDone(t *testing.T) {
	f, _ := NewMultilineFolder(WithStartPattern(`^X`))
	done := make(chan struct{})
	in := make(chan string)
	out := f.Fold(in, done)
	close(done)
	for range out {
	}
}
