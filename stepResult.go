package pipelines

import (
	"context"
	"sync"
)

// StepResultProcess is a function that processes the input data and does not return any data.
type StepResultProcess[I any] func(I) error

// StepResult is a struct that represents a step in the pipeline that does not return any data.
type StepResult[I any] struct {
	stepBase[I]
	process StepResultProcess[I]
}

// NewStepResult creates a new result step with the given label, number of replicas and process.
func NewStepResult[I any](label string, replicas uint16, process StepResultProcess[I]) IStep[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := &StepResult[I]{}
	step.label = label
	step.replicas = replicas
	step.process = process
	return step
}

// NewStepResultWithErrorHandler creates a new result step with the given label, number of replicas, process and error handler.
func NewStepResultWithErrorHandler[I any](label string, replicas uint16, process StepResultProcess[I], reportErrorHandler ReportError) IStep[I] {
	step := NewStepResult(label, replicas, process)
	step.setReportErrorHanler(reportErrorHandler)
	return step
}

func (s *StepResult[I]) run(ctx context.Context, wg *sync.WaitGroup) {
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

func (s *StepResult[I]) setOuputChannel(output chan I) {
	panic("Not Supported Operation!!!")
}

func (s *StepResult[I]) getOuputChannel() chan I {
	panic("Not Supported Operation!!!")
}

func (s *StepResult[I]) setIncrementTokensCountHandler(handler func()) {
	panic("Not Supported Operation!!!")
}
