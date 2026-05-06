package slicer

import (
	"testing"
)

func TestSampler_KeepEveryLine_WhenNIsOne(t *testing.T) {
	s := NewSampler(1)
	for i := 0; i < 10; i++ {
		if !s.Keep("line") {
			t.Fatalf("expected every line to be kept at rate=1, failed at iteration %d", i)
		}
	}
}

func TestSampler_KeepEveryNthLine(t *testing.T) {
	s := NewSampler(3)
	kept := 0
	for i := 0; i < 9; i++ {
		if s.Keep("line") {
			kept++
		}
	}
	if kept != 3 {
		t.Fatalf("expected 3 lines kept out of 9 at rate=3, got %d", kept)
	}
}

func TestSampler_ZeroTreatedAsOne(t *testing.T) {
	s := NewSampler(0)
	if s.Rate() != 1 {
		t.Fatalf("expected rate=1 when constructed with 0, got %d", s.Rate())
	}
	if !s.Keep("line") {
		t.Fatal("expected first line to be kept when rate=1")
	}
}

func TestSampler_Reset(t *testing.T) {
	s := NewSampler(2)
	s.Keep("a") // counter = 1, not kept
	s.Keep("b") // counter = 2, kept
	s.Reset()
	// After reset counter is 0; first Keep increments to 1 → not kept
	if s.Keep("c") {
		t.Fatal("expected first line after reset to not be kept at rate=2")
	}
	if !s.Keep("d") {
		t.Fatal("expected second line after reset to be kept at rate=2")
	}
}

func TestSampleStage_FiltersCorrectly(t *testing.T) {
	in := make(chan string, 6)
	done := make(chan struct{})
	defer close(done)

	lines := []string{"a", "b", "c", "d", "e", "f"}
	for _, l := range lines {
		in <- l
	}
	close(in)

	stage := SampleStage(2)
	out := stage(in, done)

	var got []string
	for line := range out {
		got = append(got, line)
	}

	// rate=2: keep lines at positions 2,4,6 → "b","d","f"
	expected := []string{"b", "d", "f"}
	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
	for i, v := range expected {
		if got[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, got[i])
		}
	}
}

func TestSampleStage_CancelViaDone(t *testing.T) {
	in := make(chan string)
	done := make(chan struct{})

	stage := SampleStage(1)
	out := stage(in, done)

	close(done)
	// out should be closed after done is signalled
	_, ok := <-out
	if ok {
		t.Fatal("expected output channel to be closed after done")
	}
}
