package slicer

import (
	"context"
	"strings"
	"testing"
)

func TestAggregator_FullBucket(t *testing.T) {
	agg := NewAggregator(3, ConcatAggregate)
	if _, ok := agg.Add("a"); ok {
		t.Fatal("expected no output before bucket full")
	}
	if _, ok := agg.Add("b"); ok {
		t.Fatal("expected no output before bucket full")
	}
	summary, ok := agg.Add("c")
	if !ok {
		t.Fatal("expected output when bucket full")
	}
	if summary != "a | b | c" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}

func TestAggregator_Flush(t *testing.T) {
	agg := NewAggregator(5, CountAggregate)
	agg.Add("x")
	agg.Add("y")
	summary, ok := agg.Flush()
	if !ok {
		t.Fatal("expected flush to return a summary")
	}
	if summary != "[2 lines]" {
		t.Fatalf("unexpected summary: %q", summary)
	}
	_, ok = agg.Flush()
	if ok {
		t.Fatal("second flush of empty bucket should return false")
	}
}

func TestAggregator_NilFnDefaultsToConcat(t *testing.T) {
	agg := NewAggregator(2, nil)
	agg.Add("hello")
	summary, ok := agg.Add("world")
	if !ok || summary != "hello | world" {
		t.Fatalf("unexpected: ok=%v summary=%q", ok, summary)
	}
}

func TestAggregator_ZeroSizeClamped(t *testing.T) {
	agg := NewAggregator(0, FirstAggregate)
	summary, ok := agg.Add("only")
	if !ok || summary != "only" {
		t.Fatalf("unexpected: ok=%v summary=%q", ok, summary)
	}
}

func collectAggLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestAggregateStage_EmitsSummaries(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 6)
	for _, l := range []string{"a", "b", "c", "d", "e", "f"} {
		in <- l
	}
	close(in)

	out := NewAggregateStage(ctx, in, WithBucketSize(2), WithAggregateFunc(ConcatAggregate))
	lines := collectAggLines(out)
	if len(lines) != 3 {
		t.Fatalf("expected 3 summaries, got %d: %v", len(lines), lines)
	}
	for _, l := range lines {
		if !strings.Contains(l, " | ") {
			t.Fatalf("unexpected summary format: %q", l)
		}
	}
}

func TestAggregateStage_FlushesPartialBucket(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 3)
	for _, l := range []string{"x", "y", "z"} {
		in <- l
	}
	close(in)

	out := NewAggregateStage(ctx, in, WithBucketSize(10), WithAggregateFunc(CountAggregate))
	lines := collectAggLines(out)
	if len(lines) != 1 || lines[0] != "[3 lines]" {
		t.Fatalf("unexpected output: %v", lines)
	}
}
