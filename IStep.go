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

	// Sets the input channel buffer size. Used by the pipeline during init.
	SetInputChannelSize(uint16)

	// Gets the input channel buffer size.
	GetInputChannelSize() uint16

	// Sets the input channel. Used by the pipeline during init.
	SetInputChannel(chan I)

	// Gets the input channel.
	GetInputChannel() chan I

	// Sets the output channel. Used by the pipeline during init.
	SetOutputChannel(chan I)

	// Gets the output channel.
	GetOutputChannel() chan I

	// Sets the handler for inclrementing the pipline's tokens count by 1. Used by the pipeline during init.
	SetIncrementTokensCountHandler(func())

	// Sets the handler for inclrementing the pipline's tokens count by 1. Used by the pipeline during init.
	SetDecrementTokensCountHandler(func())

	// Runs the listeners to input channels. Used by the pipeline during Run.
	Run(context.Context, *sync.WaitGroup)
}
