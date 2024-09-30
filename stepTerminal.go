package pipelines

import (
	"context"
	"sync"
)

// StepTerminalProcess is a function that processes the input data and does not return any data.
type StepTerminalProcess[I any] func(I)

// StepTerminalConfig is a struct that defines the configuration for a terminal step
type StepTerminalConfig[I any] struct {

	// Label is the name of the step.
	Label string

	// InputChannelSize is the buffer size for the input channel to the step
	InputChannelSize uint16

	// Replicas is the number of replicas (go routines) created to run the step.
	Replicas uint16

	// Process is the function that processes the input data and does not return any data.
	Process StepTerminalProcess[I]
}

// stepTerminal is a struct that represents a step in the pipeline that does not return any data.
type stepTerminal[I any] struct {
	stepBase[I]
	process StepTerminalProcess[I]
}

func newStepTerminal[I any](config StepTerminalConfig[I]) IStep[I] {
	if config.Process == nil {
		panic("process is required")
	}
	return &stepTerminal[I]{
		stepBase: newBaseStep[I](config.Label, config.Replicas, config.InputChannelSize),
		process:  config.Process,
	}
}

func (s *stepTerminal[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case i, ok := <-s.input:
			if !ok {
				return
			}
			s.process(i)
			s.decrementTokensCount()
		}
	}
}
