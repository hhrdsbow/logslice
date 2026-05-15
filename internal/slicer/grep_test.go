package slicer

import (
	"testing"
)

func TestNewGrepFilter_EmptyPattern(t *testing.T) {
	_, err := NewGrepFilter("")
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewGrepFilter_InvalidRegex(t *testing.T) {
	_, err := NewGrepFilter("[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestGrepFilter_Match(t *testing.T) {
	f, err := NewGrepFilter(`error`)
	if err != nil {
		t.Fatal(err)
	}
	line, keep := f.Apply("an error occurred")
	if !keep {
		t.Fatal("expected line to be kept")
	}
	if line != "an error occurred" {
		t.Fatalf("unexpected line: %q", line)
	}
}

func TestGrepFilter_NoMatch(t *testing.T) {
	f, _ := NewGrepFilter(`error`)
	_, keep := f.Apply("everything is fine")
	if keep {
		t.Fatal("expected line to be dropped")
	}
}

func TestGrepFilter_Invert(t *testing.T) {
	f, _ := NewGrepFilter(`error`, WithGrepInvert())
	_, keep := f.Apply("an error occurred")
	if keep {
		t.Fatal("inverted: matching line should be dropped")
	}
	_, keep = f.Apply("all good")
	if !keep {
		t.Fatal("inverted: non-matching line should be kept")
	}
}

func TestGrepFilter_CaptureGroup(t *testing.T) {
	f, err := NewGrepFilter(`level=(\w+)`, WithGrepGroup(1))
	if err != nil {
		t.Fatal(err)
	}
	result, keep := f.Apply("2024-01-01 level=warn msg=disk")
	if !keep {
		t.Fatal("expected line to be kept")
	}
	if result != "warn" {
		t.Fatalf("expected capture group text %q, got %q", "warn", result)
	}
}

func TestGrepFilter_CaptureGroup_OutOfRange(t *testing.T) {
	// group index beyond available groups falls back to whole match
	f, _ := NewGrepFilter(`error`, WithGrepGroup(5))
	result, keep := f.Apply("fatal error here")
	if !keep {
		t.Fatal("expected keep")
	}
	if result != "fatal error here" {
		t.Fatalf("unexpected result: %q", result)
	}
}
