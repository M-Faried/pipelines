package pipelines

import (
	"context"
	"sync"
)

// StepProcess is a function that processes a single input data and returns a single output data.
type StepProcess[I any] func(I) (I, error)

// StepConfig is a struct that defines the configuration for a standard step
type StepConfig[I any] struct {
	Label        string
	Replicas     uint16
	ErrorHandler ErrorHandler
	Process      StepProcess[I]
}

type stepStandard[I any] struct {
	stepBase[I]
	// process is a function that will be applied to the incoming data.
	process StepProcess[I]
}

// run is a method that runs the step process and will be executed in a separate goroutine.
func (s *stepStandard[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if !ok {
				return
			}
			o, err := s.process(i)
			if err != nil {
				// since we will not proceed with the current token, we need to decrement the tokens count.
				s.decrementTokensCount()
				s.reportError(err)
			} else {
				s.output <- o
			}
		}
	}
}
