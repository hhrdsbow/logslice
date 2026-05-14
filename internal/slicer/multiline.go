package slicer

import (
	"regexp"
	"strings"
)

// MultilineFolder folds consecutive lines that match a continuation pattern
// into a single logical line joined by a separator.
type MultilineFolder struct {
	start    *regexp.Regexp
	continue_ *regexp.Regexp
	sep      string
}

// MultilineOption configures a MultilineFolder.
type MultilineOption func(*MultilineFolder)

// WithStartPattern sets the regex that identifies the first line of a block.
func WithStartPattern(pat string) MultilineOption {
	return func(m *MultilineFolder) {
		m.start = regexp.MustCompile(pat)
	}
}

// WithContinuePattern sets the regex that identifies continuation lines.
func WithContinuePattern(pat string) MultilineOption {
	return func(m *MultilineFolder) {
		m.continue_ = regexp.MustCompile(pat)
	}
}

// WithJoinSep sets the separator used when joining folded lines (default: " ").
func WithJoinSep(sep string) MultilineOption {
	return func(m *MultilineFolder) { m.sep = sep }
}

// NewMultilineFolder creates a MultilineFolder.
// Either startPat or contPat (or both) must be non-empty.
func NewMultilineFolder(opts ...MultilineOption) (*MultilineFolder, error) {
	m := &MultilineFolder{sep: " "}
	for _, o := range opts {
		o(m)
	}
	if m.start == nil && m.continue_ == nil {
		return nil, errorf("multiline: at least one of start or continue pattern required")
	}
	return m, nil
}

// Fold reads lines from in and emits folded logical lines on out.
// A new block begins when a line matches the start pattern (if set).
// Lines matching the continue pattern are appended to the current block.
// Lines that match neither flush the current block and are emitted as-is.
func (m *MultilineFolder) Fold(in <-chan string, done <-chan struct{}) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		var buf []string
		send := func(line string) {
			select {
			case out <- line:
			case <-done:
			}
		}
		flush := func() {
			if len(buf) == 0 {
				return
			}
			send(strings.Join(buf, m.sep))
			buf = buf[:0]
		}
		for {
			select {
			case line, ok := <-in:
				if !ok {
					flush()
					return
				}
				switch {
				case m.start != nil && m.start.MatchString(line):
					flush()
					buf = append(buf, line)
				case m.continue_ != nil && m.continue_.MatchString(line):
					if len(buf) == 0 {
						buf = append(buf, line)
					} else {
						buf = append(buf, line)
					}
				default:
					flush()
					send(line)
				}
			case <-done:
				return
			}
		}
	}()
	return out
}

// errorf is a local helper to avoid importing fmt at package level.
func errorf(s string) error {
	return &multilineErr{msg: s}
}

type multilineErr struct{ msg string }

func (e *multilineErr) Error() string { return e.msg }
