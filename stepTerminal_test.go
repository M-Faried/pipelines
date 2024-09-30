package pipelines

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStepTerminal_SuccessfulProcess(t *testing.T) {

	processHandler := &mockTerminalProcessHandler[int]{}
	decrementTokensHandler := &mockDecrementTokensHandler{}
	incrementTokensHandler := &mockIncrementTokensHandler{}

	step := &stepTerminal[int]{
		stepBase: stepBase[int]{
			label:                "testStep",
			input:                make(chan int, 1),
			decrementTokensCount: decrementTokensHandler.Handle,
			incrementTokensCount: incrementTokensHandler.Handle,
		},
		process: processHandler.Handle,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go step.Run(ctx, wg)

	step.input <- 42

	time.Sleep(100 * time.Millisecond) // Give some time for the goroutine to process

	if !processHandler.called {
		t.Errorf("expected process handler to be called")
	}

	if !decrementTokensHandler.called {
		t.Errorf("expected decrement tokens handler to be called")
	}

	if incrementTokensHandler.called {
		t.Errorf("did not expect increment tokens handler to be called")
	}

	cancel()
	wg.Wait()
	close(step.input)
}

func TestStepTerminal_ClosingInputChannel(t *testing.T) {
	processHandler := &mockTerminalProcessHandler[int]{}
	decrementTokensHandler := &mockDecrementTokensHandler{}

	step := &stepTerminal[int]{
		stepBase: stepBase[int]{
			label:                "testStep",
			input:                make(chan int, 1),
			decrementTokensCount: decrementTokensHandler.Handle,
		},
		process: processHandler.Handle,
	}

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go step.Run(ctx, wg)

	step.input <- 42

	close(step.input)
	before := time.Now()
	wg.Wait()
	after := time.Now()

	if after.Sub(before) > 10*time.Millisecond {
		t.Error("expected step to stop immediately after context is cancelled")
	}
}

func TestStepTerminal_NewStep(t *testing.T) {

	process := func(input int) {}

	// Test with StepConfig
	stepConfig := StepTerminalConfig[int]{
		Label:            "fragmenterStep",
		InputChannelSize: 10,
		Replicas:         5,
		Process:          process,
	}

	step := newStepTerminal(stepConfig)
	var concreteStep *stepTerminal[int]
	var ok bool
	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepTerminal[int]); !ok {
		t.Error("Expected step to be of type stepTerminal")
	}

	if step.GetLabel() != "fragmenterStep" {
		t.Errorf("Expected label to be 'fragmenterStep', got '%s'", step.GetLabel())
	}

	if step.GetReplicas() != 5 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.process == nil {
		t.Error("Expected process to be set, got nil")
	}
	if concreteStep.inputChannelSize != 10 {
		t.Errorf("Expected input channel size to be 10, got %d", concreteStep.inputChannelSize)
	}
}

func TestStepTerminal_NewStep_MissingProcess(t *testing.T) {

	// Test with StepConfig
	stepConfig := StepTerminalConfig[int]{
		Label:    "testStep",
		Replicas: 1,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	newStepTerminal(stepConfig)
}
