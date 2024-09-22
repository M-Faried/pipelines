package pipelines

import (
	"context"
	"fmt"
	"sync"
)

type StepProcess[I any] func(I) (I, error)

type Step[I any] struct {
	baseStep[I]
	// ouput is a channel for outgoing data from the step.
	output chan I
	// process is a function that will be applied to the incoming data.
	process StepProcess[I]
	// errorsQueue is a queue for errors that may be reported during the step execution.
	errorsQueue *Queue[error]
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
					wrappedErr := fmt.Errorf("error in %s: %w", s.id, err)
					s.errorsQueue.Enqueue(wrappedErr)
				} else {
					s.output <- o
				}
			}
		}
	}
}

// NewStep is a constructor for the Step struct.
func NewStep[I any](id string, replicas uint8, process StepProcess[I]) *Step[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := &Step[I]{}
	step.id = id
	step.replicas = replicas
	step.process = process
	return step
}
