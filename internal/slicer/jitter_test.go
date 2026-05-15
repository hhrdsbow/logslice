package slicer

import (
	"context"
	"testing"
	"time"
)

func sendJitterLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectJitterLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestJitterStage_NoDelay_PassesAll(t *testing.T) {
	input := []string{"alpha", "beta", "gamma"}
	ctx := context.Background()
	out := NewJitterStage(ctx, sendJitterLines(input))
	got := collectJitterLines(out)
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
	for i, l := range got {
		if l != input[i] {
			t.Errorf("line %d: want %q, got %q", i, input[i], l)
		}
	}
}

func TestJitterStage_WithDelay_PassesAll(t *testing.T) {
	input := []string{"line1", "line2", "line3"}
	ctx := context.Background()
	out := NewJitterStage(ctx, sendJitterLines(input), WithMaxJitter(2*time.Millisecond))
	got := collectJitterLines(out)
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
}

func TestJitterStage_NegativeDelayIgnored(t *testing.T) {
	input := []string{"x", "y"}
	ctx := context.Background()
	// negative duration should be ignored, lines pass through
	out := NewJitterStage(ctx, sendJitterLines(input), WithMaxJitter(-1*time.Millisecond))
	got := collectJitterLines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestJitterStage_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := make(chan string, 3)
	ch <- "a"
	ch <- "b"
	ch <- "c"
	close(ch)

	out := NewJitterStage(ctx, ch, WithMaxJitter(50*time.Millisecond))
	// drain whatever arrives; must not block
	done := make(chan struct{})
	go func() {
		for range out {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stage did not exit after context cancel")
	}
}

func TestJitterStage_EmptyInput(t *testing.T) {
	ctx := context.Background()
	ch := make(chan string)
	close(ch)
	out := NewJitterStage(ctx, ch, WithMaxJitter(time.Millisecond))
	got := collectJitterLines(out)
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}
