package pipelines

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStepBuffer_Run_InputTriggered(t *testing.T) {

	incrementTokens := &mockIncrementTokensHandler{}
	decrementTokens := &mockDecrementTokensHandler{}

	step := &stepBuffer[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			incrementTokensCount: incrementTokens.Handle,
			decrementTokensCount: decrementTokens.Handle,
		},
		bufferSize:  3,
		buffer:      make([]int, 0, 3),
		passThrough: true,
		inputTriggeredProcess: func(buffer []int) (int, BufferFlags) {
			fmt.Println("inputTriggeredProcess", buffer)
			return 0, BufferFlags{
				SendProcessOuput: false,
				FlushBuffer:      false,
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

func TestStepBuffer_Run_InputTriggered_FlushBuffer(t *testing.T) {
	incrementTokens := &mockIncrementTokensHandler{}
	decrementTokens := &mockDecrementTokensHandler{}

	step := &stepBuffer[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			incrementTokensCount: incrementTokens.Handle,
			decrementTokensCount: decrementTokens.Handle,
		},
		bufferSize:  3,
		passThrough: false,
		inputTriggeredProcess: func(buffer []int) (int, BufferFlags) {
			result := buffer[len(buffer)-1] * 2
			return result, BufferFlags{
				SendProcessOuput: true,
				FlushBuffer:      true,
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

func TestStepBuffer_HandleTimeTriggeredProcess_FlushBuffer(t *testing.T) {

	incrementTokens := &mockIncrementTokensHandler{}
	decrementTokens := &mockDecrementTokensHandler{}

	step := &stepBuffer[int]{
		stepBase: stepBase[int]{
			input:                make(chan int),
			output:               make(chan int),
			incrementTokensCount: incrementTokens.Handle,
			decrementTokensCount: decrementTokens.Handle,
		},
		bufferSize:  3,
		passThrough: false,
		timeTriggeredProcess: func(buffer []int) (int, BufferFlags) {
			if len(buffer) < 3 {
				return 0, BufferFlags{
					SendProcessOuput: false,
				}
			}
			sum := 0
			for _, v := range buffer {
				sum += v
			}
			return sum, BufferFlags{
				SendProcessOuput: true,
				FlushBuffer:      true,
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

func TestStepBuffer_NewStep(t *testing.T) {

	processInputTriggered := func(input []int) (int, BufferFlags) {
		return 0, BufferFlags{}
	}
	processTimeTriggered := func(input []int) (int, BufferFlags) {
		return 0, BufferFlags{}
	}

	// Test with StepConfig
	stepConfig := StepBufferConfig[int]{
		Label:                        "testStep",
		Replicas:                     1,
		BufferSize:                   20,
		InputChannelSize:             10,
		PassThrough:                  true,
		InputTriggeredProcess:        processInputTriggered,
		TimeTriggeredProcess:         processTimeTriggered,
		TimeTriggeredProcessInterval: 10 * time.Second,
	}

	step := newStepBuffer(stepConfig)

	var concreteStep *stepBuffer[int]
	var ok bool

	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepBuffer[int]); !ok {
		t.Error("Expected step to be of type stepBuffer")
	}

	if concreteStep.label != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", step.GetLabel())
	}

	if concreteStep.replicas != 1 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.timeTriggeredProcess == nil {
		t.Error("Expected time triggered process to be set, got nil")
	}

	if concreteStep.inputTriggeredProcess == nil {
		t.Error("Expected input triggered process to be set, got nil")
	}

	if concreteStep.timeTriggeredProcessInterval != 10*time.Second {
		t.Errorf("Expected time triggered process interval to be 10s, got %s", concreteStep.timeTriggeredProcessInterval)
	}

	if !concreteStep.passThrough {
		t.Error("Expected pass through to be true, got false")
	}

	if concreteStep.bufferSize != 20 {
		t.Errorf("Expected buffer size to be 20, got %d", concreteStep.bufferSize)
	}

	if cap(concreteStep.buffer) != 20 {
		t.Errorf("Expected buffer capacity to be 20, got %d", cap(concreteStep.buffer))
	}

	if concreteStep.inputChannelSize != 10 {
		t.Errorf("Expected input channel size to be 10, got %d", concreteStep.inputChannelSize)
	}
}

func TestStepBuffer_NewStep_InputTriggeredOnly(t *testing.T) {

	processInputTriggered := func(input []int) (int, BufferFlags) {
		return 0, BufferFlags{}
	}

	// Test with StepConfig
	stepConfig := StepBufferConfig[int]{
		Label:                 "testStep",
		Replicas:              1,
		BufferSize:            15,
		InputChannelSize:      5,
		InputTriggeredProcess: processInputTriggered,
	}

	step := newStepBuffer(stepConfig)

	var concreteStep *stepBuffer[int]
	var ok bool

	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepBuffer[int]); !ok {
		t.Error("Expected step to be of type stepBuffer")
	}

	if concreteStep.label != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", step.GetLabel())
	}

	if concreteStep.replicas != 1 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.inputTriggeredProcess == nil {
		t.Error("Expected input triggered process to be set, got nil")
	}

	if concreteStep.timeTriggeredProcess != nil {
		t.Errorf("Expected input triggered process to be nil, got %v", concreteStep.timeTriggeredProcess)
	}

	if concreteStep.timeTriggeredProcessInterval > 0 {
		t.Errorf("Expected time triggered process interval to be 0s, got %s", concreteStep.timeTriggeredProcessInterval)
	}

	if concreteStep.passThrough {
		t.Error("Expected pass through to be false, got true")
	}

	if concreteStep.bufferSize != 15 {
		t.Errorf("Expected buffer size to be 15, got %d", concreteStep.bufferSize)
	}

	if cap(concreteStep.buffer) != 15 {
		t.Errorf("Expected buffer capacity to be 20, got %d", cap(concreteStep.buffer))
	}

	if concreteStep.inputChannelSize != 5 {
		t.Errorf("Expected input channel size to be 10, got %d", concreteStep.inputChannelSize)
	}
}

func TestStepBuffer_NewStep_MissingTimeAndInputTriggeres(t *testing.T) {

	// Test with StepConfig
	stepConfig := StepBufferConfig[int]{
		Label:                        "testStep",
		Replicas:                     1,
		TimeTriggeredProcessInterval: 10 * time.Second,
		BufferSize:                   10,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	newStepBuffer(stepConfig)
}

func TestStepBuffer_NewStep_TimeTriggeredWithoutInterval(t *testing.T) {

	// Test with StepConfig
	stepConfig := StepBufferConfig[int]{
		Label:                "testStep",
		Replicas:             1,
		TimeTriggeredProcess: func(input []int) (int, BufferFlags) { return 0, BufferFlags{} },
		BufferSize:           10,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	newStepBuffer(stepConfig)
}

func TestStepBuffer_NewStep_MissingBufferSize(t *testing.T) {

	// Test with StepConfig
	stepConfig := StepBufferConfig[int]{
		Label:                        "testStep",
		Replicas:                     1,
		TimeTriggeredProcess:         func(input []int) (int, BufferFlags) { return 0, BufferFlags{} },
		InputTriggeredProcess:        func(input []int) (int, BufferFlags) { return 0, BufferFlags{} },
		TimeTriggeredProcessInterval: 10 * time.Second,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	newStepBuffer(stepConfig)
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
