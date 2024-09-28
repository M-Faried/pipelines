package pipelines

import (
	"context"
	"sync"
)

// StepResultProcess is a function that processes the input data and does not return any data.
type StepResultProcess[I any] func(I) error

// StepResultConfig is a struct that defines the configuration for a result step
type StepResultConfig[I any] struct {

	// Label is the name of the step.
	Label string

	// Replicas is the number of replicas (go routines) created to run the step.
	Replicas uint16

	// ErrorHandler is the function that handles the error occurred during the processing of the token.
	ErrorHandler ErrorHandler

	// Process is the function that processes the input data and does not return any data.
	Process StepResultProcess[I]
}

// stepResult is a struct that represents a step in the pipeline that does not return any data.
type stepResult[I any] struct {
	stepBase[I]
	process StepResultProcess[I]
}

func (s *stepResult[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
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
			err := s.process(i)
			s.decrementTokensCount()
			if err != nil {
				s.reportError(err)
			}
		}
	}
}
