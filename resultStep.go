package pipelines

import (
	"context"
	"fmt"
	"sync"
)

type ResultStepProcess[I any] func(I) error

type ResultStep[I any] struct {
	baseStep[I]
	decrementTokensCount func()
	process              ResultStepProcess[I]
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
				if err != nil {
					wrappedErr := fmt.Errorf("error in %s: %w", s.id, err)
					s.errorsQueue.Enqueue(wrappedErr)
				}
				s.decrementTokensCount()
			}
		}
	}
}

func NewResultStep[I any](id string, replicas uint8, process ResultStepProcess[I]) *ResultStep[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := &ResultStep[I]{}
	step.id = id
	step.replicas = replicas
	step.process = process
	return step
}
