package pipelines

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStepBasic_SuccessfullProcess(t *testing.T) {

	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}

	// Create a basic step instance
	step := &stepBasic[int]{
		stepBase: stepBase[int]{
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		// Define a simple process function that just returns the input as output
		process: func(input int) int {
			return input
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

func TestStepBasic_ClosingChannelShouldTerminateTheStep(t *testing.T) {

	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}

	// Create a stepBasic instance
	step := &stepBasic[int]{
		stepBase: stepBase[int]{
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		// Define a simple process function that just returns the input as output
		process: func(input int) int {
			return input
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

func TestStepBasic_NewStep(t *testing.T) {

	process := func(input int) int { return input }

	// Test with StepConfig
	stepConfig := StepBasicConfig[int]{
		Label:            "testStep",
		Replicas:         0, //should be rectified to 1
		Process:          process,
		InputChannelSize: 10,
	}

	step := newStepBasic(stepConfig)
	var concreteStep *stepBasic[int]
	var ok bool
	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepBasic[int]); !ok {
		t.Error("Expected step to be of type stepBasic")
	}

	if step.GetLabel() != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", step.GetLabel())
	}

	if step.GetReplicas() != 1 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.label != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", concreteStep.label)
	}
	if concreteStep.replicas != 1 {
		t.Errorf("Expected replicas to be 1, got %d", concreteStep.replicas)
	}
	if concreteStep.process == nil {
		t.Error("Expected process to be set, got nil")
	}
	if concreteStep.inputChannelSize != 10 {
		t.Errorf("Expected input channel size to be 10, got %d", concreteStep.inputChannelSize)
	}
}

func TestStepBasic_NewStep_MissingProcess(t *testing.T) {

	// Test with StepConfig
	stepConfig := StepBasicConfig[int]{
		Label:    "testStep",
		Replicas: 1,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	newStepBasic(stepConfig)
}
