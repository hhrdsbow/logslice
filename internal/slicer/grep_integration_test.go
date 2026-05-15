package slicer

import (
	"context"
	"strings"
	"testing"
)

func TestGrepStage_IntegrationWithLineReader(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		"DEBUG starting up",
		"INFO  listening on :8080",
		"ERROR failed to connect",
		"INFO  request received",
		"ERROR timeout waiting for db",
	}, "\n"))

	ctx := context.Background()
	lines := NewLineReader(ctx, input)

	out, err := NewGrepStage(ctx, lines, WithGrepPattern(`^ERROR`))
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for l := range out {
		got = append(got, l)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 error lines, got %d: %v", len(got), got)
	}
	for _, l := range got {
		if !strings.HasPrefix(l, "ERROR") {
			t.Errorf("unexpected line: %q", l)
		}
	}
}

func TestGrepStage_IntegrationCaptureGroup(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		"ts=2024-01-01 level=info msg=started",
		"ts=2024-01-01 level=error msg=crash",
		"ts=2024-01-01 level=warn msg=slow",
	}, "\n"))

	ctx := context.Background()
	lines := NewLineReader(ctx, input)

	out, err := NewGrepStage(ctx, lines,
		WithGrepPattern(`level=(\w+)`),
		WithGrepCaptureGroup(1),
	)
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for l := range out {
		got = append(got, l)
	}

	want := []string{"info", "error", "warn"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("line %d: expected %q, got %q", i, w, got[i])
		}
	}
}
