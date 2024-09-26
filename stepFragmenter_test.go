package pipelines

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestStepFragmenter_SuccessfulProcess(t *testing.T) {
	errorHandler := &mockErrorHandler{}
	decrementTokens := &mockDecrementTokensHandler{}
	incrementTokens := &mockIncrementTokensHandler{}

	process := func(input int) ([]int, error) {
		res := make([]int, input)
		for i := 0; i < input; i++ {
			res[i] = i
		}
		return res, nil
	}

	step := &stepFragmenter[int]{
		stepBase: stepBase[int]{
			label:                "testFragmenter",
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			errorHandler:         errorHandler.Handle,
			decrementTokensCount: decrementTokens.Handle,
			incrementTokensCount: incrementTokens.Handle,
		},
		process: process,
	}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go step.Run(ctx, &wg)

	step.input <- 42
	done := false
	for !done {
		select {
		case out := <-step.output:
			if out == 41 {
				done = true
			}
		case <-time.After(2 * time.Second):
			t.Error("timeout waiting for output")
			done = true // Exit the loop on timeout
		}
	}

	time.Sleep(100 * time.Millisecond) // Give some time for the goroutine to process

	if !decrementTokens.called {
		t.Error("expected decrement handler to be called")
	}
	if decrementTokens.value != -1 {
		t.Errorf("expected value -1, got %d", decrementTokens.value)
	}
	if !incrementTokens.called {
		t.Error("expected increment handler to be called")
	}
	if incrementTokens.value != 42 {
		t.Errorf("expected value 42, got %d", incrementTokens.value)
	}
	if errorHandler.called {
		t.Errorf("expected error handler not to be called")
	}

	cancel()
	wg.Wait()
	close(step.input)
	close(step.output)
}

func TestStepFragmenter_ProcessWithError(t *testing.T) {
	mockHandler := &mockErrorHandler{}
	decrementTokens := &mockDecrementTokensHandler{}

	process := func(input int) ([]int, error) {
		return nil, errors.New("process error")
	}

	step := &stepFragmenter[int]{
		stepBase: stepBase[int]{
			label:                "testFragmenter",
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			errorHandler:         mockHandler.Handle,
			decrementTokensCount: decrementTokens.Handle,
		},
		process: process,
	}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go step.Run(ctx, &wg)

	step.input <- 42

	// Give some time for the goroutine to process
	time.Sleep(100 * time.Millisecond)

	if !mockHandler.called {
		t.Errorf("expected error handler to be called")
	}
	if mockHandler.label != "testFragmenter" {
		t.Errorf("expected label %s, got %s", "testFragmenter", mockHandler.label)
	}
	if mockHandler.err == nil || mockHandler.err.Error() != "process error" {
		t.Errorf("expected error 'process error', got %v", mockHandler.err)
	}
	if !decrementTokens.called {
		t.Error("expected decrement handler to be called")
	}
	if decrementTokens.value != -1 {
		t.Errorf("expected value -1, got %d", decrementTokens.value)
	}

	cancel()
	wg.Wait()
	close(step.input)
	close(step.output)
}
