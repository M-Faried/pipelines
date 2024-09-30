package pipelines

import (
	"context"
	"sync"
)

// IStep is an interface for all steps of the pipeline.
type IStep[I any] interface {

	// GetLabel returns the label of the step.
	GetLabel() string

	// GetReplicas returns the number of replicas of the step.
	GetReplicas() uint16
}

type iStepInternal[I any] interface {
	IStep[I]

	SetInputChannelSize(uint16)
	GetInputChannelSize() uint16

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
