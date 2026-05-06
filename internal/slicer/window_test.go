package slicer

import (
	"testing"
	"time"
)

func TestWindow_AddAndLen(t *testing.T) {
	w := NewWindow(5 * time.Minute)
	now := time.Now()
	w.Add("line1", now)
	w.Add("line2", now.Add(time.Minute))
	if got := w.Len(); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestWindow_EvictsOldEntries(t *testing.T) {
	w := NewWindow(5 * time.Minute)
	base := time.Now()
	w.Add("old", base)
	w.Add("new", base.Add(6*time.Minute)) // triggers eviction of "old"
	lines := w.Lines()
	if len(lines) != 1 {
		t.Fatalf("expected 1 line after eviction, got %d", len(lines))
	}
	if lines[0] != "new" {
		t.Errorf("expected 'new', got %q", lines[0])
	}
}

func TestWindow_Reset(t *testing.T) {
	w := NewWindow(time.Hour)
	now := time.Now()
	w.Add("a", now)
	w.Add("b", now)
	w.Reset()
	if w.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", w.Len())
	}
}

func TestWindow_LinesSnapshot(t *testing.T) {
	w := NewWindow(time.Hour)
	now := time.Now()
	w.Add("x", now)
	w.Add("y", now.Add(time.Second))
	lines := w.Lines()
	if len(lines) != 2 || lines[0] != "x" || lines[1] != "y" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestWindow_EmptyWindow(t *testing.T) {
	w := NewWindow(time.Minute)
	if w.Len() != 0 {
		t.Fatal("new window should be empty")
	}
	if lines := w.Lines(); len(lines) != 0 {
		t.Fatal("Lines() on empty window should return empty slice")
	}
}
