package slicer

import (
	"context"
	"testing"
	"time"
)

func sendRegexLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectSegments(ctx context.Context, ch <-chan []string) [][]string {
	var segs [][]string
	for {
		select {
		case seg, ok := <-ch:
			if !ok {
				return segs
			}
			segs = append(segs, seg)
		case <-ctx.Done():
			return segs
		}
	}
}

func TestNewRegexSplitter_EmptyPattern(t *testing.T) {
	_, err := NewRegexSplitter("")
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewRegexSplitter_InvalidRegex(t *testing.T) {
	_, err := NewRegexSplitter("[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestRegexSplitter_BasicSplit(t *testing.T) {
	rs, err := NewRegexSplitter(`^---`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	input := []string{"a", "b", "---", "c", "d", "---", "e"}
	ctx := context.Background()
	segs := collectSegments(ctx, rs.Split(ctx, sendRegexLines(input)))
	if len(segs) != 3 {
		t.Fatalf("expected 3 segments, got %d: %v", len(segs), segs)
	}
	if segs[0][0] != "a" || segs[1][0] != "---" || segs[2][0] != "---" {
		t.Errorf("unexpected segment content: %v", segs)
	}
}

func TestRegexSplitter_NoMatch(t *testing.T) {
	rs, _ := NewRegexSplitter(`^NEVER`)
	input := []string{"line1", "line2", "line3"}
	ctx := context.Background()
	segs := collectSegments(ctx, rs.Split(ctx, sendRegexLines(input)))
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if len(segs[0]) != 3 {
		t.Errorf("expected 3 lines in single segment, got %d", len(segs[0]))
	}
}

func TestRegexSplitter_InvertBoundary(t *testing.T) {
	rs, _ := NewRegexSplitter(`^data`, WithSplitInvert(true))
	// non-data lines trigger new segment
	input := []string{"header", "data1", "data2", "sep", "data3"}
	ctx := context.Background()
	segs := collectSegments(ctx, rs.Split(ctx, sendRegexLines(input)))
	if len(segs) != 3 {
		t.Fatalf("expected 3 segments, got %d: %v", len(segs), segs)
	}
}

func TestRegexSplitter_MaxSegments(t *testing.T) {
	rs, _ := NewRegexSplitter(`^---`, WithMaxSegments(2))
	input := []string{"a", "---", "b", "---", "c", "---", "d"}
	ctx := context.Background()
	segs := collectSegments(ctx, rs.Split(ctx, sendRegexLines(input)))
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments due to cap, got %d", len(segs))
	}
}

func TestRegexSplitter_CancelViaContext(t *testing.T) {
	rs, _ := NewRegexSplitter(`^---`)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	blocking := make(chan string) // never sends
	segs := collectSegments(ctx, rs.Split(ctx, blocking))
	if len(segs) != 0 {
		t.Errorf("expected no segments on cancel, got %d", len(segs))
	}
}
