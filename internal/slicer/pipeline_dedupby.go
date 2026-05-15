package slicer

import "fmt"

// DedupByOption configures a DedupBy pipeline stage.
type DedupByOption func(*dedupByConfig)

type dedupByConfig struct {
	extract  DedupByExtractor
	pattern string
}

// WithDedupByExtractor sets a custom key extractor function.
func WithDedupByExtractor(fn DedupByExtractor) DedupByOption {
	return func(c *dedupByConfig) { c.extract = fn }
}

// WithDedupByPattern sets a regex pattern whose first capture group (or full
// match) is used as the dedup key.
func WithDedupByPattern(pattern string) DedupByOption {
	return func(c *dedupByConfig) { c.pattern = pattern }
}

// NewDedupByStageFromOptions builds a pipeline stage function from the given
// options. Returns an error if the configuration is invalid.
func NewDedupByStageFromOptions(opts ...DedupByOption) (func(in <-chan string) <-chan string, error) {
	cfg := &dedupByConfig{}
	for _, o := range opts {
		o(cfg)
	}

	var f *DedupByFilter
	switch {
	case cfg.pattern != "":
		var err error
		f, err = NewRegexDedupByFilter(cfg.pattern)
		if err != nil {
			return nil, fmt.Errorf("pipeline dedupby: %w", err)
		}
	case cfg.extract != nil:
		f = NewDedupByFilter(cfg.extract)
	default:
		// Whole-line dedup — equivalent to DedupFilter.
		f = NewDedupByFilter(nil)
	}

	return NewDedupByStage(f), nil
}
