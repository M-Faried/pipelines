package pipelines

import (
	"context"
	"sync"
)

// StepBasicErrorHandler is the definition of error reporting handler which may or may not be set by the user during creation of the step.
// The first parameter is the label of the step where the error occurred and the second parameter is the error itself.
type StepBasicErrorHandler func(string, error)

// StepBasicProcess is a function that processes a single input data and returns a single output data.
type StepBasicProcess[I any] func(I) (I, error)

// StepBasicConfig is a struct that defines the configuration for a basic step
type StepBasicConfig[I any] struct {
	// Label is a human-readable label for the step
	Label string

	// InputChannelSize is the buffer size for the input channel to the step
	InputChannelSize uint16

	// Replicas is the number of replicas of the step that should be run in parallel
	Replicas uint16

	// ErrorHandler is a function that will be called when an error occurs in the step
	ErrorHandler StepBasicErrorHandler

	// Process is a function that will be applied to the incoming data
	Process StepBasicProcess[I]
}

type stepBasic[I any] struct {
	stepBase[I]

	// errorHandler is the function called when an error occurs in the step.
	errorHandler StepBasicErrorHandler

	// process is a function that will be applied to the incoming data.
	process StepBasicProcess[I]
}

func newStepBasic[I any](config StepBasicConfig[I]) IStep[I] {
	if config.Process == nil {
		panic("process is required")
	}
	return &stepBasic[I]{
		stepBase:     newBaseStep[I](config.Label, config.Replicas, config.InputChannelSize),
		errorHandler: config.ErrorHandler,
		process:      config.Process,
	}
}

func (s *stepBasic[I]) reportError(err error) {
	if s.errorHandler != nil {
		s.errorHandler(s.label, err)
	}
}

// run is a method that runs the step process and will be executed in a separate goroutine.
func (s *stepBasic[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
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
