package slicer

import (
	"context"
	"strings"
	"testing"
)

func TestSplitter_NoLimits(t *testing.T) {
	input := "line1\nline2\nline3\n"
	sp := NewSplitter(strings.NewReader(input), SplitConfig{})
	segs := collectSegments(t, sp)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if segs[0].Len() != 3 {
		t.Errorf("expected 3 lines, got %d", segs[0].Len())
	}
}

func TestSplitter_MaxLines(t *testing.T) {
	input := "a\nb\nc\nd\ne\n"
	sp := NewSplitter(strings.NewReader(input), SplitConfig{MaxLines: 2})
	segs := collectSegments(t, sp)
	if len(segs) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segs))
	}
	if segs[0].Len() != 2 || segs[1].Len() != 2 || segs[2].Len() != 1 {
		t.Errorf("unexpected segment sizes: %d %d %d",
			segs[0].Len(), segs[1].Len(), segs[2].Len())
	}
}

func TestSplitter_MaxBytes(t *testing.T) {
	// Each line is 5 bytes + newline = 6; limit to 12 bytes => 2 lines per segment
	input := "aaaaa\nbbbbb\nccccc\nddddd\n"
	sp := NewSplitter(strings.NewReader(input), SplitConfig{MaxBytes: 12})
	segs := collectSegments(t, sp)
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segs))
	}
}

func TestSplitter_EmptyInput(t *testing.T) {
	sp := NewSplitter(strings.NewReader(""), SplitConfig{MaxLines: 10})
	segs := collectSegments(t, sp)
	if len(segs) != 0 {
		t.Fatalf("expected 0 segments for empty input, got %d", len(segs))
	}
}

func TestSplitter_SegmentNaming(t *testing.T) {
	input := "x\ny\nz\n"
	sp := NewSplitter(strings.NewReader(input), SplitConfig{MaxLines: 1})
	segs := collectSegments(t, sp)
	expected := []string{"segment-000", "segment-001", "segment-002"}
	for i, seg := range segs {
		if seg.Name != expected[i] {
			t.Errorf("segment %d: expected name %q, got %q", i, expected[i], seg.Name)
		}
	}
}

func TestSplitter_CancelContext(t *testing.T) {
	input := strings.Repeat("line\n", 1000)
	ctx, cancel := context.WithCancel(context.Background())
	sp := NewSplitter(strings.NewReader(input), SplitConfig{MaxLines: 1})
	ch, _ := sp.Split(ctx)
	// Read one segment then cancel
	<-ch
	cancel()
	// Drain remaining; should not block
	for range ch {
	}
}

// collectSegments is a test helper that runs Split and returns all segments.
func collectSegments(t *testing.T, sp *Splitter) []*Segment {
	t.Helper()
	ch, errs := sp.Split(context.Background())
	var segs []*Segment
	for seg := range ch {
		segs = append(segs, seg)
	}
	if err := <-errs; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return segs
}
