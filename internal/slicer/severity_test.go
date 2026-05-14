package slicer

import (
	"context"
	"testing"
)

func TestDetectSeverity_Known(t *testing.T) {
	cases := []struct {
		line string
		want SeverityLevel
	}{
		{"2024/01/01 DEBUG starting server", SeverityDebug},
		{"[INFO] connection accepted", SeverityInfo},
		{"WARNING: disk usage high", SeverityWarn},
		{"ERROR failed to open file", SeverityError},
		{"FATAL out of memory", SeverityFatal},
		{"PANIC unexpected nil pointer", SeverityFatal},
	}
	for _, tc := range cases {
		got := DetectSeverity(tc.line)
		if got != tc.want {
			t.Errorf("DetectSeverity(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestDetectSeverity_Unknown(t *testing.T) {
	if got := DetectSeverity("plain log line without level"); got != SeverityUnknown {
		t.Errorf("expected SeverityUnknown, got %v", got)
	}
}

func TestSeverityLevel_String(t *testing.T) {
	if SeverityError.String() != "ERROR" {
		t.Fatalf("unexpected string for SeverityError")
	}
	if SeverityUnknown.String() != "UNKNOWN" {
		t.Fatalf("unexpected string for SeverityUnknown")
	}
}

func TestSeverityFilter_Accept(t *testing.T) {
	f := NewSeverityFilter(SeverityWarn)
	if f.Accept("DEBUG verbose output") {
		t.Error("DEBUG should be rejected when min=WARN")
	}
	if f.Accept("INFO server started") {
		t.Error("INFO should be rejected when min=WARN")
	}
	if !f.Accept("WARN disk nearly full") {
		t.Error("WARN should be accepted when min=WARN")
	}
	if !f.Accept("ERROR connection refused") {
		t.Error("ERROR should be accepted when min=WARN")
	}
	if !f.Accept("FATAL system crash") {
		t.Error("FATAL should be accepted when min=WARN")
	}
}

func sendSeverityLines(lines []string) <-chan string {
	ch := make(chan string, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectSeverityLines(ch <-chan string) []string {
	var out []string
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNewSeverityStage_FiltersCorrectly(t *testing.T) {
	input := []string{
		"DEBUG trace detail",
		"INFO service ready",
		"WARN high latency",
		"ERROR timeout",
	}
	in := sendSeverityLines(input)
	out := NewSeverityStage(context.Background(), in, WithMinSeverity(SeverityWarn))
	got := collectSeverityLines(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
}

func TestNewSeverityStage_CancelViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := sendSeverityLines([]string{"ERROR something bad"})
	out := NewSeverityStage(ctx, in)
	_ = collectSeverityLines(out) // must not block
}

func TestNewSeverityStage_EmptyInput(t *testing.T) {
	in := sendSeverityLines(nil)
	out := NewSeverityStage(context.Background(), in)
	got := collectSeverityLines(out)
	if len(got) != 0 {
		t.Fatalf("expected empty output, got %v", got)
	}
}
