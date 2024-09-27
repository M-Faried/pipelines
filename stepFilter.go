package pipelines

import (
	"context"
	"sync"
)

// StepFilterPassCriteria is function that determines if the data should be passed or not.
type StepFilterPassCriteria[I any] func(I) bool

// StepFilterConfig is a struct that defines the configuration for a filter step. The filter step filters the incoming data based on a certain criteria.
type StepFilterConfig[I any] struct {
	Label    string
	Replicas uint16
	// PassCriteria is a function that determines if the data should be passed or not.
	PassCriteria StepFilterPassCriteria[I]
}

type stepFilter[I any] struct {
	stepBase[I]
	passCriteria StepFilterPassCriteria[I]
}

// run is a method that runs the step process and will be executed in a separate goroutine.
func (s *stepFilter[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if !ok {
				wg.Done()
				return
			}
			if s.passCriteria(i) {
				s.output <- i
			} else {
				s.decrementTokensCount()
			}
		}
	}
}
