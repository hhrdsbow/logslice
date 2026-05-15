package slicer

import (
	"context"
	"testing"
)

func sendGrepLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectGrepLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewGrepStage_MissingPattern(t *testing.T) {
	ctx := context.Background()
	in := sendGrepLines([]string{"hello"})
	_, err := NewGrepStage(ctx, in)
	if err == nil {
		t.Fatal("expected error when no pattern provided")
	}
}

func TestNewGrepStage_InvalidPattern(t *testing.T) {
	ctx := context.Background()
	in := sendGrepLines([]string{"hello"})
	_, err := NewGrepStage(ctx, in, WithGrepPattern("[bad"))
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestNewGrepStage_PassesMatchingLines(t *testing.T) {
	ctx := context.Background()
	in := sendGrepLines([]string{"info: started", "error: boom", "info: done"})
	out, err := NewGrepStage(ctx, in, WithGrepPattern(`error`))
	if err != nil {
		t.Fatal(err)
	}
	lines := collectGrepLines(out)
	if len(lines) != 1 || lines[0] != "error: boom" {
		t.Fatalf("unexpected lines: %v", lines)
	}
}

func TestNewGrepStage_InvertedStage(t *testing.T) {
	ctx := context.Background()
	in := sendGrepLines([]string{"info: started", "error: boom", "info: done"})
	out, err := NewGrepStage(ctx, in, WithGrepPattern(`error`), WithGrepStageInvert())
	if err != nil {
		t.Fatal(err)
	}
	lines := collectGrepLines(out)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestNewGrepStage_EmptyInput(t *testing.T) {
	ctx := context.Background()
	in := sendGrepLines(nil)
	out, err := NewGrepStage(ctx, in, WithGrepPattern(`error`))
	if err != nil {
		t.Fatal(err)
	}
	lines := collectGrepLines(out)
	if len(lines) != 0 {
		t.Fatalf("expected empty output, got %v", lines)
	}
}

func TestNewGrepStage_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := sendGrepLines([]string{"error: too late"})
	out, err := NewGrepStage(ctx, in, WithGrepPattern(`error`))
	if err != nil {
		t.Fatal(err)
	}
	// drain; should not block
	collectGrepLines(out)
}
