package pipelines

import (
	"context"
	"sync"
)

// StepBasicProcess is a function that processes a single input data and returns a single output data.
type StepBasicProcess[I any] func(I) I

// StepBasicConfig is a struct that defines the configuration for a basic step
type StepBasicConfig[I any] struct {
	// Label is a human-readable label for the step
	Label string

	// InputChannelSize is the buffer size for the input channel to the step
	InputChannelSize uint16

	// Replicas is the number of replicas of the step that should be run in parallel
	Replicas uint16

	// Process is a function that will be applied to the incoming data
	Process StepBasicProcess[I]
}

type stepBasic[I any] struct {
	stepBase[I]

	// process is a function that will be applied to the incoming data.
	process StepBasicProcess[I]
}

func newStepBasic[I any](config StepBasicConfig[I]) IStep[I] {
	if config.Process == nil {
		panic("process is required")
	}
	return &stepBasic[I]{
		stepBase: newBaseStep[I](config.Label, config.Replicas, config.InputChannelSize),
		process:  config.Process,
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
			o := s.process(i)
			s.output <- o
		}
	}
}
