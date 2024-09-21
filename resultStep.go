package pipelines

import (
	"context"
	"sync"
)

type ResultStepProcess[I any] func(I)

type ResultStep[I any] struct {
	baseStep[I]
	process ResultStepProcess[I]
}

func (s *ResultStep[I]) run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if ok {
				s.process(i)
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
