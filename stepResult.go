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
	Process        StepResultProcess[I]
	ReverseProcess StepResultProcess[I]
}

// stepResult is a struct that represents a step in the pipeline that does not return any data.
type stepResult[I any] struct {
	stepBase[I]
	process        StepResultProcess[I]
	reverseProcess StepResultProcess[I]
}

func newStepResult[I any](config StepResultConfig[I]) IStep[I] {
	if config.Process == nil {
		panic("process is required")
	}
	return &stepResult[I]{
		stepBase:       newBaseStep[I](config.Label, config.Replicas),
		process:        config.Process,
		reverseProcess: config.ReverseProcess,
	}
}

func (s *stepResult[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
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

func (s *stepResult[I]) RunReverse(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case o, ok := <-s.output:
			if !ok {
				return
			}
			s.reverseProcess(o)
			s.incrementTokensCount()
		}
	}
}
