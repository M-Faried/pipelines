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

type internalStep[I any] interface {
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

// stepBase is a base struct for all steps
type stepBase[I any] struct {

	// label is an label for the step set by the user.
	label string

	// input is a channel for incoming data to the step.
	input chan I

	// replicas is a number of goroutines that will be running the step.
	replicas uint16

	// reportError is the function called when an error occurs in the step.
	reportError ReportError

	// decrementTokensCount is a function that decrements the number of tokens in the pipeline.
	decrementTokensCount func()
}

func (s *stepBase[I]) GetLabel() string {
	return s.label
}

func (s *stepBase[I]) setInputChannel(input chan I) {
	s.input = input
}

func (s *stepBase[I]) getInputChannel() chan I {
	return s.input
}

func (s *stepBase[I]) setDecrementTokensCountHandler(handler func()) {
	s.decrementTokensCount = handler
}

func (s *stepBase[I]) setReportErrorHanler(handler ReportError) {
	s.reportError = handler
}

func (s *stepBase[I]) getReplicas() uint16 {
	return s.replicas
}
