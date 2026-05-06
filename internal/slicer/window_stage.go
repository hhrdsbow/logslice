package slicer

import (
	"context"
	"time"
)

// WindowStageOption configures a WindowStage.
type WindowStageOption func(*WindowStage)

// WindowStage filters lines through a sliding time Window, forwarding only
// lines whose parsed timestamp falls within the configured duration.
type WindowStage struct {
	win    *Window
	parser func(string) (time.Time, bool)
}

// NewWindowStage creates a stage that keeps lines within duration d.
// parser extracts a time.Time from a raw log line; if it returns false
// the line is forwarded unconditionally (no timestamp found).
func NewWindowStage(d time.Duration, parser func(string) (time.Time, bool)) *WindowStage {
	return &WindowStage{
		win:    NewWindow(d),
		parser: parser,
	}
}

// Run reads from in, applies the sliding window filter, and writes matching
// lines to the returned channel. It closes the output channel when in is
// exhausted or ctx is cancelled.
func (s *WindowStage) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string, 64)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				t, ok := s.parser(line)
				if !ok {
					// no timestamp — pass through
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
					continue
				}
				s.win.Add(line, t)
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// Snapshot returns all lines currently held in the window.
func (s *WindowStage) Snapshot() []string { return s.win.Lines() }
