package pipelines

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestStepResult_SuccessfulProcess(t *testing.T) {

	errorHandler := &mockErrorHandler{}
	processHandler := &mockResultProcessHandler[int]{}
	decrementTokensHandler := &mockDecrementTokensHandler{}
	incrementTokensHandler := &mockIncrementTokensHandler{}

	step := &stepResult[int]{
		stepBase: stepBase[int]{
			label:                "testStep",
			input:                make(chan int, 1),
			errorHandler:         errorHandler.Handle,
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

	if errorHandler.called {
		t.Errorf("did not expect error handler to be called")
	}

	cancel()
	wg.Wait()
	close(step.input)
}

func TestStepResult_ProcessWithError(t *testing.T) {

	decrementHandler := &mockDecrementTokensHandler{}
	incrementHandler := &mockIncrementTokensHandler{}
	errorHandler := &mockErrorHandler{}

	step := &stepResult[int]{
		stepBase: stepBase[int]{
			label:                "testStep",
			input:                make(chan int, 3),
			errorHandler:         errorHandler.Handle,
			decrementTokensCount: decrementHandler.Handle,
			incrementTokensCount: incrementHandler.Handle,
		},
		process: func(i int) error {
			if i == 2 {
				return errors.New("process error")
			}
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go step.Run(ctx, wg)

	step.input <- 1
	step.input <- 2
	step.input <- 3

	time.Sleep(100 * time.Millisecond) // Give some time for the goroutine to process

	if !errorHandler.called {
		t.Errorf("expected error handler to be called")
	}
	if errorHandler.label != "testStep" {
		t.Errorf("expected label %v, got %v", "testStep", errorHandler.label)
	}
	if errorHandler.err == nil || errorHandler.err.Error() != "process error" {
		t.Errorf("expected error %v, got %v", "process error", errorHandler.err)
	}
	if !decrementHandler.called {
		t.Errorf("expected decrement handler to be called")
	}
	if incrementHandler.called {
		t.Errorf("did not expect increment handler to be called")
	}

	cancel()
	wg.Wait()
	close(step.input)
}

func TestStepResult_ClosingInputChannel(t *testing.T) {

	errorHandler := &mockErrorHandler{}
	processHandler := &mockResultProcessHandler[int]{}
	decrementTokensHandler := &mockDecrementTokensHandler{}

	step := &stepResult[int]{
		stepBase: stepBase[int]{
			label:                "testStep",
			input:                make(chan int, 1),
			errorHandler:         errorHandler.Handle,
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
	before := time.Now().Unix()
	wg.Wait()
	after := time.Now().Unix()

	if after-before > 1 {
		t.Errorf("expected less than 1 second, got %d", after-before)
	}
}
