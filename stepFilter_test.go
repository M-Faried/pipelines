package pipelines

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStepFilter_Run(t *testing.T) {

	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}

	step := &stepFilter[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		passCriteria: func(i int) bool {
			return i%2 == 0
		},
	}

	// Create a context with a timeout to ensure the test doesn't run indefinitely
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Run the step in a separate goroutine
	go step.Run(ctx, &wg)

	go func() {
		for i := 0; i < 10; i++ {
			step.input <- i
		}
		// giving time for the receiver to process the input
		time.Sleep(100 * time.Millisecond)
		close(step.input)
	}()

	var results []int
	go func() {
		for {
			select {
			case res := <-step.output:
				results = append(results, res)
				if len(results) == 5 {
					return
				}
			case <-time.After(1 * time.Second):
				t.Error("timeout waiting for output")
				return
			}
		}
	}()

	wg.Wait()
	close(step.output)

	expectedResults := []int{0, 2, 4, 6, 8}
	if len(results) != len(expectedResults) {
		t.Fatalf("expected %v results, got %v", len(expectedResults), len(results))
	}
	for i, res := range results {
		if res != expectedResults[i] {
			t.Errorf("expected result %v at index %v, got %v", expectedResults[i], i, res)
		}
	}
	if incrementHandler.called {
		t.Error("did not expect increment handler to be called")
	}
	if !decrementHandler.called {
		t.Errorf("expect decrement handler to be called")
	}
	if decrementHandler.counter != -5 {
		t.Errorf("expected value -5, got %d", decrementHandler.counter)
	}
}

func TestStepFilter_ClosingWithParentContext(t *testing.T) {

	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}

	step := &stepFilter[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		passCriteria: func(i int) bool {
			return i%2 == 0
		},
	}

	// Create a context with a timeout to ensure the test doesn't run indefinitely
	ctx, cancelCtx := context.WithCancel(context.Background())

	// Create a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Run the step in a separate goroutine
	go step.Run(ctx, &wg)

	// step.input <- 42

	time.Sleep(100 * time.Millisecond)

	cancelCtx()
	before := time.Now()
	wg.Wait()
	after := time.Now()

	if after.Sub(before) > 1*time.Second {
		t.Error("expected step to stop immediately after context is cancelled")
	}

	close(step.input)
	close(step.output)
}

func TestStepFilter_NewStep(t *testing.T) {

	process := func(input int) bool { return true }

	// Test with StepConfig
	stepConfig := StepFilterConfig[int]{
		Label:            "testStep",
		Replicas:         2,
		InputChannelSize: 10,
		PassCriteria:     process,
	}

	step := newStepFilter(stepConfig)

	var concreteStep *stepFilter[int]
	var ok bool

	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepFilter[int]); !ok {
		t.Error("Expected step to be of type stepFilter")
	}

	if step.GetLabel() != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", step.GetLabel())
	}
	if step.GetReplicas() != 2 {
		t.Errorf("Expected replicas to be 2, got %d", step.GetReplicas())
	}

	if concreteStep.label != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", concreteStep.label)
	}
	if concreteStep.replicas != 2 {
		t.Errorf("Expected replicas to be 2, got %d", concreteStep.replicas)
	}
	if concreteStep.passCriteria == nil {
		t.Error("Expected process to be set, got nil")
	}
	if concreteStep.inputChannelSize != 10 {
		t.Errorf("Expected input channel size to be 10, got %d", concreteStep.inputChannelSize)
	}
}

func TestStepFilter_NewStep_MissingProcess(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := StepFilterConfig[int]{
		Label:    "testStep",
		Replicas: 1,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
}
