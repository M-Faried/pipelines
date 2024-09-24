package pipelines

import (
	"context"
	"sync"
)

type IStep[I any] interface {
	GetLabel() string
	GetReplicas() uint16
}

type iStepInternal[I any] interface {
	IStep[I]

	SetInputChannel(chan I)
	GetInputChannel() chan I

	SetOutputChannel(chan I)
	GetOutputChannel() chan I

	SetIncrementTokensCountHandler(func())
	SetDecrementTokensCountHandler(func())

	Run(context.Context, *sync.WaitGroup)
}

func castToInternalSteps[I any](step []IStep[I]) []iStepInternal[I] {
	internalSteps := make([]iStepInternal[I], len(step))
	for i, s := range step {
		internalSteps[i] = s.(iStepInternal[I])
	}
	return internalSteps
}
