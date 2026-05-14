package slicer

import (
	"math"
	"testing"
)

func TestShannonEntropy_Empty(t *testing.T) {
	if got := shannonEntropy(""); got != 0 {
		t.Fatalf("expected 0 for empty string, got %f", got)
	}
}

func TestShannonEntropy_SingleChar(t *testing.T) {
	// Single repeated character has entropy 0.
	if got := shannonEntropy("aaaa"); got != 0 {
		t.Fatalf("expected 0 for uniform string, got %f", got)
	}
}

func TestShannonEntropy_TwoEqualChars(t *testing.T) {
	// "ab" → entropy = 1.0 bit
	got := shannonEntropy("ab")
	if math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("expected 1.0, got %f", got)
	}
}

func TestEntropyFilter_DefaultAcceptsAll(t *testing.T) {
	f := NewEntropyFilter()
	lines := []string{"hello world", "aaaaaa", "abc123!@#XYZ"}
	for _, l := range lines {
		if !f.Accept(l) {
			t.Errorf("default filter should accept %q", l)
		}
	}
}

func TestEntropyFilter_MinEntropy(t *testing.T) {
	// Require entropy >= 2.0; "aaaa" (entropy=0) should be rejected.
	f := NewEntropyFilter(WithMinEntropy(2.0))
	if f.Accept("aaaa") {
		t.Error("low-entropy line should be rejected")
	}
	if !f.Accept("hello world") {
		t.Error("higher-entropy line should be accepted")
	}
}

func TestEntropyFilter_MaxEntropy(t *testing.T) {
	// Reject high-entropy lines (e.g. random-looking base64).
	f := NewEntropyFilter(WithMaxEntropy(2.0))
	if f.Accept("aAbBcCdDeEfFgG") {
		t.Error("high-entropy line should be rejected")
	}
	if !f.Accept("aaab") {
		t.Error("low-entropy line should be accepted")
	}
}

func TestEntropyFilter_Invert(t *testing.T) {
	// Inverted: keep lines OUTSIDE [2, 8] — i.e. very low entropy.
	f := NewEntropyFilter(WithMinEntropy(2.0), WithEntropyInvert())
	if !f.Accept("aaaa") {
		t.Error("low-entropy line should pass inverted filter")
	}
	if f.Accept("hello world 123") {
		t.Error("higher-entropy line should be rejected by inverted filter")
	}
}

func TestEntropyFilter_EntropyMethod(t *testing.T) {
	f := NewEntropyFilter()
	if got := f.Entropy(""); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
}

func TestNewEntropyStage_FiltersLines(t *testing.T) {
	in := make(chan string, 4)
	in <- "aaaa"        // entropy ≈ 0
	in <- "hello world" // entropy > 2
	in <- "bbbb"        // entropy = 0
	in <- "abcdefgh"    // entropy = 3
	close(in)

	out := NewEntropyStage(in, WithMinEntropy(2.0))

	var got []string
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
}

func TestNewEntropyStage_EmptyInput(t *testing.T) {
	in := make(chan string)
	close(in)
	out := NewEntropyStage(in)
	var count int
	for range out {
		count++
	}
	if count != 0 {
		t.Fatalf("expected 0 lines, got %d", count)
	}
}
