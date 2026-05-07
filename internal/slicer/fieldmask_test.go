package slicer

import (
	"testing"
)

func TestNewFieldMask_InvalidRegex(t *testing.T) {
	_, err := NewFieldMask("[invalid", "***")
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestFieldMask_DefaultMask(t *testing.T) {
	fm, err := NewFieldMask(`\d+`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := fm.Apply("error code 404 on line 12")
	want := "error code *** on line ***"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFieldMask_WholeMatch(t *testing.T) {
	fm, err := NewFieldMask(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`, "[IP]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := fm.Apply("client 192.168.1.1 connected from 10.0.0.5")
	want := "client [IP] connected from [IP]"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFieldMask_CaptureGroup(t *testing.T) {
	// Pattern captures the token value inside quotes; only the value is masked.
	fm, err := NewFieldMask(`token="([^"]+)"`, "REDACTED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := fm.Apply(`user=alice token="s3cr3t" level=info`)
	want := `user=alice token="REDACTED" level=info`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFieldMask_NoMatch(t *testing.T) {
	fm, err := NewFieldMask(`password=\S+`, "***")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	line := "user=alice level=info msg=login"
	got := fm.Apply(line)
	if got != line {
		t.Errorf("expected unchanged line, got %q", got)
	}
}

func TestFieldMaskTransform(t *testing.T) {
	fm, _ := NewFieldMask(`\d+`, "#")
	tf := FieldMaskTransform(fm)
	got := tf("request 99 took 300ms")
	want := "request # took #ms"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNewFieldMaskStage_MultipleRules(t *testing.T) {
	ip, _ := NewFieldMask(`\b\d{1,3}(?:\.\d{1,3}){3}\b`, "[IP]")
	num, _ := NewFieldMask(`\b\d+\b`, "#")

	in := make(chan string, 3)
	in <- "client 10.0.0.1 sent 42 bytes"
	in <- "no match here"
	in <- "port 8080 from 127.0.0.1"
	close(in)

	out := NewFieldMaskStage(in, ip, num)

	want := []string{
		"client [IP] sent # bytes",
		"no match here",
		"port # from [IP]",
	}
	for i, w := range want {
		got, ok := <-out
		if !ok {
			t.Fatalf("channel closed early at index %d", i)
		}
		if got != w {
			t.Errorf("line %d: got %q, want %q", i, got, w)
		}
	}
	if _, ok := <-out; ok {
		t.Error("expected channel to be closed")
	}
}
