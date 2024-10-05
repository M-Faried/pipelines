package pipelines

import "fmt"

// StepConfig is an interface that defines the configuration for a step
type StepConfig[I any] interface{}

// Builder is a struct that represents a builder for the pipeline parts.
type Builder[I any] struct{}

// NewStep creates a new step based on the configuration
func (s *Builder[I]) NewStep(config StepConfig[I]) IStep[I] {
	switch c := config.(type) {
	case StepBasicConfig[I]:
		return newStepBasic(c)
	case StepFragmenterConfig[I]:
		return newStepFragmenter(c)
	case StepTerminalConfig[I]:
		return newStepTerminal(c)
	case StepFilterConfig[I]:
		return newStepFilter(c)
	case StepBufferConfig[I]:
		return newStepBuffer(c)
	default:
		panic(fmt.Sprintf("unknown step configuration: %v", config))
	}
}

// NewPipeline creates a new pipeline with the given channel size and steps.
func (s *Builder[I]) NewPipeline(config PipelineConfig, steps ...IStep[I]) IPipeline[I] {

	if config.DefaultStepInputChannelSize == 0 {
		panic("DefaultStepChannelSize configuration is not set")
	}

	pipe := &pipeline[I]{}
	pipe.steps = steps
	pipe.trackTokensCount = config.TrackTokensCount
	pipe.defaultChannelSize = config.DefaultStepInputChannelSize
	return pipe
}
