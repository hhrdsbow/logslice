package slicer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckpoint_UpdateAndSnapshot(t *testing.T) {
	cp := NewCheckpoint("/var/log/app.log")
	cp.Update(1024, 50)
	filePath, offset, line, savedAt := cp.Snapshot()
	if filePath != "/var/log/app.log" {
		t.Errorf("expected file path /var/log/app.log, got %s", filePath)
	}
	if offset != 1024 {
		t.Errorf("expected offset 1024, got %d", offset)
	}
	if line != 50 {
		t.Errorf("expected line 50, got %d", line)
	}
	if savedAt.IsZero() {
		t.Error("expected non-zero savedAt")
	}
}

func TestCheckpoint_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	cp := NewCheckpoint("/logs/service.log")
	cp.Update(4096, 200)
	if err := cp.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := LoadCheckpoint(path)
	if err != nil {
		t.Fatalf("LoadCheckpoint: %v", err)
	}
	if loaded.FilePath != "/logs/service.log" {
		t.Errorf("FilePath mismatch: %s", loaded.FilePath)
	}
	if loaded.Offset != 4096 {
		t.Errorf("Offset mismatch: %d", loaded.Offset)
	}
	if loaded.Line != 200 {
		t.Errorf("Line mismatch: %d", loaded.Line)
	}
}

func TestLoadCheckpoint_Missing(t *testing.T) {
	_, err := LoadCheckpoint("/nonexistent/path/cp.json")
	if !os.IsNotExist(err) {
		t.Errorf("expected not-exist error, got %v", err)
	}
}

func TestCheckpointStage_ForwardsLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cp.json")
	cp := NewCheckpoint("input.log")
	stage := NewCheckpointStage(cp, path, 2)

	in := make(chan string, 5)
	lines := []string{"alpha", "beta", "gamma", "delta"}
	for _, l := range lines {
		in <- l
	}
	close(in)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := stage.Run(ctx, in)

	var got []string
	for l := range out {
		got = append(got, l)
	}
	if len(got) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(got))
	}
	for i, l := range got {
		if l != lines[i] {
			t.Errorf("line %d: expected %q, got %q", i, lines[i], l)
		}
	}
	_, offset, line, _ := cp.Snapshot()
	if line != int64(len(lines)) {
		t.Errorf("expected line count %d, got %d", len(lines), line)
	}
	if offset == 0 {
		t.Error("expected non-zero byte offset")
	}
}
