package slicer

import (
	"context"
	"strings"
	"testing"
	"time"
)

// windowParser parses lines prefixed with a RFC3339 timestamp followed by a space.
func windowParser(line string) (time.Time, bool) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func sendLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectWindowLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestWindowStage_PassesAllWithinDuration(t *testing.T) {
	stage := NewWindowStage(10*time.Minute, windowParser)
	base := time.Now().Truncate(time.Second)
	lines := []string{
		base.Format(time.RFC3339) + " first",
		base.Add(3*time.Minute).Format(time.RFC3339) + " second",
	}
	out := stage.Run(context.Background(), sendLines(lines))
	got := collectWindowLines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestWindowStage_NoTimestampPassThrough(t *testing.T) {
	stage := NewWindowStage(time.Minute, windowParser)
	lines := []string{"no timestamp here", "also no timestamp"}
	out := stage.Run(context.Background(), sendLines(lines))
	got := collectWindowLines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 pass-through lines, got %d", len(got))
	}
}

func TestWindowStage_ContextCancel(t *testing.T) {
	stage := NewWindowStage(time.Minute, windowParser)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := make(chan string) // never sends
	out := stage.Run(ctx, in)
	got := collectWindowLines(out)
	if len(got) != 0 {
		t.Fatalf("expected 0 lines after cancel, got %d", len(got))
	}
}

func TestWindowStage_Snapshot(t *testing.T) {
	stage := NewWindowStage(time.Hour, windowParser)
	base := time.Now().Truncate(time.Second)
	lines := []string{
		base.Format(time.RFC3339) + " a",
		base.Add(time.Minute).Format(time.RFC3339) + " b",
	}
	out := stage.Run(context.Background(), sendLines(lines))
	collectWindowLines(out)
	snap := stage.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 in snapshot, got %d", len(snap))
	}
}
