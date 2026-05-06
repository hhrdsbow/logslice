package slicer

import (
	"bufio"
	"context"
	"io"
	"time"
)

// TailOptions configures the tail follower behaviour.
type TailOptions struct {
	// PollInterval is how often to check for new data when EOF is reached.
	PollInterval time.Duration
	// MaxRetries is the number of consecutive empty reads before giving up.
	// Zero means follow indefinitely until ctx is cancelled.
	MaxRetries int
}

// TailReader follows a reader like `tail -f`, emitting lines over a channel.
// It stops when ctx is cancelled or MaxRetries consecutive empty reads occur.
type TailReader struct {
	r       io.Reader
	opts    TailOptions
	lines   chan string
}

// NewTailReader creates a TailReader wrapping r with the given options.
func NewTailReader(r io.Reader, opts TailOptions) *TailReader {
	if opts.PollInterval <= 0 {
		opts.PollInterval = 250 * time.Millisecond
	}
	return &TailReader{
		r:     r,
		opts:  opts,
		lines: make(chan string, 64),
	}
}

// Lines returns the channel on which new lines are emitted.
func (t *TailReader) Lines() <-chan string {
	return t.lines
}

// Follow starts reading from r, sending lines to the Lines channel.
// It blocks until ctx is cancelled or the retry limit is reached.
func (t *TailReader) Follow(ctx context.Context) error {
	defer close(t.lines)

	scanner := bufio.NewScanner(t.r)
	retries := 0

	for {
		if scanner.Scan() {
			retries = 0
			select {
			case t.lines <- scanner.Text():
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		// EOF — decide whether to keep polling.
		retries++
		if t.opts.MaxRetries > 0 && retries >= t.opts.MaxRetries {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(t.opts.PollInterval):
			// re-scan: bufio.Scanner doesn't support reset, so we rely on the
			// underlying reader advancing (e.g. an os.File or pipe).
			scanner = bufio.NewScanner(t.r)
		}
	}
}
