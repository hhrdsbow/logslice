package slicer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// OutputMode determines where sliced output is written.
type OutputMode int

const (
	OutputStdout OutputMode = iota
	OutputFile
	OutputDir
)

// OutputConfig holds configuration for writing sliced output.
type OutputConfig struct {
	Mode     OutputMode
	FilePath string // used for OutputFile
	DirPath  string // used for OutputDir
	Prefix   string // filename prefix for OutputDir
	Writer   io.Writer // used for OutputStdout or custom writers
}

// OutputWriter wraps an io.Writer with buffering and optional file management.
type OutputWriter struct {
	cfg    OutputConfig
	buf    *bufio.Writer
	file   *os.File
	closed bool
}

// NewOutputWriter creates an OutputWriter based on the provided config.
func NewOutputWriter(cfg OutputConfig) (*OutputWriter, error) {
	ow := &OutputWriter{cfg: cfg}

	switch cfg.Mode {
	case OutputStdout:
		w := cfg.Writer
		if w == nil {
			w = os.Stdout
		}
		ow.buf = bufio.NewWriter(w)
	case OutputFile:
		f, err := os.Create(cfg.FilePath)
		if err != nil {
			return nil, fmt.Errorf("output: create file %q: %w", cfg.FilePath, err)
		}
		ow.file = f
		ow.buf = bufio.NewWriter(f)
	case OutputDir:
		if err := os.MkdirAll(cfg.DirPath, 0o755); err != nil {
			return nil, fmt.Errorf("output: create dir %q: %w", cfg.DirPath, err)
		}
		name := filepath.Join(cfg.DirPath, cfg.Prefix+".log")
		f, err := os.Create(name)
		if err != nil {
			return nil, fmt.Errorf("output: create file %q: %w", name, err)
		}
		ow.file = f
		ow.buf = bufio.NewWriter(f)
	default:
		return nil, fmt.Errorf("output: unknown mode %d", cfg.Mode)
	}

	return ow, nil
}

// WriteLine writes a single line followed by a newline.
func (ow *OutputWriter) WriteLine(line string) error {
	_, err := fmt.Fprintln(ow.buf, line)
	return err
}

// Close flushes buffered data and closes any open file.
func (ow *OutputWriter) Close() error {
	if ow.closed {
		return nil
	}
	ow.closed = true
	if err := ow.buf.Flush(); err != nil {
		return err
	}
	if ow.file != nil {
		return ow.file.Close()
	}
	return nil
}
