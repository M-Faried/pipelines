package pipelines

import (
	"context"
	"sync"
)

type StepProcess[I any] func(I) (I, error)

type Step[I any] struct {
	baseStep[I]
	// ouput is a channel for outgoing data from the step.
	output chan I
	// process is a function that will be applied to the incoming data.
	process StepProcess[I]
}

// NewStep creates a new step with the given id, number of replicas and process.
func NewStep[I any](id string, replicas uint16, process StepProcess[I]) *Step[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := &Step[I]{}
	step.id = id
	step.replicas = replicas
	step.process = process
	return step
}

func NewStepWithErrorHandler[I any](id string, replicas uint16, process StepProcess[I], reportErrorHandler ReportError) *Step[I] {
	step := NewStep(id, replicas, process)
	step.reportError = reportErrorHandler
	return step
}

// run is a method that runs the step process and will be executed in a separate goroutine.
func (s *Step[I]) run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if ok {
				o, err := s.process(i)
				if err != nil {
					// since we will not proceed with the current token, we need to decrement the tokens count.
					s.decrementTokensCount()
					if s.reportError != nil {
						s.reportError(s.id, err)
					}
				} else {
					s.output <- o
				}
			}
		}
	}
}
