package slicer

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestProgress_Tick_Counts(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress(&buf, false)
	p.Tick(true, 10)
	p.Tick(false, 5)
	p.Tick(true, 20)

	if p.LinesRead() != 3 {
		t.Errorf("expected 3 lines, got %d", p.LinesRead())
	}
	if p.BytesMatched() != 30 {
		t.Errorf("expected 30 bytes matched, got %d", p.BytesMatched())
	}
}

func TestProgress_Tick_NoMatch(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress(&buf, false)
	p.Tick(false, 100)
	p.Tick(false, 200)

	if p.LinesRead() != 2 {
		t.Errorf("expected 2 lines read, got %d", p.LinesRead())
	}
	if p.BytesMatched() != 0 {
		t.Errorf("expected 0 bytes matched for non-matching ticks, got %d", p.BytesMatched())
	}
}

func TestProgress_Report_NonVerbose(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress(&buf, false)
	p.Tick(true, 100)
	p.Tick(true, 200)
	p.Report()

	out := buf.String()
	if !strings.Contains(out, "2 lines read") {
		t.Errorf("expected line count in output, got: %s", out)
	}
	if strings.Contains(out, "bytes") {
		t.Errorf("non-verbose should not contain bytes, got: %s", out)
	}
}

func TestProgress_Report_Verbose(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress(&buf, true)
	p.Tick(true, 50)
	p.Tick(false, 10)
	p.Report()

	out := buf.String()
	if !strings.Contains(out, "50 bytes matched") {
		t.Errorf("expected bytes matched in verbose output, got: %s", out)
	}
}

func TestProgress_Summary(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress(&buf, false)

	s := &Stats{
		LinesRead:     200,
		LinesMatched:  80,
		SegmentsTotal: 4,
		StartedAt:     time.Now().Add(-300 * time.Millisecond),
		FinishedAt:    time.Now(),
	}
	p.Summary(s)

	out := buf.String()
	for _, want := range []string{"80/200", "4 segment", "done:"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q; got: %s", want, out)
		}
	}
}
