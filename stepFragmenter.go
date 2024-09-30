package pipelines

import (
	"context"
	"sync"
)

// StepFragmenterProcess is a function that converts a token in the pipeline into multiple tokens.
type StepFragmenterProcess[I any] func(I) []I

// StepFragmenterConfig is a struct that defines the configuration for a fragmenter step
type StepFragmenterConfig[I any] struct {

	// Label is the name of the step.
	Label string

	// InputChannelSize is the buffer size for the input channel to the step
	InputChannelSize uint16

	// Replicas is the number of replicas (go routines) created to run the step.
	Replicas uint16

	// Process is the function that converts a token in the pipeline into multiple tokens.
	Process StepFragmenterProcess[I]
}

type stepFragmenter[I any] struct {
	stepBase[I]
	process StepFragmenterProcess[I]
}

func newStepFragmenter[I any](config StepFragmenterConfig[I]) IStep[I] {
	if config.Process == nil {
		panic("process is required")
	}
	return &stepFragmenter[I]{
		stepBase: newBaseStep[I](config.Label, config.Replicas, config.InputChannelSize),
		process:  config.Process,
	}
}

func (s *stepFragmenter[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case i, ok := <-s.input:
			if !ok {
				return
			}
			outFragments := s.process(i)
			for _, fragment := range outFragments {
				// adding fragmented tokens to the count.
				s.incrementTokensCount()
				s.output <- fragment
			}
			// whether the token is framented or filtered with error, it is discarded from the pipeline.
			s.decrementTokensCount()
		}
	}
}
