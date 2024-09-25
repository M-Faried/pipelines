package pipelines

import "sync"

// IStepConfig is an interface that defines the configuration for a step
type IStepConfig[I any] interface{}

// Builder is a struct that represents a builder for the pipeline parts.
type Builder[I any] struct{}

// NewStep creates a new step based on the configuration
func (s *Builder[I]) NewStep(config IStepConfig[I]) IStep[I] {

	if c, ok := config.(*StepConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepStandard[I]{
			stepBase: createBaseStep[I](c.Label, c.Replicas, c.ErrorHandler),
			process:  c.Process,
		}
	}

	if c, ok := config.(*StepFragmenterConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepFragmenter[I]{
			stepBase: createBaseStep[I](c.Label, c.Replicas, c.ErrorHandler),
			process:  c.Process,
		}
	}

	if c, ok := config.(*StepResultConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepResult[I]{
			stepBase: createBaseStep[I](c.Label, c.Replicas, c.ErrorHandler),
			process:  c.Process,
		}
	}

	return nil
}

// NewPipeline creates a new pipeline with the given channel size and steps.
// The channel size is the buffer size for all channels used to connect steps.
func (s *Builder[I]) NewPipeline(channelSize uint16, steps ...IStep[I]) IPipeline[I] {
	pipe := &pipeline[I]{}
	pipe.steps = castToInternalSteps(steps)
	pipe.channelSize = channelSize
	pipe.doneCond = sync.NewCond(&pipe.tokensMutex)
	return pipe
}

func createBaseStep[I any](label string, replicas uint16, errorHandler ErrorHandler) stepBase[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := stepBase[I]{}
	step.label = label
	step.replicas = replicas
	step.errorHandler = errorHandler
	return step
}
