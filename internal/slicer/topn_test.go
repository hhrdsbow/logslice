package slicer

import (
	"bytes"
	"strings"
	"testing"
)

func TestTopN_KeepsHighestScores(t *testing.T) {
	scorer := func(line string) float64 { return float64(len(line)) }
	top := NewTopN(3, scorer)

	for _, l := range []string{"ab", "abcde", "x", "abcd", "abc"} {
		top.Add(l)
	}

	snap := top.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(snap))
	}
	if snap[0].Line != "abcde" {
		t.Errorf("expected top entry 'abcde', got %q", snap[0].Line)
	}
	if snap[1].Line != "abcd" {
		t.Errorf("expected second entry 'abcd', got %q", snap[1].Line)
	}
}

func TestTopN_NClampedToOne(t *testing.T) {
	top := NewTopN(0, nil)
	top.Add("hello")
	top.Add("world")
	snap := top.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap))
	}
}

func TestTopN_NilScorerDefaultsToLength(t *testing.T) {
	top := NewTopN(2, nil)
	top.Add("short")
	top.Add("a longer line")
	snap := top.Snapshot()
	if snap[0].Line != "a longer line" {
		t.Errorf("expected 'a longer line' at top, got %q", snap[0].Line)
	}
}

func TestTopN_Reset(t *testing.T) {
	top := NewTopN(5, nil)
	top.Add("line one")
	top.Add("line two")
	top.Reset()
	if snap := top.Snapshot(); len(snap) != 0 {
		t.Errorf("expected empty snapshot after reset, got %d entries", len(snap))
	}
}

func TestTopN_WriteSummary(t *testing.T) {
	scorer := func(line string) float64 { return float64(len(line)) }
	top := NewTopN(2, scorer)
	top.Add("hello")
	top.Add("hi")

	var buf bytes.Buffer
	top.WriteSummary(&buf)
	out := buf.String()
	if !strings.Contains(out, "#1") || !strings.Contains(out, "hello") {
		t.Errorf("unexpected summary output: %q", out)
	}
}

func TestTopNStage_PassesAllLines(t *testing.T) {
	in := make(chan string, 4)
	lines := []string{"alpha", "beta", "gamma", "delta"}
	for _, l := range lines {
		in <- l
	}
	close(in)

	top := NewTopN(2, nil)
	out := NewTopNStage(in, top)

	var collected []string
	for l := range out {
		collected = append(collected, l)
	}
	if len(collected) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(collected))
	}
	if len(top.Snapshot()) != 2 {
		t.Errorf("expected top 2 collected")
	}
}

func TestTopN_ConcurrentSafe(t *testing.T) {
	top := NewTopN(5, nil)
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 20; j++ {
				top.Add(strings.Repeat("x", i*j+1))
			}
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if len(top.Snapshot()) > 5 {
		t.Error("snapshot exceeded N entries")
	}
}
