package pipelines

import (
	"context"
	"sync"
)

// ReportError is the definition of error reporting handler which may or may not be set by the user during creation of the step.
// The first parameter is the label of the step where the error occurred and the second parameter is the error itself.
type ReportError func(string, error)

type IStep[I any] interface {
	GetLabel() string
}

type iInternalStep[I any] interface {
	IStep[I]

	SetInputChannel(chan I)
	GetInputChannel() chan I

	SetOutputChannel(chan I)
	GetOutputChannel() chan I

	GetReplicas() uint16
	SetIncrementTokensCountHandler(func())
	SetDecrementTokensCountHandler(func())

	Run(context.Context, *sync.WaitGroup)
}

func castToInternalSteps[I any](step []IStep[I]) []iInternalStep[I] {
	internalSteps := make([]iInternalStep[I], len(step))
	for i, s := range step {
		internalSteps[i] = s.(iInternalStep[I])
	}
	return internalSteps
}

// step is a base struct for all steps
type step[I any] struct {

	// label is an label for the step set by the user.
	label string

	// input is a channel for incoming data to the step.
	input chan I

	// ouput is a channel for outgoing data from the step.
	output chan I

	// replicas is a number of goroutines that will be running the step.
	replicas uint16

	// reportError is the function called when an error occurs in the step.
	reportError ReportError

	// decrementTokensCount is a function that decrements the number of tokens in the pipeline.
	decrementTokensCount func()

	// incrementTokensCount is a function that increments the number of tokens in the pipeline.
	incrementTokensCount func()
}

func (s *step[I]) GetLabel() string {
	return s.label
}

func (s *step[I]) SetInputChannel(input chan I) {
	s.input = input
}

func (s *step[I]) GetInputChannel() chan I {
	return s.input
}

func (s *step[I]) SetOutputChannel(output chan I) {
	s.output = output
}

func (s *step[I]) GetOutputChannel() chan I {
	return s.output
}

func (s *step[I]) GetReplicas() uint16 {
	return s.replicas
}

func (s *step[I]) SetDecrementTokensCountHandler(handler func()) {
	s.decrementTokensCount = handler
}

func (s *step[I]) SetIncrementTokensCountHandler(handler func()) {
	s.incrementTokensCount = handler
}

func (s *step[I]) SetReportErrorHanler(handler ReportError) {
	s.reportError = handler
}
