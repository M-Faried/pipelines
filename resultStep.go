package pipelines

import (
	"context"
	"sync"
)

type ResultStepProcess[I any] func(I) error

type ResultStep[I any] struct {
	baseStep[I]
	process ResultStepProcess[I]
}

// NewResultStep creates a new result step with the given id, number of replicas and process.
func NewResultStep[I any](id string, replicas uint16, process ResultStepProcess[I]) *ResultStep[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := &ResultStep[I]{}
	step.id = id
	step.replicas = replicas
	step.process = process
	return step
}

func NewResultStepWithErrorHandler[I any](id string, replicas uint16, process ResultStepProcess[I], reportErrorHandler ReportError) *ResultStep[I] {
	step := NewResultStep(id, replicas, process)
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
					s.reportError(s.id, err)
				}
			}
		}
	}
}
