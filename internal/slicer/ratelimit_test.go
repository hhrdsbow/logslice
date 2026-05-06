package slicer

import (
	"context"
	"testing"
	"time"
)

func feedLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestRateLimiter_NoLimit(t *testing.T) {
	rl := NewRateLimiter(0)
	if rl.Rate() != 0 {
		t.Fatalf("expected rate 0, got %d", rl.Rate())
	}
	input := []string{"a", "b", "c"}
	ctx := context.Background()
	out := rl.Apply(ctx, feedLines(input))
	got := collectLines(out)
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
	for i, l := range got {
		if l != input[i] {
			t.Errorf("line %d: want %q got %q", i, input[i], l)
		}
	}
}

func TestRateLimiter_Rate(t *testing.T) {
	rl := NewRateLimiter(500)
	if rl.Rate() != 500 {
		t.Fatalf("expected rate 500, got %d", rl.Rate())
	}
}

func TestRateLimiter_PassesAllLines(t *testing.T) {
	// Use a high rate so the test is fast but still exercises the throttle path.
	rl := NewRateLimiter(10000)
	input := []string{"x", "y", "z", "w"}
	ctx := context.Background()
	out := rl.Apply(ctx, feedLines(input))
	got := collectLines(out)
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
}

func TestRateLimiter_CancelViaContext(t *testing.T) {
	rl := NewRateLimiter(1) // 1 line/sec — very slow
	ctx, cancel := context.WithCancel(context.Background())

	// Infinite source
	infinite := make(chan string)
	go func() {
		for {
			select {
			case infinite <- "line":
			case <-ctx.Done():
				close(infinite)
				return
			}
		}
	}()

	out := rl.Apply(ctx, infinite)

	// Cancel quickly — we should not block forever.
	time.AfterFunc(20*time.Millisecond, cancel)

	start := time.Now()
	collectLines(out)
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("Apply did not respect context cancellation (elapsed %v)", elapsed)
	}
}

func TestRateLimiter_NegativeRate(t *testing.T) {
	rl := NewRateLimiter(-5)
	if rl.interval != 0 {
		t.Fatalf("expected zero interval for negative rate, got %v", rl.interval)
	}
	input := []string{"a", "b"}
	ctx := context.Background()
	out := rl.Apply(ctx, feedLines(input))
	got := collectLines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}
