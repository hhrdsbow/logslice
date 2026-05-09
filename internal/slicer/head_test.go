package slicer

import (
	"context"
	"testing"
	"time"
)

func feedHead(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectHead(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestHeadReader_LimitLines(t *testing.T) {
	h := NewHeadReader(3)
	in := feedHead([]string{"a", "b", "c", "d", "e"})
	out := collectHead(h.Read(context.Background(), in))
	if len(out) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(out))
	}
	if out[2] != "c" {
		t.Errorf("expected third line 'c', got %q", out[2])
	}
}

func TestHeadReader_FewerThanMax(t *testing.T) {
	h := NewHeadReader(10)
	in := feedHead([]string{"x", "y"})
	out := collectHead(h.Read(context.Background(), in))
	if len(out) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(out))
	}
}

func TestHeadReader_ZeroClampsToOne(t *testing.T) {
	h := NewHeadReader(0)
	in := feedHead([]string{"only", "ignored"})
	out := collectHead(h.Read(context.Background(), in))
	if len(out) != 1 {
		t.Fatalf("expected 1 line, got %d", len(out))
	}
}

func TestHeadReader_EmptyInput(t *testing.T) {
	h := NewHeadReader(5)
	in := feedHead(nil)
	out := collectHead(h.Read(context.Background(), in))
	if len(out) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(out))
	}
}

func TestHeadReader_CancelViaContext(t *testing.T) {
	h := NewHeadReader(100)
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan string)
	out := h.Read(ctx, ch)
	cancel()
	select {
	case <-out:
	case <-time.After(time.Second):
		t.Fatal("output channel not closed after context cancel")
	}
}

func TestHeadReader_Summary(t *testing.T) {
	h := NewHeadReader(7)
	s := h.Summary()
	if s != "head: first 7 line(s)" {
		t.Errorf("unexpected summary: %q", s)
	}
}
