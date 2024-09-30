package pipelines

import (
	"testing"
)

func TestStepBase_GetLabel(t *testing.T) {
	step := stepBase[int]{label: "testLabel"}
	if step.GetLabel() != "testLabel" {
		t.Errorf("expected label to be 'testLabel', got '%s'", step.GetLabel())
	}
}

func TestStepBase_SetInputChannelSize(t *testing.T) {
	step := stepBase[int]{}
	step.SetInputChannelSize(10)
	if step.GetInputChannelSize() != 10 {
		t.Errorf("expected input channel size to be 10, got %d", step.GetInputChannelSize())
	}
}

func TestStepBase_SetInputChannel(t *testing.T) {
	inputChan := make(chan int)
	step := stepBase[int]{}
	step.SetInputChannel(inputChan)
	if step.GetInputChannel() != inputChan {
		t.Errorf("expected input channel to be set correctly")
	}
}

func TestStepBase_SetOutputChannel(t *testing.T) {
	outputChan := make(chan int)
	step := stepBase[int]{}
	step.SetOutputChannel(outputChan)
	if step.GetOutputChannel() != outputChan {
		t.Errorf("expected output channel to be set correctly")
	}
}

func TestStepBase_GetReplicas(t *testing.T) {
	step := stepBase[int]{replicas: 5}
	if step.GetReplicas() != 5 {
		t.Errorf("expected replicas to be 5, got %d", step.GetReplicas())
	}
}

func TestStepBase_SetDecrementTokensCountHandler(t *testing.T) {
	var called bool
	handler := func() { called = true }
	step := stepBase[int]{}
	step.SetDecrementTokensCountHandler(handler)
	step.decrementTokensCount()
	if !called {
		t.Errorf("expected decrementTokensCount handler to be called")
	}
}

func TestStepBase_SetIncrementTokensCountHandler(t *testing.T) {
	var called bool
	handler := func() { called = true }
	step := stepBase[int]{}
	step.SetIncrementTokensCountHandler(handler)
	step.incrementTokensCount()
	if !called {
		t.Errorf("expected incrementTokensCount handler to be called")
	}
}

func TestStepBase_NewStep(t *testing.T) {
	step := newBaseStep[int]("testStep", 6, 7)
	if step.label != "testStep" {
		t.Errorf("expected label to be 'testStep', got '%s'", step.label)
	}
	if step.replicas != 6 {
		t.Errorf("expected replicas to be 6, got %d", step.replicas)
	}
	if step.inputChannelSize != 7 {
		t.Errorf("expected input channel size to be 7, got %d", step.inputChannelSize)
	}

	if step.GetLabel() != "testStep" {
		t.Errorf("expected label to be 'testStep', got '%s'", step.GetLabel())
	}
	if step.GetReplicas() != 6 {
		t.Errorf("expected replicas to be 6, got %d", step.GetReplicas())
	}
	if step.GetInputChannelSize() != 7 {
		t.Errorf("expected input channel size to be 7, got %d", step.GetInputChannelSize())
	}
}
