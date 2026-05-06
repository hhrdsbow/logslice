package slicer

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestTailReader_EmitsLines(t *testing.T) {
	input := "line1\nline2\nline3\n"
	r := strings.NewReader(input)
	tr := NewTailReader(r, TailOptions{MaxRetries: 1})

	ctx := context.Background()
	var got []string

	done := make(chan error, 1)
	go func() { done <- tr.Follow(ctx) }()

	for line := range tr.Lines() {
		got = append(got, line)
	}
	if err := <-done; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	for i, want := range []string{"line1", "line2", "line3"} {
		if got[i] != want {
			t.Errorf("line %d: got %q, want %q", i, got[i], want)
		}
	}
}

func TestTailReader_CancelViaContext(t *testing.T) {
	// Use a pipe so Follow blocks waiting for data after the first read.
	pr, pw := io.Pipe()
	defer pw.Close()

	tr := NewTailReader(pr, TailOptions{PollInterval: 50 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- tr.Follow(ctx) }()

	// Write one line then cancel.
	_, _ = pw.Write([]byte("hello\n"))
	time.Sleep(80 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Follow did not return after context cancellation")
	}
}

func TestTailReader_DefaultPollInterval(t *testing.T) {
	r := strings.NewReader("")
	tr := NewTailReader(r, TailOptions{MaxRetries: 1})
	if tr.opts.PollInterval != 250*time.Millisecond {
		t.Errorf("expected default poll interval 250ms, got %v", tr.opts.PollInterval)
	}
}

func TestTailReader_EmptyInput(t *testing.T) {
	r := strings.NewReader("")
	tr := NewTailReader(r, TailOptions{MaxRetries: 1, PollInterval: 10 * time.Millisecond})

	ctx := context.Background()
	done := make(chan error, 1)
	go func() { done <- tr.Follow(ctx) }()

	var lines []string
	for line := range tr.Lines() {
		lines = append(lines, line)
	}
	if err := <-done; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected no lines, got %v", lines)
	}
}
