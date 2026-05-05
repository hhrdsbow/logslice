package slicer

import (
	"fmt"
	"sync"
	"testing"
)

func TestDedupFilter_AcceptsUniqueLines(t *testing.T) {
	d := NewDedupFilter()
	lines := []string{"alpha", "beta", "gamma"}
	for _, l := range lines {
		if !d.Accept(l) {
			t.Errorf("expected Accept(%q) = true, got false", l)
		}
	}
	total, skipped := d.Stats()
	if total != 3 || skipped != 0 {
		t.Errorf("expected total=3 skipped=0, got total=%d skipped=%d", total, skipped)
	}
}

func TestDedupFilter_RejectsDuplicates(t *testing.T) {
	d := NewDedupFilter()
	d.Accept("hello")
	if d.Accept("hello") {
		t.Error("expected second Accept(\"hello\") = false")
	}
	total, skipped := d.Stats()
	if total != 2 || skipped != 1 {
		t.Errorf("expected total=2 skipped=1, got total=%d skipped=%d", total, skipped)
	}
}

func TestDedupFilter_EmptyString(t *testing.T) {
	d := NewDedupFilter()
	if !d.Accept("") {
		t.Error("expected first empty string to be accepted")
	}
	if d.Accept("") {
		t.Error("expected second empty string to be rejected")
	}
}

func TestDedupFilter_Reset(t *testing.T) {
	d := NewDedupFilter()
	d.Accept("line1")
	d.Accept("line1")
	d.Reset()
	total, skipped := d.Stats()
	if total != 0 || skipped != 0 {
		t.Errorf("after Reset expected total=0 skipped=0, got total=%d skipped=%d", total, skipped)
	}
	// After reset the same line should be accepted again.
	if !d.Accept("line1") {
		t.Error("expected Accept after Reset to return true")
	}
}

func TestDedupFilter_ConcurrentSafe(t *testing.T) {
	d := NewDedupFilter()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			line := fmt.Sprintf("line-%d", n%10)
			d.Accept(line)
		}(i)
	}
	wg.Wait()
	total, _ := d.Stats()
	if total != 50 {
		t.Errorf("expected total=50, got %d", total)
	}
}
