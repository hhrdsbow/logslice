package slicer

import (
	"sync"
	"time"
)

// Window holds lines within a sliding time window relative to a reference time.
type Window struct {
	mu       sync.Mutex
	lines    []string
	times    []time.Time
	duration time.Duration
}

// NewWindow creates a Window that retains lines within the given duration.
func NewWindow(d time.Duration) *Window {
	return &Window{duration: d}
}

// Add inserts a line with its associated timestamp into the window,
// then evicts entries older than w.duration relative to t.
func (w *Window) Add(line string, t time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lines = append(w.lines, line)
	w.times = append(w.times, t)
	w.evict(t)
}

// Lines returns a snapshot of all lines currently in the window.
func (w *Window) Lines() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]string, len(w.lines))
	copy(out, w.lines)
	return out
}

// Len returns the number of lines currently in the window.
func (w *Window) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.lines)
}

// Reset clears all entries from the window.
func (w *Window) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lines = w.lines[:0]
	w.times = w.times[:0]
}

// evict removes entries older than w.duration relative to now.
// Caller must hold w.mu.
func (w *Window) evict(now time.Time) {
	cutoff := now.Add(-w.duration)
	i := 0
	for i < len(w.times) && w.times[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		w.lines = w.lines[i:]
		w.times = w.times[i:]
	}
}
