package slicer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputWriter_Stdout(t *testing.T) {
	var buf bytes.Buffer
	cfg := OutputConfig{
		Mode:   OutputStdout,
		Writer: &buf,
	}
	ow, err := NewOutputWriter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ow.WriteLine("hello world"); err != nil {
		t.Fatalf("WriteLine error: %v", err)
	}
	if err := ow.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestOutputWriter_File(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.log")
	cfg := OutputConfig{
		Mode:     OutputFile,
		FilePath: path,
	}
	ow, err := NewOutputWriter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := []string{"line one", "line two", "line three"}
	for _, l := range lines {
		if err := ow.WriteLine(l); err != nil {
			t.Fatalf("WriteLine error: %v", err)
		}
	}
	if err := ow.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	got := strings.TrimSpace(string(data))
	expected := strings.Join(lines, "\n")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestOutputWriter_Dir(t *testing.T) {
	tmp := t.TempDir()
	cfg := OutputConfig{
		Mode:    OutputDir,
		DirPath: filepath.Join(tmp, "slices"),
		Prefix:  "segment-01",
	}
	ow, err := NewOutputWriter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = ow.WriteLine("data line")
	if err := ow.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	entries, err := os.ReadDir(cfg.DirPath)
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	if entries[0].Name() != "segment-01.log" {
		t.Errorf("unexpected filename: %s", entries[0].Name())
	}
}

func TestOutputWriter_InvalidMode(t *testing.T) {
	_, err := NewOutputWriter(OutputConfig{Mode: OutputMode(99)})
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestOutputWriter_CloseIdempotent(t *testing.T) {
	var buf bytes.Buffer
	ow, _ := NewOutputWriter(OutputConfig{Mode: OutputStdout, Writer: &buf})
	if err := ow.Close(); err != nil {
		t.Fatalf("first Close error: %v", err)
	}
	if err := ow.Close(); err != nil {
		t.Fatalf("second Close error: %v", err)
	}
}
