package slicer

import (
	"context"
	"testing"
	"time"
)

func feedSkip(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectSkip(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestSkipReader_SkipsFirstN(t *testing.T) {
	input := feedSkip([]string{"a", "b", "c", "d", "e"})
	sr := NewSkipReader(input, 2)
	got := collectSkip(sr.Lines(context.Background()))
	want := []string{"c", "d", "e"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("index %d: want %q, got %q", i, v, got[i])
		}
	}
}

func TestSkipReader_ZeroSkip(t *testing.T) {
	input := feedSkip([]string{"x", "y"})
	sr := NewSkipReader(input, 0)
	got := collectSkip(sr.Lines(context.Background()))
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestSkipReader_NegativeClampedToZero(t *testing.T) {
	input := feedSkip([]string{"p", "q"})
	sr := NewSkipReader(input, -5)
	got := collectSkip(sr.Lines(context.Background()))
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestSkipReader_SkipMoreThanAvailable(t *testing.T) {
	input := feedSkip([]string{"only"})
	sr := NewSkipReader(input, 10)
	got := collectSkip(sr.Lines(context.Background()))
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %v", got)
	}
}

func TestSkipReader_CancelViaContext(t *testing.T) {
	ch := make(chan string)
	sr := NewSkipReader(ch, 0)
	ctx, cancel := context.WithCancel(context.Background())
	out := sr.Lines(ctx)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for channel close")
	}
}
