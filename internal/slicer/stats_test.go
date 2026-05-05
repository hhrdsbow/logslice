package slicer

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestStats_Duration_Finished(t *testing.T) {
	s := &Stats{
		StartedAt:  time.Now().Add(-2 * time.Second),
		FinishedAt: time.Now(),
	}
	if s.Duration() < time.Second {
		t.Errorf("expected duration >= 1s, got %s", s.Duration())
	}
}

func TestStats_Duration_Running(t *testing.T) {
	s := &Stats{StartedAt: time.Now().Add(-100 * time.Millisecond)}
	if s.Duration() <= 0 {
		t.Error("expected positive duration for running stats")
	}
}

func TestStats_MatchRatio_Zero(t *testing.T) {
	s := &Stats{}
	if s.MatchRatio() != 0 {
		t.Errorf("expected 0 ratio, got %f", s.MatchRatio())
	}
}

func TestStats_MatchRatio_Half(t *testing.T) {
	s := &Stats{LinesRead: 10, LinesMatched: 5}
	if s.MatchRatio() != 0.5 {
		t.Errorf("expected 0.5, got %f", s.MatchRatio())
	}
}

func TestStats_WriteSummary(t *testing.T) {
	s := &Stats{
		LinesRead:     100,
		LinesMatched:  42,
		SegmentsTotal: 3,
		BytesWritten:  2048,
		StartedAt:     time.Now().Add(-500 * time.Millisecond),
		FinishedAt:    time.Now(),
	}
	var buf bytes.Buffer
	s.WriteSummary(&buf)
	out := buf.String()
	for _, want := range []string{"100", "42", "42.0%", "3", "2048"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q; got:\n%s", want, out)
		}
	}
}

func TestStatsCollector_RecordLine(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordLine(true)
	sc.RecordLine(false)
	sc.RecordLine(true)
	if sc.LinesRead != 3 {
		t.Errorf("expected LinesRead=3, got %d", sc.LinesRead)
	}
	if sc.LinesMatched != 2 {
		t.Errorf("expected LinesMatched=2, got %d", sc.LinesMatched)
	}
}

func TestStatsCollector_RecordSegmentAndBytes(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordSegment()
	sc.RecordSegment()
	sc.RecordBytes(512)
	sc.RecordBytes(256)
	if sc.SegmentsTotal != 2 {
		t.Errorf("expected 2 segments, got %d", sc.SegmentsTotal)
	}
	if sc.BytesWritten != 768 {
		t.Errorf("expected 768 bytes, got %d", sc.BytesWritten)
	}
}

func TestStatsCollector_Finish(t *testing.T) {
	sc := NewStatsCollector()
	if !sc.FinishedAt.IsZero() {
		t.Error("FinishedAt should be zero before Finish()")
	}
	sc.Finish()
	if sc.FinishedAt.IsZero() {
		t.Error("FinishedAt should be set after Finish()")
	}
}
