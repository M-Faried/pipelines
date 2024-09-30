package pipelines

import (
	"context"
	"sync"
)

// StepFilterPassCriteria is function that determines if the data should be passed or not.
type StepFilterPassCriteria[I any] func(I) bool

// StepFilterConfig is a struct that defines the configuration for a filter step. The filter step filters the incoming data based on a certain criteria.
type StepFilterConfig[I any] struct {

	// Label is the name of the step.
	Label string

	// Replicas is the number of replicas (go routines) created to run the step.
	Replicas uint16

	// InputChannelSize is the buffer size for the input channel to the step
	InputChannelSize uint16

	// PassCriteria is a function that determines if the data should be passed or not.
	PassCriteria StepFilterPassCriteria[I]
}

type stepFilter[I any] struct {
	stepBase[I]
	passCriteria StepFilterPassCriteria[I]
}

func newStepFilter[I any](config StepFilterConfig[I]) IStep[I] {
	if config.PassCriteria == nil {
		panic("process is required")
	}
	return &stepFilter[I]{
		stepBase:     newBaseStep[I](config.Label, config.Replicas, config.InputChannelSize),
		passCriteria: config.PassCriteria,
	}
}

// run is a method that runs the step process and will be executed in a separate goroutine.
func (s *stepFilter[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case i, ok := <-s.input:
			if !ok {
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
