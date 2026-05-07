package slicer

import (
	"context"
	"strings"
	"testing"
)

func TestAnnotator_NoFuncs(t *testing.T) {
	a := NewAnnotator(" ")
	got := a.Apply("hello", 1)
	if got != "hello" {
		t.Fatalf("expected unchanged line, got %q", got)
	}
}

func TestAnnotator_LineNumber(t *testing.T) {
	a := NewAnnotator(" ", LineNumberAnnotation())
	got := a.Apply("world", 3)
	if got != "[3] world" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestAnnotator_PrefixAnnotation(t *testing.T) {
	a := NewAnnotator("|", PrefixAnnotation("INFO"))
	got := a.Apply("msg", 1)
	if got != "INFO|msg" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestAnnotator_MultipleAnnotations(t *testing.T) {
	a := NewAnnotator(" ", PrefixAnnotation("APP"), LineNumberAnnotation())
	got := a.Apply("line", 7)
	if got != "APP [7] line" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestTimestampAnnotation_NonEmpty(t *testing.T) {
	a := NewAnnotator(" ", TimestampAnnotation(""))
	got := a.Apply("data", 1)
	// timestamp prefix should be non-empty and line appended
	if !strings.HasSuffix(got, " data") {
		t.Fatalf("expected suffix ' data', got %q", got)
	}
}

func TestAnnotateStage_AnnotatesLines(t *testing.T) {
	a := NewAnnotator(" ", LineNumberAnnotation())
	stage := NewAnnotateStage(a)

	in := make(chan string, 3)
	in <- "alpha"
	in <- "beta"
	in <- "gamma"
	close(in)

	out := stage.Run(context.Background(), in)
	var results []string
	for line := range out {
		results = append(results, line)
	}

	expected := []string{"[1] alpha", "[2] beta", "[3] gamma"}
	if len(results) != len(expected) {
		t.Fatalf("expected %d lines, got %d", len(expected), len(results))
	}
	for i, want := range expected {
		if results[i] != want {
			t.Errorf("line %d: want %q, got %q", i, want, results[i])
		}
	}
}

func TestAnnotateStage_CancelViaContext(t *testing.T) {
	a := NewAnnotator(" ", LineNumberAnnotation())
	stage := NewAnnotateStage(a)

	in := make(chan string) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := stage.Run(ctx, in)
	cancel()

	// channel should close after cancellation
	for range out {
	}
}
