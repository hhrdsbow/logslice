package slicer

import (
	"hash/fnv"
	"sync"
)

// DedupFilter removes duplicate log lines using a hash-based seen set.
// It is safe for concurrent use.
type DedupFilter struct {
	mu   sync.Mutex
	seen map[uint64]struct{}
	skip uint64
	total uint64
}

// NewDedupFilter creates a new DedupFilter ready for use.
func NewDedupFilter() *DedupFilter {
	return &DedupFilter{
		seen: make(map[uint64]struct{}),
	}
}

// Accept returns true if the line has not been seen before.
// Duplicate lines are silently counted and rejected.
func (d *DedupFilter) Accept(line string) bool {
	h := hash(line)
	d.mu.Lock()
	defer d.mu.Unlock()
	d.total++
	if _, exists := d.seen[h]; exists {
		d.skip++
		return false
	}
	d.seen[h] = struct{}{}
	return true
}

// Stats returns the total lines seen and the number of duplicates skipped.
func (d *DedupFilter) Stats() (total, skipped uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.total, d.skip
}

// Reset clears the seen set and resets all counters.
func (d *DedupFilter) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[uint64]struct{})
	d.skip = 0
	d.total = 0
}

func hash(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
