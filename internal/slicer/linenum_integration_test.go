package slicer

import (
	"context"
	"strings"
	"testing"
)

func TestLineNumStage_IntegrationWithLineReader(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5\n"
	ctx := context.Background()

	lines, err := NewLineReader(ctx, strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	out, err := NewLineNumStage(lines, WithLineStart(2), WithLineEnd(4))
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for l := range out {
		got = append(got, l)
	}

	expected := []string{"line2", "line3", "line4"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d lines, got %d: %v", len(expected), len(got), got)
	}
	for i, e := range expected {
		if got[i] != e {
			t.Errorf("line %d: expected %q, got %q", i, e, got[i])
		}
	}
}

func TestLineNumStage_IntegrationCtxCancel(t *testing.T) {
	input := "a\nb\nc\nd\ne\nf\ng\n"
	ctx, cancel := context.WithCancel(context.Background())

	lines, err := NewLineReader(ctx, strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	out, err := NewLineNumStageCtx(ctx, lines, WithLineStart(1), WithLineEnd(0))
	if err != nil {
		t.Fatal(err)
	}

	// read two lines then cancel
	<-out
	<-out
	cancel()

	// drain remaining
	for range out {
	}
	// no deadlock = pass
}

func TestLineNumStage_IntegrationSingleLine(t *testing.T) {
	input := "only\n"
	ctx := context.Background()

	lines, err := NewLineReader(ctx, strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	out, err := NewLineNumStage(lines, WithLineStart(1), WithLineEnd(1))
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for l := range out {
		got = append(got, l)
	}

	if len(got) != 1 || got[0] != "only" {
		t.Fatalf("expected [only], got %v", got)
	}
}
