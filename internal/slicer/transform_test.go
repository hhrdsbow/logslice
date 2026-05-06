package slicer

import (
	"strings"
	"testing"
)

func TestTransformer_Apply_NoFuncs(t *testing.T) {
	tr := NewTransformer()
	got := tr.Apply("hello world")
	if got != "hello world" {
		t.Errorf("expected unchanged line, got %q", got)
	}
}

func TestTransformer_Apply_Chain(t *testing.T) {
	tr := NewTransformer(
		TrimSpaceTransform(),
		LowercaseTransform(),
	)
	got := tr.Apply("  HELLO WORLD  ")
	if got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestTrimSpaceTransform(t *testing.T) {
	fn := TrimSpaceTransform()
	got := fn("  trimmed  ")
	if got != "trimmed" {
		t.Errorf("expected %q, got %q", "trimmed", got)
	}
}

func TestLowercaseTransform(t *testing.T) {
	fn := LowercaseTransform()
	got := fn("UPPER CASE")
	if got != "upper case" {
		t.Errorf("expected %q, got %q", "upper case", got)
	}
}

func TestRedactTransform(t *testing.T) {
	fn := RedactTransform("secret", "[REDACTED]")
	got := fn("password=secret token=secret")
	want := "password=[REDACTED] token=[REDACTED]"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestTruncateTransform(t *testing.T) {
	fn := TruncateTransform(5)
	got := fn("hello world")
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
	// Short line should be unchanged.
	got = fn("hi")
	if got != "hi" {
		t.Errorf("expected %q, got %q", "hi", got)
	}
}

func TestTruncateTransform_Unicode(t *testing.T) {
	fn := TruncateTransform(3)
	got := fn("日本語テスト")
	if len([]rune(got)) != 3 {
		t.Errorf("expected 3 runes, got %d in %q", len([]rune(got)), got)
	}
}

func TestStripControlTransform(t *testing.T) {
	fn := StripControlTransform()
	input := "hello\x00world\x01\x1b[31m"
	got := fn(input)
	if strings.ContainsAny(got, "\x00\x01\x1b") {
		t.Errorf("control characters not stripped: %q", got)
	}
	if !strings.Contains(got, "hello") || !strings.Contains(got, "world") {
		t.Errorf("expected visible content preserved, got %q", got)
	}
}

func TestStripControlTransform_PreservesTab(t *testing.T) {
	fn := StripControlTransform()
	got := fn("col1\tcol2")
	if got != "col1\tcol2" {
		t.Errorf("expected tab preserved, got %q", got)
	}
}
