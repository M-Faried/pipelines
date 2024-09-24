package pipelines

import (
	"context"
	"sync"
)

// StepProcess is a function that processes the input data and returns the output data.
type StepProcess[I any] func(I) (I, error)

type stepStandard[I any] struct {
	Step[I]
	// process is a function that will be applied to the incoming data.
	process StepProcess[I]
}

// NewStep creates a new step with the given label, number of replicas and process.
func NewStep[I any](label string, replicas uint16, process StepProcess[I]) IStep[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := &stepStandard[I]{}
	step.label = label
	step.replicas = replicas
	step.process = process
	return step
}

// NewStepWithErrorHandler creates a new step with the given label, number of replicas, process and error handler.
func NewStepWithErrorHandler[I any](label string, replicas uint16, process StepProcess[I], reportErrorHandler ReportError) IStep[I] {
	step := NewStep(label, replicas, process).(*stepStandard[I])
	step.setReportErrorHanler(reportErrorHandler)
	return step
}

// run is a method that runs the step process and will be executed in a separate goroutine.
func (s *stepStandard[I]) run(ctx context.Context, wg *sync.WaitGroup) {
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
						s.reportError(s.label, err)
					}
				} else {
					s.output <- o
				}
			}
		}
	}
}
