package pipelines

import (
	"context"
	"fmt"
	"sync"
)

// mockErrorHandler is a mock implementation of the error handler
type mockErrorHandler struct {
	called bool
	label  string
	err    error
}

func (m *mockErrorHandler) Handle(label string, err error) {
	m.called = true
	m.label = label
	m.err = err
}

// mockResultProcessHandler is a mock implementation of the result process handler
type mockResultProcessHandler[I any] struct {
	called bool
	input  I
}

func (m *mockResultProcessHandler[I]) Handle(input I) {
	m.called = true
	m.input = input
}

// mockDecrementTokensHandler is a mock implementation of the decrement tokens handler
type mockDecrementTokensHandler struct {
	called  bool
	counter int
}

func (m *mockDecrementTokensHandler) Handle() {
	m.called = true
	m.counter--
}

// mockIncrementTokensHandler is a mock implementation of the increment tokens handler
type mockIncrementTokensHandler struct {
	called  bool
	counter int
}

func (m *mockIncrementTokensHandler) Handle() {
	m.called = true
	m.counter++
}

// mockStep is a mock implementation of the step
type mockStep[I any] struct {
	label            string
	inputChannel     chan I
	outputChannel    chan I
	replicas         uint16
	incrementHandler func()
	decrementHandler func()
	errorHandler     *mockErrorHandler
	finalStep        bool
	isDuplex         bool
}

func (m *mockStep[I]) GetLabel() string {
	return m.label
}

func (m *mockStep[I]) SetInputChannel(ch chan I) {
	fmt.Println("setting input channel", ch)
	m.inputChannel = ch
}

func (m *mockStep[I]) GetInputChannel() chan I {
	return m.inputChannel
}

func (m *mockStep[I]) SetOutputChannel(ch chan I) {
	m.outputChannel = ch
}

func (m *mockStep[I]) GetOutputChannel() chan I {
	return m.outputChannel
}

func (m *mockStep[I]) GetReplicas() uint16 {
	return m.replicas
}

func (m *mockStep[I]) SetDecrementTokensCountHandler(handler func()) {
	m.decrementHandler = handler
}

func (m *mockStep[I]) SetIncrementTokensCountHandler(handler func()) {
	m.incrementHandler = handler
}

func (m *mockStep[I]) SetErrorHandler(handler *mockErrorHandler) {
	m.errorHandler = handler
}

func (m *mockStep[I]) SetIsDuplex(reverse bool) {
	m.isDuplex = reverse
}

func (m *mockStep[I]) GetIsDuplex() bool {
	return m.isDuplex
}

func (m *mockStep[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case item, ok := <-m.inputChannel:
			if !ok {
				wg.Done()
				return
			}
			if m.finalStep {
				m.decrementHandler()
			} else {
				m.outputChannel <- item
			}
		}
	}
}

func (m *mockStep[I]) RunReverse(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-m.outputChannel:
			if !ok {
				return
			}
			if m.finalStep {
				m.decrementHandler()
			} else {
				m.inputChannel <- item
			}
		}
	}
}
