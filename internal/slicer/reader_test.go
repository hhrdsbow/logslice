package slicer

import (
	"strings"
	"testing"
	"time"
)

func TestLineReader_AllLines(t *testing.T) {
	input := "line1\nline2\nline3\n"
	lr := NewLineReader(strings.NewReader(input), 0)
	done := make(chan struct{})
	defer close(done)

	lines, errCh := lr.Lines(done)

	var got []string
	for l := range lines {
		got = append(got, l)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"line1", "line2", "line3"}
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("line %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestLineReader_EmptyInput(t *testing.T) {
	lr := NewLineReader(strings.NewReader(""), 0)
	done := make(chan struct{})
	defer close(done)

	lines, errCh := lr.Lines(done)

	var got []string
	for l := range lines {
		got = append(got, l)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no lines, got %v", got)
	}
}

func TestLineReader_CancelViaDone(t *testing.T) {
	// Large input so the goroutine would block without cancellation.
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("logline\n")
	}

	lr := NewLineReader(strings.NewReader(sb.String()), 0)
	done := make(chan struct{})

	lines, _ := lr.Lines(done)

	// Read a few lines then cancel.
	<-lines
	<-lines
	close(done)

	// Drain remaining lines; channel must close within a reasonable time.
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	for {
		select {
		case _, ok := <-lines:
			if !ok {
				return // success
			}
		case <-timer.C:
			t.Fatal("lines channel did not close after done was closed")
		}
	}
}
