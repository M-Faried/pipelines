package pipelines

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStepBuffered_Run_InputTriggered(t *testing.T) {

	incrementTokens := &mockIncrementTokensHandler{}
	decrementTokens := &mockDecrementTokensHandler{}

	step := &stepBuffered[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			incrementTokensCount: incrementTokens.Handle,
			decrementTokensCount: decrementTokens.Handle,
		},
		bufferSize:  3,
		buffer:      make([]int, 0, 3),
		passThrough: true,
		inputTriggeredProcess: func(buffer []int) StepBufferedProcessOutput[int] {
			fmt.Println("inputTriggeredProcess", buffer)
			return StepBufferedProcessOutput[int]{
				HasResult:   false,
				Result:      0,
				FlushBuffer: false,
			}
		},
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go step.Run(ctx, &wg)

	go func() {
		step.input <- 1
		step.input <- 2
		step.input <- 3
		step.input <- 4
	}()

	var results []int
	go func() {
		for res := range step.output {
			results = append(results, res)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	close(step.input)
	close(step.output)
	cancelCtx()
	wg.Wait()

	expectedBuffer := []int{2, 3, 4}
	if !equal(step.buffer, expectedBuffer) {
		t.Errorf("expected %v, got %v", expectedBuffer, step.buffer)
	}

	expectedResults := []int{1, 2, 3, 4}
	if !equal(results, expectedResults) {
		t.Errorf("expected %v, got %v", expectedResults, results)
	}

	if !incrementTokens.called {
		t.Errorf("expected incrementTokens to be called")
	}

	if incrementTokens.counter != 3 {
		t.Errorf("expected incrementTokens to be called 3 times, got %d", incrementTokens.counter)
	}

	if decrementTokens.called {
		t.Errorf("expected decrementTokens to not be called")
	}
}

func TestStepBuffered_Run_InputTriggered_FlushBuffer(t *testing.T) {
	incrementTokens := &mockIncrementTokensHandler{}
	decrementTokens := &mockDecrementTokensHandler{}

	step := &stepBuffered[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			incrementTokensCount: incrementTokens.Handle,
			decrementTokensCount: decrementTokens.Handle,
		},
		bufferSize:  3,
		passThrough: false,
		inputTriggeredProcess: func(buffer []int) StepBufferedProcessOutput[int] {
			return StepBufferedProcessOutput[int]{
				HasResult:   true,
				Result:      buffer[len(buffer)-1] * 2,
				FlushBuffer: true,
			}
		},
		timeTriggeredProcessInterval: 50 * time.Millisecond,
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go step.Run(ctx, &wg)

	go func() {
		step.input <- 1
		step.input <- 2
		step.input <- 3
	}()

	var results []int
	go func() {
		for result := range step.output {
			results = append(results, result)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	cancelCtx()
	wg.Wait()
	close(step.input)
	close(step.output)

	expected := []int{2, 4, 6}
	if !equal(results, expected) {
		t.Errorf("expected %v, got %v", expected, results)
	}
	if len(step.buffer) != 0 {
		t.Errorf("expected buffer to be empty, got %v", step.buffer)
	}

	if !incrementTokens.called {
		t.Errorf("expected incrementTokens to be called")
	}

	if incrementTokens.counter != 3 {
		t.Errorf("expected incrementTokens to be called 3 times, got %d", incrementTokens.counter)
	}

	if !decrementTokens.called {
		t.Errorf("expected decrementTokens to be called")
	}

	if decrementTokens.counter != -3 {
		t.Errorf("expected decrementTokens to be called 3 times, got %d", decrementTokens.counter)
	}
}

func TestStepBuffered_HandleTimeTriggeredProcess_FlushBuffer(t *testing.T) {

	incrementTokens := &mockIncrementTokensHandler{}
	decrementTokens := &mockDecrementTokensHandler{}

	step := &stepBuffered[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			incrementTokensCount: incrementTokens.Handle,
			decrementTokensCount: decrementTokens.Handle,
		},
		bufferSize:  3,
		passThrough: false,
		timeTriggeredProcess: func(buffer []int) StepBufferedProcessOutput[int] {
			if len(buffer) < 3 {
				return StepBufferedProcessOutput[int]{
					HasResult: false,
					Result:    0,
				}
			}
			sum := 0
			for _, v := range buffer {
				sum += v
			}
			return StepBufferedProcessOutput[int]{
				HasResult:   true,
				Result:      sum,
				FlushBuffer: true,
			}
		},
		timeTriggeredProcessInterval: 100 * time.Millisecond,
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go step.Run(ctx, &wg)

	go func() {
		step.input <- 1
		step.input <- 2
		step.input <- 3
	}()

	var results []int
	go func() {
		for result := range step.output {
			results = append(results, result)
		}
	}()

	time.Sleep(120 * time.Millisecond)

	cancelCtx()
	close(step.input)
	close(step.output)
	wg.Wait()

	expected := []int{6}
	if !equal(results, expected) {
		t.Errorf("expected %v, got %v", expected, results)
	}

	expectedBuffer := []int{}
	if !equal(step.buffer, expectedBuffer) {
		t.Errorf("expected %v, got %v", expectedBuffer, step.buffer)
	}

	if !incrementTokens.called {
		t.Errorf("expected incrementTokens to be called")
	}

	if incrementTokens.counter != 1 {
		t.Errorf("expected incrementTokens to be called 1 time, got %d", incrementTokens.counter)
	}

	if !decrementTokens.called {
		t.Errorf("expected decrementTokens to be called")
	}

	if decrementTokens.counter != -3 {
		t.Errorf("expected decrementTokens to be called 3 times, got %d", decrementTokens.counter)
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
