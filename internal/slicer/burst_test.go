package slicer

import (
	"context"
	"testing"
	"time"
)

func TestBurstDetector_NoBurst(t *testing.T) {
	b := NewBurstDetector(time.Second, 5)
	now := time.Now()
	for i := 0; i < 5; i++ {
		if b.Record(now) {
			t.Fatalf("expected no burst on line %d", i+1)
		}
	}
}

func TestBurstDetector_DetectsBurst(t *testing.T) {
	b := NewBurstDetector(time.Second, 3)
	now := time.Now()
	for i := 0; i < 3; i++ {
		b.Record(now)
	}
	if !b.Record(now) {
		t.Fatal("expected burst to be detected on 4th line")
	}
}

func TestBurstDetector_EvictsOldTimestamps(t *testing.T) {
	b := NewBurstDetector(100*time.Millisecond, 2)
	old := time.Now().Add(-200 * time.Millisecond)
	b.Record(old)
	b.Record(old)
	// old timestamps should be evicted; new ones start fresh
	if b.Record(time.Now()) {
		t.Fatal("expected no burst after eviction of old timestamps")
	}
}

func TestBurstDetector_Reset(t *testing.T) {
	b := NewBurstDetector(time.Second, 2)
	now := time.Now()
	b.Record(now)
	b.Record(now)
	b.Record(now)
	b.Reset()
	if b.Record(now) {
		t.Fatal("expected no burst after reset")
	}
}

func TestBurstDetector_ZeroThresholdClamped(t *testing.T) {
	b := NewBurstDetector(time.Second, 0)
	if !b.Record(time.Now()) {
		t.Fatal("threshold clamped to 1; first line should trigger burst")
	}
}

func TestBurstStage_TagsBurstLines(t *testing.T) {
	detector := NewBurstDetector(time.Second, 2)
	stage := NewBurstStage(detector, "[BURST] ")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan string, 5)
	lines := []string{"a", "b", "c", "d"}
	for _, l := range lines {
		in <- l
	}
	close(in)

	out := stage.Run(ctx, in)
	var got []string
	for l := range out {
		got = append(got, l)
	}

	if len(got) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(got))
	}
	// First two lines should not be burst-tagged; lines 3+ should be
	for i, l := range got {
		isBurst := len(l) >= len("[BURST] ") && l[:len("[BURST] ")] == "[BURST] "
		if i < 2 && isBurst {
			t.Errorf("line %d should not be tagged as burst", i)
		}
		if i >= 2 && !isBurst {
			t.Errorf("line %d should be tagged as burst", i)
		}
	}
}

func TestBurstStage_DefaultPrefix(t *testing.T) {
	detector := NewBurstDetector(time.Second, 0) // threshold=1 after clamp
	stage := NewBurstStage(detector, "")

	ctx := context.Background()
	in := make(chan string, 1)
	in <- "hello"
	close(in)

	out := stage.Run(ctx, in)
	l := <-out
	if l != "[BURST] hello" {
		t.Fatalf("expected default burst prefix, got %q", l)
	}
}
