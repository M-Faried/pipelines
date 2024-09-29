package pipelines

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStepStandard_SuccessfullProcess(t *testing.T) {

	errorHandler := &mockErrorHandler{}
	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}

	// Create a stepStandard instance
	step := &stepBasic[int]{
		stepBase: stepBase[int]{
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		errorHandler: errorHandler.Handle,
		// Define a simple process function that just returns the input as output
		process: func(input int) (int, error) {
			return input, nil
		},
	}

	// Create a context with a timeout to ensure the test doesn't run indefinitely
	ctx, cancel := context.WithCancel(context.Background())

	// Create a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Run the step in a separate goroutine
	go step.Run(ctx, &wg)

	// Send an input value to the step
	step.input <- 42

	// Check the output value
	select {
	case output := <-step.output:
		if output != 42 {
			t.Errorf("expected output 42, got %d", output)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for output")
	}

	// Check that the error handler was not called
	if errorHandler.called {
		t.Errorf("did not expect error handler to be called")
	}

	// Check that the decrement handler was not called
	if decrementHandler.called {
		t.Errorf("did not expect decrement handler to be called")
	}
	if incrementHandler.called {
		t.Error("did not expect increment handler to be called")
	}
	if incrementHandler.counter != 0 {
		t.Errorf("expected value 0, got %d", incrementHandler.counter)
	}

	// Wait for the goroutine to finish
	cancel()
	wg.Wait()

	close(step.input)
	close(step.output)
}

func TestStepStandard_ProcessWithError(t *testing.T) {

	errorHandler := &mockErrorHandler{}
	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}

	// Define a simple process function that just returns the input as output
	process := func(input int) (int, error) {
		return 0, fmt.Errorf("test error")
	}

	// Create a stepStandard instance
	step := &stepBasic[int]{
		stepBase: stepBase[int]{
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		errorHandler: errorHandler.Handle,
		process:      process,
	}

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Create a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Run the step in a separate goroutine
	go step.Run(ctx, &wg)

	// Send an input value to the step
	step.input <- 42

	// Check the output value
	select {
	case output := <-step.output:
		t.Error("expected error but got an output", output)
	case <-time.After(1 * time.Second):
		// do nothing the timout is expected
	}

	if !errorHandler.called {
		t.Error("expected error handler to be called")
	}
	if errorHandler.err == nil || errorHandler.err.Error() != "test error" {
		t.Errorf("expected error 'test error', got %v", errorHandler.err)
	}
	if !decrementHandler.called {
		t.Error("expected decrement handler to be called")
	}
	if decrementHandler.counter != -1 {
		t.Errorf("expected value -1, got %d", decrementHandler.counter)
	}
	if incrementHandler.called {
		t.Error("did not expect increment handler to be called")
	}
	if incrementHandler.counter != 0 {
		t.Errorf("expected value 0, got %d", incrementHandler.counter)
	}

	cancel()
	wg.Wait()
	close(step.input)
	close(step.output)
}

func TestStepStandard_ClosingChannelShouldTerminateTheStep(t *testing.T) {

	errorHandler := &mockErrorHandler{}
	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}

	// Create a stepStandard instance
	step := &stepBasic[int]{
		stepBase: stepBase[int]{
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		errorHandler: errorHandler.Handle,
		// Define a simple process function that just returns the input as output
		process: func(input int) (int, error) {
			return input, nil
		},
	}

	// Create a context with a timeout to ensure the test doesn't run indefinitely
	ctx := context.Background()

	// Create a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Run the step in a separate goroutine
	go step.Run(ctx, &wg)

	// Send an input value to the step
	step.input <- 42

	// Check the output value
	select {
	case output := <-step.output:
		if output != 42 {
			t.Errorf("expected output 42, got %d", output)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for output")
	}

	close(step.input)
	close(step.output)
	before := time.Now()
	wg.Wait()
	after := time.Now()

	if after.Sub(before) > 10*time.Millisecond {
		t.Error("expected step to stop immediately after context is cancelled")
	}
}
