package slicer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

// RotateWriter writes lines to sequentially numbered files, rotating when
// a file reaches maxLines or maxBytes. Zero values disable the limit.
type RotateWriter struct {
	dir      string
	prefix   string
	maxLines int64
	maxBytes int64

	file      *os.File
	fileIndex int
	lines     int64
	bytes     int64
	rotations int64
}

// NewRotateWriter creates a RotateWriter that writes files under dir.
func NewRotateWriter(dir, prefix string, maxLines, maxBytes int64) (*RotateWriter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("rotate: mkdir %s: %w", dir, err)
	}
	rw := &RotateWriter{
		dir:      dir,
		prefix:   prefix,
		maxLines: maxLines,
		maxBytes: maxBytes,
	}
	if err := rw.openNext(); err != nil {
		return nil, err
	}
	return rw, nil
}

// WriteLine writes a single line followed by a newline, rotating if needed.
func (rw *RotateWriter) WriteLine(line string) error {
	if rw.needsRotation(int64(len(line) + 1)) {
		if err := rw.rotate(); err != nil {
			return err
		}
	}
	n, err := fmt.Fprintln(rw.file, line)
	if err != nil {
		return fmt.Errorf("rotate: write: %w", err)
	}
	rw.lines++
	rw.bytes += int64(n)
	return nil
}

// Rotations returns the number of file rotations that have occurred.
func (rw *RotateWriter) Rotations() int64 { return atomic.LoadInt64(&rw.rotations) }

// Close closes the current underlying file.
func (rw *RotateWriter) Close() error {
	if rw.file != nil {
		return rw.file.Close()
	}
	return nil
}

func (rw *RotateWriter) needsRotation(incoming int64) bool {
	if rw.maxLines > 0 && rw.lines >= rw.maxLines {
		return true
	}
	if rw.maxBytes > 0 && rw.bytes+incoming > rw.maxBytes {
		return true
	}
	return false
}

func (rw *RotateWriter) rotate() error {
	if err := rw.file.Close(); err != nil {
		return fmt.Errorf("rotate: close: %w", err)
	}
	atomic.AddInt64(&rw.rotations, 1)
	return rw.openNext()
}

func (rw *RotateWriter) openNext() error {
	name := filepath.Join(rw.dir, fmt.Sprintf("%s%04d.log", rw.prefix, rw.fileIndex))
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("rotate: open %s: %w", name, err)
	}
	rw.file = f
	rw.fileIndex++
	rw.lines = 0
	rw.bytes = 0
	return nil
}

// WriteTo copies all content from the current file to w (seek to start first).
func (rw *RotateWriter) WriteTo(w io.Writer) (int64, error) {
	if _, err := rw.file.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	return io.Copy(w, rw.file)
}
