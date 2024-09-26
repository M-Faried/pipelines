package pipelines

import (
	"context"
	"sync"
)

// StepAggregatorProcess is the definition of the function that will be executed by the step.
type StepAggregatorProcess[I any] func(I) error

// StepAggregatorConfig is a struct that defines the configuration for an aggregator step
type StepAggregatorConfig[I any] struct {
	Label                 string
	Replicas              uint16
	ErrorHandler          ErrorHandler
	AggregatedDataChannel <-chan I
	Process               StepAggregatorProcess[I]
}

// stepAggregator is a struct that defines an aggregator step
type stepAggregator[I any] struct {
	stepBase[I]
	aggregatedDataChannel <-chan I
	process               StepAggregatorProcess[I]
}

func (s *stepAggregator[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if !ok {
				return
			}
			err := s.process(i)
			if err != nil && s.errorHandler != nil {
				s.errorHandler(s.label, err)
			}
			// whether there is an error or something is done, we have to decrement the tokens count.
			s.decrementTokensCount()
		case aggData, ok := <-s.aggregatedDataChannel:
			if !ok {
				return
			}
			s.incrementTokensCount()
			s.output <- aggData
		}
	}
}
