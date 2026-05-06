package slicer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotateWriter_WritesLines(t *testing.T) {
	dir := t.TempDir()
	rw, err := NewRotateWriter(dir, "log-", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rw.Close()

	for _, line := range []string{"alpha", "beta", "gamma"} {
		if err := rw.WriteLine(line); err != nil {
			t.Fatalf("WriteLine: %v", err)
		}
	}

	var buf strings.Builder
	if _, err := rw.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	got := buf.String()
	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output", want)
		}
	}
}

func TestRotateWriter_RotateOnMaxLines(t *testing.T) {
	dir := t.TempDir()
	rw, err := NewRotateWriter(dir, "seg-", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rw.Close()

	for i := 0; i < 5; i++ {
		if err := rw.WriteLine("line"); err != nil {
			t.Fatalf("WriteLine: %v", err)
		}
	}

	if rw.Rotations() < 2 {
		t.Errorf("expected at least 2 rotations, got %d", rw.Rotations())
	}

	files, _ := filepath.Glob(filepath.Join(dir, "seg-*.log"))
	if len(files) < 3 {
		t.Errorf("expected at least 3 files, got %d", len(files))
	}
}

func TestRotateWriter_RotateOnMaxBytes(t *testing.T) {
	dir := t.TempDir()
	// each "hello" line is 6 bytes (with newline); cap at 12 bytes
	rw, err := NewRotateWriter(dir, "byte-", 0, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rw.Close()

	for i := 0; i < 4; i++ {
		if err := rw.WriteLine("hello"); err != nil {
			t.Fatalf("WriteLine: %v", err)
		}
	}

	if rw.Rotations() < 1 {
		t.Errorf("expected at least 1 rotation, got %d", rw.Rotations())
	}
}

func TestRotateWriter_InvalidDir(t *testing.T) {
	// Use a file path as a directory to force failure.
	tmp, _ := os.CreateTemp("", "logslice-*")
	tmp.Close()
	defer os.Remove(tmp.Name())

	_, err := NewRotateWriter(tmp.Name(), "x-", 0, 0)
	if err == nil {
		t.Fatal("expected error for invalid dir, got nil")
	}
}

func TestRotateWriter_CloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	rw, err := NewRotateWriter(dir, "c-", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := rw.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	// second close on nil file should be safe
	rw.file = nil
	if err := rw.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}
