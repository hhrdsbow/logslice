package slicer

import (
	"context"
	"strings"
	"testing"
	"time"
)

func sendContextLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectContextLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func matchWord(word string) func(string) bool {
	return func(line string) bool {
		return strings.Contains(line, word)
	}
}

func TestContextFilter_MatchOnly(t *testing.T) {
	cf := NewContextFilter(matchWord("ERROR"), 0, 0)
	in := sendContextLines([]string{"info a", "ERROR b", "info c"})
	out := make(chan string, 10)
	cf.Run(context.Background(), in, out)
	got := collectContextLines(out)
	if len(got) != 1 || got[0] != "ERROR b" {
		t.Fatalf("expected [ERROR b], got %v", got)
	}
}

func TestContextFilter_BeforeContext(t *testing.T) {
	cf := NewContextFilter(matchWord("ERROR"), 2, 0)
	in := sendContextLines([]string{"a", "b", "c", "ERROR d", "e"})
	out := make(chan string, 10)
	cf.Run(context.Background(), in, out)
	got := collectContextLines(out)
	want := []string{"b", "c", "ERROR d"}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] want %q got %q", i, want[i], got[i])
		}
	}
}

func TestContextFilter_AfterContext(t *testing.T) {
	cf := NewContextFilter(matchWord("ERROR"), 0, 2)
	in := sendContextLines([]string{"a", "ERROR b", "c", "d", "e"})
	out := make(chan string, 10)
	cf.Run(context.Background(), in, out)
	got := collectContextLines(out)
	want := []string{"ERROR b", "c", "d"}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestContextFilter_NegativeClampedToZero(t *testing.T) {
	cf := NewContextFilter(matchWord("X"), -5, -3)
	if cf.before != 0 || cf.after != 0 {
		t.Fatalf("expected before=0 after=0, got %d %d", cf.before, cf.after)
	}
}

func TestContextFilter_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cf := NewContextFilter(matchWord("X"), 1, 1)
	in := make(chan string)
	out := make(chan string, 10)
	done := make(chan struct{})
	go func() {
		cf.Run(ctx, in, out)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not respect context cancellation")
	}
}
