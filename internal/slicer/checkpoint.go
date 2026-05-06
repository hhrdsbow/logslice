package slicer

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Checkpoint holds the state needed to resume a slicing operation.
type Checkpoint struct {
	mu       sync.Mutex
	FilePath string    `json:"file_path"`
	Offset   int64     `json:"offset"`
	Line     int64     `json:"line"`
	SavedAt  time.Time `json:"saved_at"`
}

// NewCheckpoint creates a new Checkpoint for the given file path.
func NewCheckpoint(filePath string) *Checkpoint {
	return &Checkpoint{FilePath: filePath}
}

// Update atomically updates the offset and line counter.
func (c *Checkpoint) Update(offset, line int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Offset = offset
	c.Line = line
	c.SavedAt = time.Now()
}

// Save persists the checkpoint to a JSON file at the given path.
func (c *Checkpoint) Save(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, err := os.CreateTemp("", "checkpoint-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if err := json.NewEncoder(f).Encode(c); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, path)
}

// LoadCheckpoint reads a checkpoint from a JSON file.
func LoadCheckpoint(path string) (*Checkpoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cp Checkpoint
	if err := json.NewDecoder(f).Decode(&cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

// Snapshot returns a copy of the current checkpoint state without the mutex.
func (c *Checkpoint) Snapshot() (filePath string, offset, line int64, savedAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.FilePath, c.Offset, c.Line, c.SavedAt
}
