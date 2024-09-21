package pipelines

import (
	"context"
	"fmt"
	"sync"
)

type StepProcess[I any] func(I) (I, error)

type Step[I any] struct {
	baseStep[I]
	output  chan I
	process StepProcess[I]
}

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
