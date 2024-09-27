package pipelines

import (
	"context"
	"sync"
)

// StepProcess is a function that processes a single input data and returns a single output data.
type FilterProcess[I any] func(I) bool

// StepConfig is a struct that defines the configuration for a standard step
type StepFilterConfig[I any] struct {
	Label    string
	Replicas uint16
	Process  FilterProcess[I]
}

type stepFilter[I any] struct {
	stepBase[I]
	// process is a function that will be applied to the incoming data.
	process FilterProcess[I]
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
				return
			}
			if !s.process(i) {
				s.output <- i
			} else {
				s.decrementTokensCount()
			}
		}
	}
}
