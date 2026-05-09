package slicer

import (
	"context"
	"testing"
	"time"
)

func sendHeadLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectHeadLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewHeadStage_DefaultMax(t *testing.T) {
	stage := NewHeadStage()
	var lines []string
	for i := 0; i < 15; i++ {
		lines = append(lines, "line")
	}
	out := collectHeadLines(stage(context.Background(), sendHeadLines(lines)))
	if len(out) != 10 {
		t.Fatalf("expected default max 10, got %d", len(out))
	}
}

func TestNewHeadStage_WithOptions(t *testing.T) {
	stage := NewHeadStage(WithHeadMax(4))
	in := sendHeadLines([]string{"a", "b", "c", "d", "e", "f"})
	out := collectHeadLines(stage(context.Background(), in))
	if len(out) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(out))
	}
}

func TestNewHeadStage_EmptyInput(t *testing.T) {
	stage := NewHeadStage(WithHeadMax(5))
	out := collectHeadLines(stage(context.Background(), sendHeadLines(nil)))
	if len(out) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(out))
	}
}

func TestNewHeadStage_CancelViaContext(t *testing.T) {
	stage := NewHeadStage(WithHeadMax(100))
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan string)
	out := stage(ctx, ch)
	cancel()
	select {
	case <-out:
	case <-time.After(time.Second):
		t.Fatal("stage did not close after context cancel")
	}
}
