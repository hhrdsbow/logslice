package slicer

import (
	"strings"
	"testing"
)

func TestNewTruncator_InvalidMaxLen(t *testing.T) {
	_, err := NewTruncator(0, TruncateEnd)
	if err == nil {
		t.Fatal("expected error for maxLen=0")
	}
}

func TestNewTruncator_MiddleTooShort(t *testing.T) {
	_, err := NewTruncator(4, TruncateMiddle)
	if err == nil {
		t.Fatal("expected error for TruncateMiddle with maxLen=4")
	}
}

func TestTruncator_ShortLineUnchanged(t *testing.T) {
	tr, _ := NewTruncator(20, TruncateEnd)
	line := "short line"
	if got := tr.Apply(line); got != line {
		t.Errorf("expected %q, got %q", line, got)
	}
}

func TestTruncator_ExactLengthUnchanged(t *testing.T) {
	tr, _ := NewTruncator(10, TruncateEnd)
	line := "1234567890"
	if got := tr.Apply(line); got != line {
		t.Errorf("expected %q, got %q", line, got)
	}
}

func TestTruncator_TruncateEnd(t *testing.T) {
	tr, _ := NewTruncator(10, TruncateEnd)
	line := "abcdefghijklmnop"
	got := tr.Apply(line)
	if len([]rune(got)) != 10 {
		t.Errorf("expected length 10, got %d: %q", len([]rune(got)), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
}

func TestTruncator_TruncateStart(t *testing.T) {
	tr, _ := NewTruncator(10, TruncateStart)
	line := "abcdefghijklmnop"
	got := tr.Apply(line)
	if len([]rune(got)) != 10 {
		t.Errorf("expected length 10, got %d: %q", len([]rune(got)), got)
	}
	if !strings.HasPrefix(got, "...") {
		t.Errorf("expected ellipsis prefix, got %q", got)
	}
}

func TestTruncator_TruncateMiddle(t *testing.T) {
	tr, _ := NewTruncator(11, TruncateMiddle)
	line := "abcdefghijklmnop"
	got := tr.Apply(line)
	if !strings.Contains(got, "...") {
		t.Errorf("expected ellipsis in middle, got %q", got)
	}
	if len([]rune(got)) > 11 {
		t.Errorf("expected length <= 11, got %d: %q", len([]rune(got)), got)
	}
}

func TestTruncateTransform_Integration(t *testing.T) {
	fn, err := TruncateTransform(8, TruncateEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := fn("hello world")
	if len([]rune(got)) != 8 {
		t.Errorf("expected length 8, got %d: %q", len([]rune(got)), got)
	}
}

func TestTruncateTransform_InvalidArgs(t *testing.T) {
	_, err := TruncateTransform(-1, TruncateEnd)
	if err == nil {
		t.Fatal("expected error for negative maxLen")
	}
}
