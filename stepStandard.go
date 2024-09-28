package pipelines

import (
	"context"
	"sync"
)

// ErrorHandler is the definition of error reporting handler which may or may not be set by the user during creation of the step.
// The first parameter is the label of the step where the error occurred and the second parameter is the error itself.
type ErrorHandler func(string, error)

// StepProcess is a function that processes a single input data and returns a single output data.
type StepProcess[I any] func(I) (I, error)

// StepConfig is a struct that defines the configuration for a standard step
type StepConfig[I any] struct {
	// Label is a human-readable label for the step
	Label string

	// Replicas is the number of replicas of the step that should be run in parallel
	Replicas uint16

	// ErrorHandler is a function that will be called when an error occurs in the step
	ErrorHandler ErrorHandler

	// Process is a function that will be applied to the incoming data
	Process StepProcess[I]
}

type stepStandard[I any] struct {
	stepBase[I]

	// errorHandler is the function called when an error occurs in the step.
	errorHandler ErrorHandler

	// process is a function that will be applied to the incoming data.
	process StepProcess[I]
}

func (s *stepStandard[I]) reportError(err error) {
	if s.errorHandler != nil {
		s.errorHandler(s.label, err)
	}
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
				wg.Done()
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
