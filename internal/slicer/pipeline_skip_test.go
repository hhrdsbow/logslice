package slicer

import (
	"context"
	"testing"
)

func sendSkipLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectSkipLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewSkipStage_DefaultSkipsNothing(t *testing.T) {
	input := sendSkipLines([]string{"a", "b", "c"})
	out := NewSkipStage(context.Background(), input)
	got := collectSkipLines(out)
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(got), got)
	}
}

func TestNewSkipStage_WithOptions(t *testing.T) {
	input := sendSkipLines([]string{"1", "2", "3", "4", "5"})
	out := NewSkipStage(context.Background(), input, WithSkipN(3))
	got := collectSkipLines(out)
	want := []string{"4", "5"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("index %d: want %q, got %q", i, v, got[i])
		}
	}
}

func TestNewSkipStage_EmptyInput(t *testing.T) {
	input := sendSkipLines(nil)
	out := NewSkipStage(context.Background(), input, WithSkipN(5))
	got := collectSkipLines(out)
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %v", got)
	}
}

func TestNewSkipStage_CancelViaContext(t *testing.T) {
	ch := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	out := NewSkipStage(ctx, ch, WithSkipN(0))
	cancel()
	var count int
	for range out {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 lines after cancel, got %d", count)
	}
}
