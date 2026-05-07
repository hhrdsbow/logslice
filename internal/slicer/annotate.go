package slicer

import (
	"fmt"
	"strings"
	"time"
)

// AnnotateFunc is a function that produces a prefix string for a given line.
type AnnotateFunc func(line string, lineNum int64) string

// Annotator prepends a generated prefix to each line.
type Annotator struct {
	fns []AnnotateFunc
	sep string
}

// NewAnnotator creates an Annotator that applies the given AnnotateFuncs in
// order, joining their outputs with sep before prepending to each line.
func NewAnnotator(sep string, fns ...AnnotateFunc) *Annotator {
	if sep == "" {
		sep = " "
	}
	return &Annotator{fns: fns, sep: sep}
}

// Apply returns the annotated version of line.
func (a *Annotator) Apply(line string, lineNum int64) string {
	if len(a.fns) == 0 {
		return line
	}
	parts := make([]string, 0, len(a.fns)+1)
	for _, fn := range a.fns {
		if p := fn(line, lineNum); p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return line
	}
	return strings.Join(parts, a.sep) + a.sep + line
}

// LineNumberAnnotation returns an AnnotateFunc that prefixes lines with their
// 1-based line number formatted as [N].
func LineNumberAnnotation() AnnotateFunc {
	return func(_ string, lineNum int64) string {
		return fmt.Sprintf("[%d]", lineNum)
	}
}

// TimestampAnnotation returns an AnnotateFunc that prefixes lines with the
// current wall-clock time in the given layout (defaults to time.RFC3339).
func TimestampAnnotation(layout string) AnnotateFunc {
	if layout == "" {
		layout = time.RFC3339
	}
	return func(_ string, _ int64) string {
		return time.Now().Format(layout)
	}
}

// PrefixAnnotation returns an AnnotateFunc that always prepends a fixed string.
func PrefixAnnotation(prefix string) AnnotateFunc {
	return func(_ string, _ int64) string {
		return prefix
	}
}
