package slicer

import (
	"context"
	"testing"
)

func TestCountBy_NilExtractorUsesWholeLine(t *testing.T) {
	cb := NewCountBy(nil)
	cb.Add("foo")
	cb.Add("bar")
	cb.Add("foo")
	res := cb.Results()
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	if res[0].Key != "foo" || res[0].Count != 2 {
		t.Errorf("expected foo=2, got %+v", res[0])
	}
}

func TestCountBy_SortedDescending(t *testing.T) {
	cb := NewCountBy(nil)
	for _, l := range []string{"a", "b", "b", "c", "c", "c"} {
		cb.Add(l)
	}
	res := cb.Results()
	if res[0].Key != "c" || res[0].Count != 3 {
		t.Errorf("first result should be c=3, got %+v", res[0])
	}
	if res[2].Key != "a" || res[2].Count != 1 {
		t.Errorf("last result should be a=1, got %+v", res[2])
	}
}

func TestCountBy_Reset(t *testing.T) {
	cb := NewCountBy(nil)
	cb.Add("x")
	cb.Reset()
	if len(cb.Results()) != 0 {
		t.Error("expected empty results after reset")
	}
}

func TestNewRegexCountBy_InvalidRegex(t *testing.T) {
	_, err := NewRegexCountBy("[invalid")
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestNewRegexCountBy_CaptureGroup(t *testing.T) {
	cb, err := NewRegexCountBy(`level=(\w+)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, l := range []string{
		"2024-01-01 level=error msg=oops",
		"2024-01-01 level=info  msg=ok",
		"2024-01-01 level=error msg=fail",
		"2024-01-01 no-level-here",
	} {
		cb.Add(l)
	}
	res := cb.Results()
	if len(res) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(res))
	}
	if res[0].Key != "error" || res[0].Count != 2 {
		t.Errorf("expected error=2, got %+v", res[0])
	}
}

func TestNewRegexCountBy_NoCapture_UsesWholeMatch(t *testing.T) {
	cb, err := NewRegexCountBy(`ERROR|WARN`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cb.Add("ERROR something")
	cb.Add("WARN something")
	cb.Add("ERROR again")
	res := cb.Results()
	if res[0].Key != "ERROR" || res[0].Count != 2 {
		t.Errorf("expected ERROR=2, got %+v", res[0])
	}
}

func TestCountByStage_ForwardsLines(t *testing.T) {
	cb := NewCountBy(nil)
	in := make(chan string, 4)
	for _, l := range []string{"a", "b", "a", "c"} {
		in <- l
	}
	close(in)

	out := NewCountByStage(context.Background(), in, cb)
	var collected []string
	for l := range out {
		collected = append(collected, l)
	}
	if len(collected) != 4 {
		t.Fatalf("expected 4 forwarded lines, got %d", len(collected))
	}
	res := cb.Results()
	if res[0].Key != "a" || res[0].Count != 2 {
		t.Errorf("expected a=2, got %+v", res[0])
	}
}

func TestCountByStage_CancelViaContext(t *testing.T) {
	cb := NewCountBy(nil)
	in := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	out := NewCountByStage(ctx, in, cb)
	cancel()
	// drain; channel must close without deadlock
	for range out {
	}
}
