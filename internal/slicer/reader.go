package slicer

import (
	"bufio"
	"io"
)

// LineReader reads lines from an io.Reader and emits them on a channel.
// It supports cancellation via a done channel.
type LineReader struct {
	r       io.Reader
	bufSize int
}

// NewLineReader creates a LineReader wrapping the given io.Reader.
// bufSize controls the scanner buffer size (0 = default 64KB).
func NewLineReader(r io.Reader, bufSize int) *LineReader {
	if bufSize <= 0 {
		bufSize = bufio.MaxScanTokenSize
	}
	return &LineReader{r: r, bufSize: bufSize}
}

// Lines streams lines to the returned channel. The channel is closed when
// the reader is exhausted or done is closed. Any scan error is sent to errCh.
func (lr *LineReader) Lines(done <-chan struct{}) (<-chan string, <-chan error) {
	lines := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(lines)
		defer close(errCh)

		scanner := bufio.NewScanner(lr.r)
		buf := make([]byte, lr.bufSize)
		scanner.Buffer(buf, lr.bufSize)

		for scanner.Scan() {
			line := scanner.Text()
			select {
			case <-done:
				return
			case lines <- line:
			}
		}

		if err := scanner.Err(); err != nil {
			errCh <- err
		}
	}()

	return lines, errCh
}
