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
