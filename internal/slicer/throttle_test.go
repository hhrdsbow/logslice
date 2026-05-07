package slicer

import (
	"context"
	"testing"
	"time"
)

func TestThrottle_AllowsUpToMax(t *testing.T) {
	th, err := NewThrottle(ThrottleConfig{MaxLines: 3, Interval: time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 3; i++ {
		if !th.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
	if th.Allow() {
		t.Fatal("expected Allow()=false after exceeding MaxLines")
	}
}

func TestThrottle_Reset(t *testing.T) {
	th, err := NewThrottle(ThrottleConfig{MaxLines: 1, Interval: time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !th.Allow() {
		t.Fatal("first Allow should pass")
	}
	if th.Allow() {
		t.Fatal("second Allow should be throttled")
	}
	th.Reset()
	if !th.Allow() {
		t.Fatal("Allow should pass after Reset")
	}
}

func TestThrottle_DefaultsClampInvalidConfig(t *testing.T) {
	th, err := NewThrottle(ThrottleConfig{MaxLines: 0, Interval: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th.cfg.MaxLines != 1 {
		t.Errorf("expected MaxLines=1, got %d", th.cfg.MaxLines)
	}
	if th.cfg.Interval != time.Second {
		t.Errorf("expected Interval=1s, got %v", th.cfg.Interval)
	}
}

func TestThrottleStage_DropsExcessLines(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 10)
	lines := []string{"a", "b", "c", "d", "e"}
	for _, l := range lines {
		in <- l
	}
	close(in)

	out, err := ThrottleStage(ctx, in, ThrottleConfig{MaxLines: 2, Interval: time.Second})
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

func TestThrottleStage_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan string)
	out, err := ThrottleStage(ctx, in, ThrottleConfig{MaxLines: 100, Interval: time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	// drain; channel must close
	for range out {
	}
}

func TestThrottleStage_PassesAllWhenUnderLimit(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 5)
	for i := 0; i < 5; i++ {
		in <- "line"
	}
	close(in)

	out, err := ThrottleStage(ctx, in, ThrottleConfig{MaxLines: 100, Interval: time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	for range out {
		count++
	}
	if count != 5 {
		t.Errorf("expected 5 lines, got %d", count)
	}
}
