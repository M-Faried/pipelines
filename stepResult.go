package pipelines

import (
	"context"
	"sync"
)

// StepResultProcess is a function that processes the input data and does not return any data.
type StepResultProcess[I any] func(I) error

// StepResultConfig is a struct that defines the configuration for a result step
type StepResultConfig[I any] struct {
	Label        string
	Replicas     uint16
	ErrorHandler ErrorHandler
	Process      StepResultProcess[I]
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
