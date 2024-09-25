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
