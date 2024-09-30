package pipelines

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStepFragmenter_SuccessfulProcess(t *testing.T) {
	decrementTokens := &mockDecrementTokensHandler{}
	incrementTokens := &mockIncrementTokensHandler{}

	process := func(input int) []int {
		res := make([]int, input)
		for i := 0; i < input; i++ {
			res[i] = i
		}
		return res
	}

	step := &stepFragmenter[int]{
		stepBase: stepBase[int]{
			label:                "testFragmenter",
			input:                make(chan int, 1),
			output:               make(chan int, 1),
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
	if decrementTokens.counter != -1 {
		t.Errorf("expected value -1, got %d", decrementTokens.counter)
	}
	if !incrementTokens.called {
		t.Error("expected increment handler to be called")
	}
	if incrementTokens.counter != 42 {
		t.Errorf("expected value 42, got %d", incrementTokens.counter)
	}

	cancel()
	wg.Wait()
	close(step.input)
	close(step.output)
}

func TestStepFragmenter_ClosingChannelShouldTerminateTheStep(t *testing.T) {

	decrementTokens := &mockDecrementTokensHandler{}
	incrementTokens := &mockIncrementTokensHandler{}

	process := func(input int) []int {
		res := make([]int, input)
		for i := 0; i < input; i++ {
			res[i] = i
		}
		return res
	}

	step := &stepFragmenter[int]{
		stepBase: stepBase[int]{
			label:                "testFragmenter",
			input:                make(chan int, 1),
			output:               make(chan int, 1),
			decrementTokensCount: decrementTokens.Handle,
			incrementTokensCount: incrementTokens.Handle,
		},
		process: process,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go step.Run(ctx, &wg)

	close(step.input)

	before := time.Now()
	wg.Wait()
	after := time.Now()

	if after.Sub(before) > 1*time.Second {
		t.Error("expected step to stop immediately after context is cancelled")
	}
	close(step.output)
}

func TestStepFragmenter_NewStep(t *testing.T) {

	process := func(input int) []int { return []int{input} }

	// Test with StepConfig
	stepConfig := StepFragmenterConfig[int]{
		Label:            "fragmenterStep",
		Replicas:         2,
		Process:          process,
		InputChannelSize: 10,
	}

	step := newStepFragmenter(stepConfig)
	var concreteStep *stepFragmenter[int]
	var ok bool
	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepFragmenter[int]); !ok {
		t.Error("Expected step to be of type stepFragmenter")
	}

	if step.GetLabel() != "fragmenterStep" {
		t.Errorf("Expected label to be 'fragmenterStep', got '%s'", step.GetLabel())
	}
	if step.GetReplicas() != 2 {
		t.Errorf("Expected replicas to be 2, got %d", step.GetReplicas())
	}

	if concreteStep.label != "fragmenterStep" {
		t.Errorf("Expected label to be 'fragmenterStep', got '%s'", concreteStep.label)
	}
	if concreteStep.replicas != 2 {
		t.Errorf("Expected replicas to be 2, got %d", concreteStep.replicas)
	}
	if concreteStep.process == nil {
		t.Error("Expected process to be set, got nil")
	}
	if concreteStep.inputChannelSize != 10 {
		t.Errorf("Expected input channel size to be 10, got %d", concreteStep.inputChannelSize)
	}
}

func TestStepFragmenter_NewStep_MissingProcess(t *testing.T) {

	// Test with StepConfig
	stepConfig := StepFragmenterConfig[int]{
		Label:    "testStep",
		Replicas: 1,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	newStepFragmenter(stepConfig)
}
