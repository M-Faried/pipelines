package pipelines

import (
	"context"
	"sync"
)

// ResultStepProcess is a function that processes the input data and does not return any data.
type ResultStepProcess[I any] func(I) error

// ResultStep is a struct that represents a step in the pipeline that does not return any data.
type ResultStep[I any] struct {
	baseStep[I]
	process ResultStepProcess[I]
}

// NewResultStep creates a new result step with the given label, number of replicas and process.
func NewResultStep[I any](label string, replicas uint16, process ResultStepProcess[I]) *ResultStep[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := &ResultStep[I]{}
	step.label = label
	step.replicas = replicas
	step.process = process
	return step
}

// NewResultStepWithErrorHandler creates a new result step with the given label, number of replicas, process and error handler.
func NewResultStepWithErrorHandler[I any](label string, replicas uint16, process ResultStepProcess[I], reportErrorHandler ReportError) *ResultStep[I] {
	step := NewResultStep(label, replicas, process)
	step.reportError = reportErrorHandler
	return step
}

func (s *ResultStep[I]) run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if ok {
				err := s.process(i)
				s.decrementTokensCount()
				if err != nil && s.reportError != nil {
					s.reportError(s.label, err)
				}
			}
		}
	}
}
