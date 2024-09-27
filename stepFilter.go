package pipelines

import (
	"context"
	"sync"
)

// StepProcess is a function that processes a single input data and returns a single output data.
type StepFilterProcess[I any] func(I) bool

// StepConfig is a struct that defines the configuration for a standard step
type StepFilterConfig[I any] struct {
	Label        string
	Replicas     uint16
	PassCriteria StepFilterProcess[I]
}

type stepFilter[I any] struct {
	stepBase[I]
	// passCriteria is a function that will be applied to the incoming data.
	passCriteria StepFilterProcess[I]
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
