package pipelines

import (
	"context"
	"sync"
)

// StepFragmenterProcess is a function that converts a token in the pipeline into multiple tokens.
type StepFragmenterProcess[I any] func(I) ([]I, error)

// StepFragmenterConfig is a struct that defines the configuration for a fragmenter step
type StepFragmenterConfig[I any] struct {
	Label        string
	Replicas     uint16
	ErrorHandler ErrorHandler
	Process      StepFragmenterProcess[I]
}

type stepFragmenter[I any] struct {
	stepBase[I]
	process StepFragmenterProcess[I]
}

func (s *stepFragmenter[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if !ok {
				return
			}
			outFragments, err := s.process(i)
			if err != nil {
				s.reportError(err)
			} else {
				for _, fragment := range outFragments {
					// adding fragmented tokens to the count.
					s.incrementTokensCount()
					s.output <- fragment
				}
			}
			// whether the token is framented or filtered with error, it is discarded from the pipeline.
			s.decrementTokensCount()
		}
	}
}
