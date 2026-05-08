package slicer

import "context"

// AggregateStageOption configures NewAggregateStage.
type AggregateStageOption func(*aggregateStageCfg)

type aggregateStageCfg struct {
	bucketSize int
	aggrFn     AggregateFunc
}

// WithBucketSize sets how many input lines form one output bucket.
func WithBucketSize(n int) AggregateStageOption {
	return func(c *aggregateStageCfg) { c.bucketSize = n }
}

// WithAggregateFunc sets the reduction function applied to each bucket.
func WithAggregateFunc(fn AggregateFunc) AggregateStageOption {
	return func(c *aggregateStageCfg) { c.aggrFn = fn }
}

// NewAggregateStage returns a pipeline stage that groups input lines into
// buckets and emits one summary line per bucket. Any partial bucket is
// flushed when the input channel closes or the context is cancelled.
func NewAggregateStage(ctx context.Context, in <-chan string, opts ...AggregateStageOption) <-chan string {
	cfg := &aggregateStageCfg{bucketSize: 10, aggrFn: ConcatAggregate}
	for _, o := range opts {
		o(cfg)
	}

	agg := NewAggregator(cfg.bucketSize, cfg.aggrFn)
	out := make(chan string)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				if summary, ok := agg.Flush(); ok {
					select {
					case out <- summary:
					default:
					}
				}
				return
			case line, ok := <-in:
				if !ok {
					if summary, fok := agg.Flush(); fok {
						out <- summary
					}
					return
				}
				if summary, full := agg.Add(line); full {
					out <- summary
				}
			}
		}
	}()

	return out
}
