package slicer

import (
	"fmt"
	"strings"
)

// TruncateMode controls how lines are truncated.
type TruncateMode int

const (
	// TruncateEnd cuts characters from the end of the line.
	TruncateEnd TruncateMode = iota
	// TruncateStart cuts characters from the beginning of the line.
	TruncateStart
	// TruncateMiddle replaces the middle with an ellipsis.
	TruncateMiddle
)

// Truncator limits line length, applying a chosen truncation strategy.
type Truncator struct {
	maxLen int
	mode   TruncateMode
	ellipsis string
}

// NewTruncator creates a Truncator that caps lines at maxLen runes.
// maxLen must be >= 4 when using TruncateMiddle (to fit the ellipsis).
func NewTruncator(maxLen int, mode TruncateMode) (*Truncator, error) {
	if maxLen <= 0 {
		return nil, fmt.Errorf("truncate: maxLen must be positive, got %d", maxLen)
	}
	if mode == TruncateMiddle && maxLen < 5 {
		return nil, fmt.Errorf("truncate: maxLen must be >= 5 for TruncateMiddle, got %d", maxLen)
	}
	return &Truncator{
		maxLen:   maxLen,
		mode:     mode,
		ellipsis: "...",
	}, nil
}

// Apply truncates line if it exceeds maxLen runes.
func (t *Truncator) Apply(line string) string {
	runes := []rune(line)
	if len(runes) <= t.maxLen {
		return line
	}
	switch t.mode {
	case TruncateEnd:
		return string(runes[:t.maxLen-len([]rune(t.ellipsis))]) + t.ellipsis
	case TruncateStart:
		cut := len(runes) - t.maxLen + len([]rune(t.ellipsis))
		return t.ellipsis + string(runes[cut:])
	case TruncateMiddle:
		half := (t.maxLen - len([]rune(t.ellipsis))) / 2
		return string(runes[:half]) + t.ellipsis + string(runes[len(runes)-half:])
	default:
		return strings.TrimRight(string(runes[:t.maxLen]), " ")
	}
}

// TruncateTransform returns a TransformFunc that truncates each line.
func TruncateTransform(maxLen int, mode TruncateMode) (TransformFunc, error) {
	t, err := NewTruncator(maxLen, mode)
	if err != nil {
		return nil, err
	}
	return func(line string) string {
		return t.Apply(line)
	}, nil
}
