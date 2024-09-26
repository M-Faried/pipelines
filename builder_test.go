package pipelines

import (
	"testing"
)

func TestNewStep_StandardStep(t *testing.T) {
	builder := &Builder[int]{}

	process := func(input int) (int, error) { return input, nil }
	errorHandler := func(label string, err error) {}

	// Test with StepConfig
	stepConfig := &StepConfig[int]{
		Label:        "testStep",
		Replicas:     1,
		ErrorHandler: errorHandler,
		Process:      process,
	}

	step := builder.NewStep(stepConfig)
	var concreteStep *stepStandard[int]
	var ok bool
	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepStandard[int]); !ok {
		t.Error("Expected step to be of type stepStandard")
	}

	if step.GetLabel() != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", step.GetLabel())
	}

	if step.GetReplicas() != 1 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.errorHandler == nil {
		t.Error("Expected error handler to be set, got nil")
	}
	if concreteStep.process == nil {
		t.Error("Expected process to be set, got nil")
	}
}

func TestNewStep_FragmenterStep(t *testing.T) {
	builder := &Builder[int]{}

	process := func(input int) ([]int, error) { return []int{input}, nil }
	errorHandler := func(label string, err error) {}

	// Test with StepConfig
	stepConfig := &StepFragmenterConfig[int]{
		Label:        "fragmenterStep",
		Replicas:     1,
		ErrorHandler: errorHandler,
		Process:      process,
	}

	step := builder.NewStep(stepConfig)
	var concreteStep *stepFragmenter[int]
	var ok bool
	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepFragmenter[int]); !ok {
		t.Error("Expected step to be of type stepStandard")
	}

	if step.GetLabel() != "fragmenterStep" {
		t.Errorf("Expected label to be 'fragmenterStep', got '%s'", step.GetLabel())
	}

	if step.GetReplicas() != 1 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.errorHandler == nil {
		t.Error("Expected error handler to be set, got nil")
	}
	if concreteStep.process == nil {
		t.Error("Expected process to be set, got nil")
	}
}

func TestNewStep_ResultStep(t *testing.T) {
	builder := &Builder[int]{}

	process := func(input int) error { return nil }
	errorHandler := func(label string, err error) {}

	// Test with StepConfig
	stepConfig := &StepResultConfig[int]{
		Label:        "fragmenterStep",
		Replicas:     5,
		ErrorHandler: errorHandler,
		Process:      process,
	}

	step := builder.NewStep(stepConfig)
	var concreteStep *stepResult[int]
	var ok bool
	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepResult[int]); !ok {
		t.Error("Expected step to be of type stepStandard")
	}

	if step.GetLabel() != "fragmenterStep" {
		t.Errorf("Expected label to be 'fragmenterStep', got '%s'", step.GetLabel())
	}

	if step.GetReplicas() != 5 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.errorHandler == nil {
		t.Error("Expected error handler to be set, got nil")
	}
	if concreteStep.process == nil {
		t.Error("Expected process to be set, got nil")
	}
}

func TestNewPipeline(t *testing.T) {
	builder := &Builder[int]{}

	step1 := &mockStep[int]{}
	step2 := &mockStep[int]{}

	pipe := builder.NewPipeline(10, step1, step2)
	if pipe == nil {
		t.Error("Expected pipeline to be created, got nil")
	}

	var concretePipeline *pipeline[int]
	var ok bool
	if concretePipeline, ok = pipe.(*pipeline[int]); !ok {
		t.Error("Expected pipeline to be of type pipeline")
	}

	if len(concretePipeline.steps) != 2 {
		t.Errorf("Expected 2 steps in pipeline, got %d", len(concretePipeline.steps))
	}

	if concretePipeline.channelSize != 10 {
		t.Errorf("Expected channel size to be 10, got %d", concretePipeline.channelSize)
	}
}
