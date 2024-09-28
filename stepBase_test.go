package pipelines

import (
	"testing"
)

func TestGetLabel(t *testing.T) {
	step := stepBase[int]{label: "testLabel"}
	if step.GetLabel() != "testLabel" {
		t.Errorf("expected label to be 'testLabel', got '%s'", step.GetLabel())
	}
}

func TestSetInputChannel(t *testing.T) {
	inputChan := make(chan int)
	step := stepBase[int]{}
	step.SetInputChannel(inputChan)
	if step.GetInputChannel() != inputChan {
		t.Errorf("expected input channel to be set correctly")
	}
}

func TestSetOutputChannel(t *testing.T) {
	outputChan := make(chan int)
	step := stepBase[int]{}
	step.SetOutputChannel(outputChan)
	if step.GetOutputChannel() != outputChan {
		t.Errorf("expected output channel to be set correctly")
	}
}

func TestGetReplicas(t *testing.T) {
	step := stepBase[int]{replicas: 5}
	if step.GetReplicas() != 5 {
		t.Errorf("expected replicas to be 5, got %d", step.GetReplicas())
	}
}

func TestSetDecrementTokensCountHandler(t *testing.T) {
	var called bool
	handler := func() { called = true }
	step := stepBase[int]{}
	step.SetDecrementTokensCountHandler(handler)
	step.decrementTokensCount()
	if !called {
		t.Errorf("expected decrementTokensCount handler to be called")
	}
}

func TestSetIncrementTokensCountHandler(t *testing.T) {
	var called bool
	handler := func() { called = true }
	step := stepBase[int]{}
	step.SetIncrementTokensCountHandler(handler)
	step.incrementTokensCount()
	if !called {
		t.Errorf("expected incrementTokensCount handler to be called")
	}
}
