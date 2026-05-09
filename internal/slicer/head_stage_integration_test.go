package slicer

import (
	"context"
	"strings"
	"testing"
)

// TestHeadStage_IntegrationWithLineReader verifies that NewHeadStage works
// correctly when chained after a LineReader, simulating real pipeline usage.
func TestHeadStage_IntegrationWithLineReader(t *testing.T) {
	input := strings.NewReader("alpha\nbeta\ngamma\ndelta\nepsilon\n")
	reader := NewLineReader(input)

	ctx := context.Background()
	lines := reader.Read(ctx)

	headStage := NewHeadStage(WithHeadMax(3))
	out := headStage(ctx, lines)

	var collected []string
	for l := range out {
		collected = append(collected, l)
	}

	if len(collected) != 3 {
		t.Fatalf("expected 3 lines from head stage, got %d", len(collected))
	}
	expected := []string{"alpha", "beta", "gamma"}
	for i, want := range expected {
		if collected[i] != want {
			t.Errorf("line %d: expected %q, got %q", i, want, collected[i])
		}
	}
}

// TestHeadStage_IntegrationExactMatch verifies that when input equals max,
// all lines are forwarded without blocking.
func TestHeadStage_IntegrationExactMatch(t *testing.T) {
	input := strings.NewReader("one\ntwo\nthree\n")
	reader := NewLineReader(input)

	ctx := context.Background()
	headStage := NewHeadStage(WithHeadMax(3))
	out := headStage(ctx, reader.Read(ctx))

	var collected []string
	for l := range out {
		collected = append(collected, l)
	}

	if len(collected) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(collected))
	}
}
