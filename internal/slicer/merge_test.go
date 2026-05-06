package slicer

import (
	"context"
	"sort"
	"testing"
	"time"
)

func makeChan(lines ...string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectMerged(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestMergeReader_SingleInput(t *testing.T) {
	r := NewMergeReader(makeChan("a", "b", "c"))
	got := collectMerged(r.Read(context.Background()))
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestMergeReader_MultipleInputs(t *testing.T) {
	r := NewMergeReader(
		makeChan("a", "b"),
		makeChan("c", "d"),
		makeChan("e"),
	)
	got := collectMerged(r.Read(context.Background()))
	sort.Strings(got)
	want := []string{"a", "b", "c", "d", "e"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestMergeReader_EmptyInputs(t *testing.T) {
	r := NewMergeReader(makeChan(), makeChan())
	got := collectMerged(r.Read(context.Background()))
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}

func TestMergeReader_NoInputs(t *testing.T) {
	r := NewMergeReader()
	got := collectMerged(r.Read(context.Background()))
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}

func TestMergeReader_CancelViaContext(t *testing.T) {
	slow := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())

	r := NewMergeReader(slow)
	out := r.Read(ctx)

	cancel()

	select {
	case <-out:
		// channel closed after cancel — ok
	case <-time.After(time.Second):
		t.Fatal("output channel not closed after context cancel")
	}
}
