package slicer

import (
	"testing"
)

func TestLineNumFilter_AcceptsRange(t *testing.T) {
	f := NewLineNumFilter(2, 4)
	lines := []string{"a", "b", "c", "d", "e"}
	var got []string
	for _, l := range lines {
		if f.Accept(l) {
			got = append(got, l)
		}
	}
	if len(got) != 3 || got[0] != "b" || got[2] != "d" {
		t.Fatalf("expected [b c d], got %v", got)
	}
}

func TestLineNumFilter_NoUpperBound(t *testing.T) {
	f := NewLineNumFilter(3, 0)
	lines := []string{"1", "2", "3", "4", "5"}
	var got []string
	for _, l := range lines {
		if f.Accept(l) {
			got = append(got, l)
		}
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestLineNumFilter_StartClampedToOne(t *testing.T) {
	f := NewLineNumFilter(-5, 2)
	if f.start != 1 {
		t.Fatalf("expected start=1, got %d", f.start)
	}
}

func TestLineNumFilter_Reset(t *testing.T) {
	f := NewLineNumFilter(1, 3)
	f.Accept("x")
	f.Accept("y")
	f.Reset()
	if f.Current() != 0 {
		t.Fatalf("expected counter 0 after reset, got %d", f.Current())
	}
}

func TestLineNumFilter_Summary(t *testing.T) {
	f1 := NewLineNumFilter(5, 10)
	if s := f1.Summary(); s != "lines 5-10" {
		t.Fatalf("unexpected summary: %s", s)
	}
	f2 := NewLineNumFilter(3, 0)
	if s := f2.Summary(); s != "lines 3+" {
		t.Fatalf("unexpected summary: %s", s)
	}
}

func sendLineNumLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectLineNumLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewLineNumStage_DefaultPassAll(t *testing.T) {
	in := sendLineNumLines([]string{"a", "b", "c"})
	out, err := NewLineNumStage(in)
	if err != nil {
		t.Fatal(err)
	}
	got := collectLineNumLines(out)
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestNewLineNumStage_WithRange(t *testing.T) {
	in := sendLineNumLines([]string{"a", "b", "c", "d", "e"})
	out, err := NewLineNumStage(in, WithLineStart(2), WithLineEnd(4))
	if err != nil {
		t.Fatal(err)
	}
	got := collectLineNumLines(out)
	if len(got) != 3 || got[0] != "b" || got[2] != "d" {
		t.Fatalf("expected [b c d], got %v", got)
	}
}

func TestNewLineNumStage_EmptyInput(t *testing.T) {
	in := sendLineNumLines(nil)
	out, err := NewLineNumStage(in, WithLineStart(1), WithLineEnd(5))
	if err != nil {
		t.Fatal(err)
	}
	got := collectLineNumLines(out)
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}
