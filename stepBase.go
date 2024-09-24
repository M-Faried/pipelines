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

type InternalStep[I any] interface {
	IStep[I]
	setOuputChannel(chan I)
	getOuputChannel() chan I

	setInputChannel(chan I)
	getInputChannel() chan I

	getReplicas() uint16
	setIncrementTokensCountHandler(func())
	setDecrementTokensCountHandler(func())
	setReportErrorHanler(ReportError)
	run(context.Context, *sync.WaitGroup)
}

func castToInternalSteps[I any](step []IStep[I]) []InternalStep[I] {
	internalSteps := make([]InternalStep[I], len(step))
	for i, s := range step {
		internalSteps[i] = s.(InternalStep[I])
	}
	return internalSteps
}

// Step is a base struct for all steps
type Step[I any] struct {

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

func (s *Step[I]) GetLabel() string {
	return s.label
}

func (s *Step[I]) setInputChannel(input chan I) {
	s.input = input
}

func (s *Step[I]) getInputChannel() chan I {
	return s.input
}

func (s *Step[I]) setOuputChannel(output chan I) {
	s.output = output
}

func (s *Step[I]) getOuputChannel() chan I {
	return s.output
}

func (s *Step[I]) getReplicas() uint16 {
	return s.replicas
}

func (s *Step[I]) setDecrementTokensCountHandler(handler func()) {
	s.decrementTokensCount = handler
}

func (s *Step[I]) setIncrementTokensCountHandler(handler func()) {
	s.incrementTokensCount = handler
}

func (s *Step[I]) setReportErrorHanler(handler ReportError) {
	s.reportError = handler
}
