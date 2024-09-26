package pipelines

import (
	"context"
	"sync"
	"time"
)

// StepAggregatorProcess is the definition of the function that will be executed by the step.
// It returns the item (could be the pass through item or the result of aggregation)
// a boolean whether this item is aggregated or not, and another boolean indicating the
// type of item whether it is a pass through or an aggregation result.
type StepAggregatorProcess[I any] func(I) (I, bool, bool)

// StepAggregatorConfig is a struct that defines the configuration for an aggregator step
type StepAggregatorConfig[I any] struct {
	Label        string
	Replicas     uint16
	ErrorHandler ErrorHandler
	Process      StepAggregatorProcess[I]
}

// stepAggregator is a struct that defines an aggregator step
type stepAggregator[I any] struct {
	stepBase[I]
	process StepAggregatorProcess[I]
}

func (s *stepAggregator[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	// The flag to skip decrementing the tokens count once to count
	// for the aggregated data as a single item still in the pipelines.
	// If we don't have this flag, the decrement may hit zero before
	// the data in the aggregator is sent as aggregator result.
	skipOnceForAggregatedToken := false
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case <-time.NewTicker(1 * time.Second).C:

		case i, ok := <-s.input:
			if !ok {
				return
			}
			item, aggregated, aggregationResult := s.process(i)
			if aggregated {
				// we need to skip decrementing once since the aggregated element is going to
				if skipOnceForAggregatedToken {
					s.decrementTokensCount()
				} else {
					skipOnceForAggregatedToken = true
				}
			} else if aggregationResult {
				// resetting the skip once flag.
				skipOnceForAggregatedToken = false
				// passing the aggregation result to the next step in the pipeline
				s.output <- item
			} else {
				// since that the element was not aggregated
				// we can pass it through to the next step in
				// the pipeline
				s.output <- item
			}
		}
	}
}
