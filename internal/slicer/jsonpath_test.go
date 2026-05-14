package slicer

import (
	"context"
	"testing"
)

func TestNewJSONPathExtractor_EmptyPath(t *testing.T) {
	_, err := NewJSONPathExtractor("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestJSONPathExtractor_TopLevel(t *testing.T) {
	ext, err := NewJSONPathExtractor("level")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ext.Extract(`{"level":"info","msg":"hello"}`)
	if got != "info" {
		t.Errorf("want 'info', got %q", got)
	}
}

func TestJSONPathExtractor_Nested(t *testing.T) {
	ext, err := NewJSONPathExtractor("http.method")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ext.Extract(`{"http":{"method":"GET","status":200}}`)
	if got != "GET" {
		t.Errorf("want 'GET', got %q", got)
	}
}

func TestJSONPathExtractor_MissingKey_Fallback(t *testing.T) {
	ext, err := NewJSONPathExtractor("missing", WithJSONFallback("-"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ext.Extract(`{"level":"warn"}`)
	if got != "-" {
		t.Errorf("want '-', got %q", got)
	}
}

func TestJSONPathExtractor_InvalidJSON_Fallback(t *testing.T) {
	ext, err := NewJSONPathExtractor("level", WithJSONFallback("unknown"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ext.Extract("not json at all")
	if got != "unknown" {
		t.Errorf("want 'unknown', got %q", got)
	}
}

func TestJSONPathExtractor_NumericValue(t *testing.T) {
	ext, err := NewJSONPathExtractor("code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ext.Extract(`{"code":404}`)
	if got != "404" {
		t.Errorf("want '404', got %q", got)
	}
}

func TestNewJSONPathStage_MissingPath(t *testing.T) {
	in := make(chan string)
	close(in)
	_, err := NewJSONPathStage(in)
	if err == nil {
		t.Fatal("expected error when no path configured")
	}
}

func TestNewJSONPathStage_ExtractsValues(t *testing.T) {
	in := make(chan string, 3)
	in <- `{"level":"debug"}`
	in <- `{"level":"error"}`
	in <- `{"level":"info"}`
	close(in)

	out, err := NewJSONPathStage(in, WithJSONPath("level"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"debug", "error", "info"}
	for _, w := range want {
		got, ok := <-out
		if !ok {
			t.Fatalf("channel closed early, want %q", w)
		}
		if got != w {
			t.Errorf("want %q, got %q", w, got)
		}
	}
}

func TestNewJSONPathStageCtx_CancelStopsOutput(t *testing.T) {
	in := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())

	out, err := NewJSONPathStageCtx(ctx, in, WithJSONPath("msg"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cancel()
	for range out {
	}
	// reaching here means the goroutine exited cleanly
}
