package pipelines

import (
	"testing"
	"time"
)

func TestNewStep_StandardStep(t *testing.T) {
	builder := &Builder[int]{}

	process := func(input int) (int, error) { return input, nil }
	errorHandler := func(label string, err error) {}

	// Test with StepConfig
	stepConfig := &StepConfig[int]{
		Label:        "testStep",
		Replicas:     0, //should be rectified to 1
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

func TestNewStep_StandardStep_MissingProcess(t *testing.T) {

	builder := &Builder[int]{}
	errorHandler := func(label string, err error) {}

	// Test with StepConfig
	stepConfig := &StepConfig[int]{
		Label:        "testStep",
		Replicas:     1,
		ErrorHandler: errorHandler,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
}

func TestNewStep_FilterStep(t *testing.T) {
	builder := &Builder[int]{}

	process := func(input int) bool { return true }

	// Test with StepConfig
	stepConfig := &StepFilterConfig[int]{
		Label:        "testStep",
		Replicas:     1,
		PassCriteria: process,
	}

	step := builder.NewStep(stepConfig)

	var concreteStep *stepFilter[int]
	var ok bool

	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepFilter[int]); !ok {
		t.Error("Expected step to be of type stepStandard")
	}

	if step.GetLabel() != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", step.GetLabel())
	}

	if step.GetReplicas() != 1 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.passCriteria == nil {
		t.Error("Expected process to be set, got nil")
	}
}

func TestNewStep_FilterStep_MissingProcess(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := &StepFilterConfig[int]{
		Label:    "testStep",
		Replicas: 1,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
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

func TestNewStep_FragmenterStep_MissingProcess(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := &StepFragmenterConfig[int]{
		Label:    "testStep",
		Replicas: 1,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
}

func TestNewStep_BufferStep(t *testing.T) {
	builder := &Builder[int]{}

	processInputTriggered := func(input []int) StepBufferedProcessOutput[int] {
		return StepBufferedProcessOutput[int]{}
	}
	processTimeTriggered := func(input []int) StepBufferedProcessOutput[int] {
		return StepBufferedProcessOutput[int]{}
	}

	// Test with StepConfig
	stepConfig := &StepBufferedConfig[int]{
		Label:                        "testStep",
		Replicas:                     1,
		BufferSize:                   20,
		PassThrough:                  true,
		InputTriggeredProcess:        processInputTriggered,
		TimeTriggeredProcess:         processTimeTriggered,
		TimeTriggeredProcessInterval: 10 * time.Second,
	}

	step := builder.NewStep(stepConfig)

	var concreteStep *stepBuffered[int]
	var ok bool

	if step == nil {
		t.Error("Expected step to be created, got nil")
	}
	if concreteStep, ok = step.(*stepBuffered[int]); !ok {
		t.Error("Expected step to be of type stepStandard")
	}

	if concreteStep.label != "testStep" {
		t.Errorf("Expected label to be 'testStep', got '%s'", step.GetLabel())
	}

	if concreteStep.replicas != 1 {
		t.Errorf("Expected replicas to be 1, got %d", step.GetReplicas())
	}

	if concreteStep.bufferSize != 20 {
		t.Errorf("Expected buffer size to be 20, got %d", concreteStep.bufferSize)
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

	if cap(concreteStep.buffer) != 20 {
		t.Errorf("Expected buffer capacity to be 20, got %d", cap(concreteStep.buffer))
	}
}

func TestNewStep_BufferStep_MissingTimeAndInputTriggeres(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := &StepBufferedConfig[int]{
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

	builder.NewStep(stepConfig)
}

func TestNewStep_BufferStep_TimeTriggeredWithoutInterval(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := &StepBufferedConfig[int]{
		Label:                "testStep",
		Replicas:             1,
		TimeTriggeredProcess: func(input []int) StepBufferedProcessOutput[int] { return StepBufferedProcessOutput[int]{} },
		BufferSize:           10,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
}

func TestNewStep_BufferStep_MissingBufferSize(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := &StepBufferedConfig[int]{
		Label:                        "testStep",
		Replicas:                     1,
		TimeTriggeredProcess:         func(input []int) StepBufferedProcessOutput[int] { return StepBufferedProcessOutput[int]{} },
		InputTriggeredProcess:        func(input []int) StepBufferedProcessOutput[int] { return StepBufferedProcessOutput[int]{} },
		TimeTriggeredProcessInterval: 10 * time.Second,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
}

func TestNewStep_ResultStep(t *testing.T) {
	builder := &Builder[int]{}

	process := func(input int) {}

	// Test with StepConfig
	stepConfig := &StepResultConfig[int]{
		Label:    "fragmenterStep",
		Replicas: 5,
		Process:  process,
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

	if concreteStep.process == nil {
		t.Error("Expected process to be set, got nil")
	}
}

func TestNewStep_ResultStep_MissingProcess(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := &StepResultConfig[int]{
		Label:    "testStep",
		Replicas: 1,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
}

func TestNewStep_InvalidConfig(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := builder

	step := builder.NewStep(stepConfig)

	if step != nil {
		t.Error("Expected step to be nil, got something")
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
