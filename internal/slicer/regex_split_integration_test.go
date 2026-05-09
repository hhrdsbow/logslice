package slicer

import (
	"context"
	"strings"
	"testing"
)

// TestRegexSplitStage_IntegrationWithLineReader verifies that NewLineReader
// feeding into NewRegexSplitStage correctly segments a multi-block log.
func TestRegexSplitStage_IntegrationWithLineReader(t *testing.T) {
	raw := strings.Join([]string{
		"START block1",
		"data line 1",
		"data line 2",
		"START block2",
		"data line 3",
		"START block3",
		"data line 4",
	}, "\n")

	ctx := context.Background()
	lineCh := NewLineReader(ctx, strings.NewReader(raw))

	stageCh, err := NewRegexSplitStage(ctx, lineCh,
		WithSplitPattern(`^START`),
		WithSegmentSeparator("---"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var lines []string
	for l := range stageCh {
		lines = append(lines, l)
	}

	// 7 original lines + 2 separators (between 3 segments) = 9
	const want = 9
	if len(lines) != want {
		t.Errorf("expected %d lines, got %d: %v", want, len(lines), lines)
	}

	// Verify separator positions
	if lines[3] != "---" {
		t.Errorf("expected separator at index 3, got %q", lines[3])
	}
	if lines[6] != "---" {
		t.Errorf("expected separator at index 6, got %q", lines[6])
	}
}

// TestRegexSplitStage_IntegrationMaxSegments ensures the stage honours the
// max-segments cap when wired to a real line source.
func TestRegexSplitStage_IntegrationMaxSegments(t *testing.T) {
	raw := strings.Join([]string{
		"A", "---", "B", "---", "C", "---", "D",
	}, "\n")

	ctx := context.Background()
	lineCh := NewLineReader(ctx, strings.NewReader(raw))

	stageCh, err := NewRegexSplitStage(ctx, lineCh,
		WithSplitPattern(`^---`),
		WithSplitMaxSegments(2),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var lines []string
	for l := range stageCh {
		lines = append(lines, l)
	}

	// Only first 2 segments: ["A"] and ["---", "B"] => 3 lines
	if len(lines) != 3 {
		t.Errorf("expected 3 lines for 2-segment cap, got %d: %v", len(lines), lines)
	}
}
