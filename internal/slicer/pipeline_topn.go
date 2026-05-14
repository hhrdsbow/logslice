package slicer

// TopNOption configures a TopN pipeline stage.
type TopNOption func(*topNConfig)

type topNConfig struct {
	n      int
	scorer ScorerFunc
}

// WithTopN sets the number of top entries to retain.
func WithTopN(n int) TopNOption {
	return func(c *topNConfig) { c.n = n }
}

// WithTopNScorer sets the scoring function used to rank lines.
func WithTopNScorer(scorer ScorerFunc) TopNOption {
	return func(c *topNConfig) { c.scorer = scorer }
}

// TopNResult bundles the output channel with the underlying TopN collector
// so callers can retrieve the final ranked entries after the pipeline drains.
type TopNResult struct {
	Out       <-chan string
	Collector *TopN
}

// NewTopNPipelineStage wires a TopN collector into the pipeline.
// All lines are forwarded unchanged; the collector accumulates the top N.
// After the returned channel closes, call Result.Collector.Snapshot() or
// Result.Collector.WriteSummary() to inspect results.
func NewTopNPipelineStage(in <-chan string, opts ...TopNOption) TopNResult {
	cfg := &topNConfig{n: 10}
	for _, o := range opts {
		o(cfg)
	}
	collector := NewTopN(cfg.n, cfg.scorer)
	return TopNResult{
		Out:       NewTopNStage(in, collector),
		Collector: collector,
	}
}
