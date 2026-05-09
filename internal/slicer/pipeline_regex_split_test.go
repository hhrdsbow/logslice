package slicer

import (
	"context"
	"testing"
	"time"
)

func sendSplitLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectSplitOutput(ctx context.Context, ch <-chan string) []string {
	var out []string
	for {
		select {
		case l, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, l)
		case <-ctx.Done():
			return out
		}
	}
}

func TestNewRegexSplitStage_MissingPattern(t *testing.T) {
	_, err := NewRegexSplitStage(context.Background(), make(chan string))
	if err == nil {
		t.Fatal("expected error when pattern is missing")
	}
}

func TestNewRegexSplitStage_InvalidPattern(t *testing.T) {
	_, err := NewRegexSplitStage(context.Background(), make(chan string),
		WithSplitPattern("[bad"),
	)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNewRegexSplitStage_PassesLines(t *testing.T) {
	ctx := context.Background()
	input := []string{"a", "b", "---", "c", "d"}
	ch, err := NewRegexSplitStage(ctx, sendSplitLines(input),
		WithSplitPattern(`^---`),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectSplitOutput(ctx, ch)
	if len(got) != len(input) {
		t.Errorf("expected %d lines, got %d: %v", len(input), len(got), got)
	}
}

func TestNewRegexSplitStage_WithSeparator(t *testing.T) {
	ctx := context.Background()
	input := []string{"a", "---", "b"}
	ch, err := NewRegexSplitStage(ctx, sendSplitLines(input),
		WithSplitPattern(`^---`),
		WithSegmentSeparator("==="),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectSplitOutput(ctx, ch)
	// expect: [a] === [--- b]  => "a", "===", "---", "b"
	if len(got) != 4 {
		t.Fatalf("expected 4 lines with separator, got %d: %v", len(got), got)
	}
	if got[1] != "===" {
		t.Errorf("expected separator at index 1, got %q", got[1])
	}
}

func TestNewRegexSplitStage_EmptyInput(t *testing.T) {
	ctx := context.Background()
	ch, err := NewRegexSplitStage(ctx, sendSplitLines(nil),
		WithSplitPattern(`^---`),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectSplitOutput(ctx, ch)
	if len(got) != 0 {
		t.Errorf("expected no output for empty input, got %v", got)
	}
}

func TestNewRegexSplitStage_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	blocking := make(chan string)
	ch, err := NewRegexSplitStage(ctx, blocking, WithSplitPattern(`^---`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := collectSplitOutput(ctx, ch)
	if len(got) != 0 {
		t.Errorf("expected no output on cancel, got %v", got)
	}
}
