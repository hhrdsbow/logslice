package slicer

import (
	"context"
	"testing"
	"time"
)

func sendSeverityStageLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectSeverityStageLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewSeverityPipelineStage_DefaultPassesAll(t *testing.T) {
	input := []string{"DEBUG init", "INFO ready", "WARN low disk", "ERROR crash"}
	in := sendSeverityStageLines(input)
	out := NewSeverityPipelineStage(in)
	got := collectSeverityStageLines(out)
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
}

func TestNewSeverityPipelineStage_FiltersBelow(t *testing.T) {
	input := []string{"DEBUG verbose", "INFO normal", "WARN attention", "ERROR bad", "FATAL boom"}
	in := sendSeverityStageLines(input)
	out := NewSeverityPipelineStage(in, WithStageSeverity(SeverityWarn))
	got := collectSeverityStageLines(out)
	// WARN, ERROR, FATAL should pass
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(got), got)
	}
}

func TestNewSeverityPipelineStage_OnlyErrors(t *testing.T) {
	input := []string{"INFO ok", "ERROR fail", "DEBUG noise", "FATAL gone"}
	in := sendSeverityStageLines(input)
	out := NewSeverityPipelineStage(in, WithStageSeverity(SeverityError))
	got := collectSeverityStageLines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
}

func TestNewSeverityPipelineStage_EmptyInput(t *testing.T) {
	in := sendSeverityStageLines(nil)
	out := NewSeverityPipelineStage(in, WithStageSeverity(SeverityInfo))
	got := collectSeverityStageLines(out)
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}

func TestNewSeverityPipelineStageCtx_CancelViaContext(t *testing.T) {
	lines := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	out := NewSeverityPipelineStageCtx(ctx, lines, SeverityDebug)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stage to close after cancel")
	}
}
