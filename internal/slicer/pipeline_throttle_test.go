package slicer

import (
	"context"
	"testing"
	"time"
)

func TestNewThrottleStage_DefaultsPassAll(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 3)
	in <- "x"
	in <- "y"
	in <- "z"
	close(in)

	// default 100 lines/s — all 3 must pass
	out, err := NewThrottleStage(ctx, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []string
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 lines, got %d", len(got))
	}
}

func TestNewThrottleStage_WithOptions(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 10)
	for i := 0; i < 10; i++ {
		in <- "line"
	}
	close(in)

	out, err := NewThrottleStage(ctx, in,
		WithThrottleMaxLines(2),
		WithThrottleInterval(time.Second),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []string
	for l := range out {
		got = append(got, l)
	}
	if len(got) > 2 {
		t.Errorf("expected at most 2 lines, got %d", len(got))
	}
}

func TestNewThrottleStage_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan string)
	out, err := NewThrottleStage(ctx, in, WithThrottleMaxLines(10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	for range out {
	}
}

func TestNewThrottleStage_EmptyInput(t *testing.T) {
	ctx := context.Background()
	in := make(chan string)
	close(in)

	out, err := NewThrottleStage(ctx, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	for range out {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 lines from empty input, got %d", count)
	}
}
