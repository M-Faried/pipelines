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
	errorHandler := &mockErrorHandler{}
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
			errorHandler:         errorHandler.Handle,
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
