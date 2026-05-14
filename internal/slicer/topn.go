package slicer

import (
	"fmt"
	"io"
	"sort"
	"sync"
)

// TopNEntry holds a line and its associated score.
type TopNEntry struct {
	Line  string
	Score float64
}

// ScorerFunc assigns a numeric score to a line.
type ScorerFunc func(line string) float64

// TopN collects the highest-scored N lines seen.
type TopN struct {
	mu      sync.Mutex
	n       int
	scorer  ScorerFunc
	entries []TopNEntry
}

// NewTopN creates a TopN collector that keeps the top n lines by score.
// n is clamped to at least 1. scorer must not be nil.
func NewTopN(n int, scorer ScorerFunc) *TopN {
	if n < 1 {
		n = 1
	}
	if scorer == nil {
		scorer = func(line string) float64 { return float64(len(line)) }
	}
	return &TopN{n: n, scorer: scorer}
}

// Add scores a line and inserts it if it belongs in the top N.
func (t *TopN) Add(line string) {
	score := t.scorer(line)
	t.mu.Lock()
	defer t.mu.Unlock()

	t.entries = append(t.entries, TopNEntry{Line: line, Score: score})
	sort.Slice(t.entries, func(i, j int) bool {
		return t.entries[i].Score > t.entries[j].Score
	})
	if len(t.entries) > t.n {
		t.entries = t.entries[:t.n]
	}
}

// Snapshot returns a copy of the current top entries in descending order.
func (t *TopN) Snapshot() []TopNEntry {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TopNEntry, len(t.entries))
	copy(out, t.entries)
	return out
}

// Reset clears all collected entries.
func (t *TopN) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = t.entries[:0]
}

// WriteSummary writes the top entries to w.
func (t *TopN) WriteSummary(w io.Writer) {
	entries := t.Snapshot()
	for i, e := range entries {
		fmt.Fprintf(w, "#%d score=%.2f %s\n", i+1, e.Score, e.Line)
	}
}

// NewTopNStage returns a pipeline stage that passes all lines through while
// collecting the top N into collector. Call collector.WriteSummary after
// the pipeline completes to print results.
func NewTopNStage(in <-chan string, collector *TopN) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for line := range in {
			collector.Add(line)
			out <- line
		}
	}()
	return out
}
