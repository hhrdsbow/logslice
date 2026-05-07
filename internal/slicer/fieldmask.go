package slicer

import (
	"regexp"
	"strings"
)

// FieldMask replaces matched capture groups in a line with a fixed mask string.
// It is useful for redacting structured fields such as IPs, tokens, or emails
// while preserving the surrounding log context.
type FieldMask struct {
	re   *regexp.Regexp
	mask string
}

// NewFieldMask creates a FieldMask using the given regular expression pattern
// and replacement mask. The pattern must compile successfully.
// If mask is empty, "***" is used as a default.
func NewFieldMask(pattern, mask string) (*FieldMask, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	if mask == "" {
		mask = "***"
	}
	return &FieldMask{re: re, mask: mask}, nil
}

// Apply replaces all substrings in line that match the pattern with the mask.
// If the pattern contains named or unnamed subgroups, only the first subgroup
// match is replaced; otherwise the whole match is replaced.
func (f *FieldMask) Apply(line string) string {
	return f.re.ReplaceAllStringFunc(line, func(match string) string {
		subs := f.re.FindStringSubmatch(match)
		if len(subs) > 1 {
			// Replace only the captured group within the full match.
			return strings.Replace(match, subs[1], f.mask, 1)
		}
		return f.mask
	})
}

// FieldMaskTransform returns a TransformFunc that applies the FieldMask to
// each line, suitable for use with NewTransformer.
func FieldMaskTransform(fm *FieldMask) TransformFunc {
	return func(line string) string {
		return fm.Apply(line)
	}
}

// NewFieldMaskStage constructs a pipeline stage that applies multiple
// FieldMask rules in sequence to every line passing through the channel.
func NewFieldMaskStage(in <-chan string, masks ...*FieldMask) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for line := range in {
			for _, m := range masks {
				line = m.Apply(line)
			}
			out <- line
		}
	}()
	return out
}
