package slicer

import (
	"context"
	"strings"
	"testing"
	"time"
)

func sendPipelineContextLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func drainContextStage(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewContextStage_DefaultNoMatch(t *testing.T) {
	in := sendPipelineContextLines([]string{"a", "b", "c"})
	out := NewContextStage(context.Background(), in)
	got := drainContextStage(out)
	if len(got) != 0 {
		t.Fatalf("expected no output with default matcher, got %v", got)
	}
}

func TestNewContextStage_WithMatcher(t *testing.T) {
	in := sendPipelineContextLines([]string{"foo", "ERROR here", "bar"})
	matcher := func(l string) bool { return strings.Contains(l, "ERROR") }
	out := NewContextStage(context.Background(), in,
		WithContextMatcher(matcher),
		WithContextBefore(1),
		WithContextAfter(1),
	)
	got := drainContextStage(out)
	want := []string{"foo", "ERROR here", "bar"}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] want %q got %q", i, want[i], got[i])
		}
	}
}

func TestNewContextStage_EmptyInput(t *testing.T) {
	in := sendPipelineContextLines(nil)
	out := NewContextStage(context.Background(), in,
		WithContextMatcher(func(string) bool { return true }),
	)
	got := drainContextStage(out)
	if len(got) != 0 {
		t.Fatalf("expected empty output, got %v", got)
	}
}

func TestNewContextStage_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := make(chan string)
	out := NewContextStage(ctx, in,
		WithContextMatcher(func(string) bool { return true }),
	)
	select {
	case <-out:
	case <-time.After(time.Second):
		t.Fatal("stage did not close output after context cancel")
	}
}
