package pipelines

import (
	"context"
	"sync"
)

// StepResultProcess is a function that processes the input data and does not return any data.
type StepResultProcess[I any] func(I) error

// stepResult is a struct that represents a step in the pipeline that does not return any data.
type stepResult[I any] struct {
	step[I]
	process StepResultProcess[I]
}

// NewStepResult creates a new result step with the given label, number of replicas and process.
func NewStepResult[I any](label string, replicas uint16, process StepResultProcess[I]) IStep[I] {
	return &stepResult[I]{
		step:    newStep[I](label, replicas, nil),
		process: process,
	}
}

// NewStepResultWithErrorHandler creates a new result step with the given label, number of replicas, process and error handler.
func NewStepResultWithErrorHandler[I any](label string, replicas uint16, process StepResultProcess[I], reportErrorHandler ErrorHandler) IStep[I] {
	return &stepResult[I]{
		step:    newStep[I](label, replicas, reportErrorHandler),
		process: process,
	}
}

func (s *stepResult[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if ok {
				err := s.process(i)
				s.decrementTokensCount()
				if err != nil && s.errorHandler != nil {
					s.errorHandler(s.label, err)
				}
			}
		}
	}
}
