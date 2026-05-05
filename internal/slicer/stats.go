package slicer

import (
	"fmt"
	"io"
	"time"
)

// Stats holds aggregated statistics for a slicing run.
type Stats struct {
	LinesRead     int64
	LinesMatched  int64
	SegmentsTotal int
	BytesWritten  int64
	StartedAt     time.Time
	FinishedAt    time.Time
}

// Duration returns the elapsed time of the slicing run.
func (s *Stats) Duration() time.Duration {
	if s.FinishedAt.IsZero() {
		return time.Since(s.StartedAt)
	}
	return s.FinishedAt.Sub(s.StartedAt)
}

// MatchRatio returns the fraction of lines that matched, or 0 if no lines were read.
func (s *Stats) MatchRatio() float64 {
	if s.LinesRead == 0 {
		return 0
	}
	return float64(s.LinesMatched) / float64(s.LinesRead)
}

// WriteSummary writes a human-readable summary to w.
func (s *Stats) WriteSummary(w io.Writer) {
	fmt.Fprintf(w, "Lines read    : %d\n", s.LinesRead)
	fmt.Fprintf(w, "Lines matched : %d (%.1f%%)\n", s.LinesMatched, s.MatchRatio()*100)
	fmt.Fprintf(w, "Segments      : %d\n", s.SegmentsTotal)
	fmt.Fprintf(w, "Bytes written : %d\n", s.BytesWritten)
	fmt.Fprintf(w, "Duration      : %s\n", s.Duration().Round(time.Millisecond))
}

// StatsCollector wraps a Stats value and provides thread-safe increment helpers.
type StatsCollector struct {
	Stats
}

// NewStatsCollector creates a StatsCollector with StartedAt set to now.
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		Stats: Stats{StartedAt: time.Now()},
	}
}

// RecordLine increments LinesRead and, if matched, LinesMatched.
func (sc *StatsCollector) RecordLine(matched bool) {
	sc.LinesRead++
	if matched {
		sc.LinesMatched++
	}
}

// RecordSegment increments the segment counter.
func (sc *StatsCollector) RecordSegment() {
	sc.SegmentsTotal++
}

// RecordBytes adds n to BytesWritten.
func (sc *StatsCollector) RecordBytes(n int64) {
	sc.BytesWritten += n
}

// Finish marks the run as complete.
func (sc *StatsCollector) Finish() {
	sc.FinishedAt = time.Now()
}
