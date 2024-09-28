package pipelines

import (
	"context"
	"sync"
)

// StepResultProcess is a function that processes the input data and does not return any data.
type StepResultProcess[I any] func(I)

// StepResultConfig is a struct that defines the configuration for a result step
type StepResultConfig[I any] struct {

	// Label is the name of the step.
	Label string

	// Replicas is the number of replicas (go routines) created to run the step.
	Replicas uint16

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
			s.process(i)
			s.decrementTokensCount()
		}
	}
}
