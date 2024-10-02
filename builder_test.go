package pipelines

import (
	"testing"
)

func TestBuilder_TestNewStep(t *testing.T) {
	builder := &Builder[int]{}

	tests := []struct {
		name    string
		config  StepConfig[int]
		wantErr bool
	}{
		{"BasicConfig", StepBasicConfig[int]{
			Process: func(int) int { return 0 },
		}, false},
		{"FragmenterConfig", StepFragmenterConfig[int]{
			Process: func(int) []int { return []int{} },
		}, false},
		{"TerminalConfig", StepTerminalConfig[int]{
			Process: func(int) {},
		}, false},
		{"FilterConfig", StepFilterConfig[int]{
			PassCriteria: func(int) bool { return false },
		}, false},
		{"BufferedConfig", StepBufferConfig[int]{
			InputTriggeredProcess: func([]int) (int, BufferFlags) { return 0, BufferFlags{} },
			BufferSize:            5,
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("NewStep() panicked unexpectedly: %v", r)
					}
				}
			}()

			step := builder.NewStep(tt.config)
			if (step == nil) != tt.wantErr {
				t.Errorf("NewStep() = %v, wantErr %v", step, tt.wantErr)
			}
		})
	}
}
func TestBuilder_InvalidConfig(t *testing.T) {

	builder := &Builder[int]{}

	// Test with StepConfig
	stepConfig := builder

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected to panic, got nil")
		}
	}()

	builder.NewStep(stepConfig)
}

func TestBuilder_NewPipeline(t *testing.T) {
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

	if concretePipeline.defaultChannelSize != 10 {
		t.Errorf("Expected channel size to be 10, got %d", concretePipeline.defaultChannelSize)
	}
}
