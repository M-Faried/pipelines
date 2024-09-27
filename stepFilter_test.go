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
	if decrementHandler.value != -5 {
		t.Errorf("expected value -5, got %d", decrementHandler.value)
	}
}
