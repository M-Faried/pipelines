package pipelines

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
			stepBase:     createBaseStep[I](c.Label, c.Replicas),
			errorHandler: c.ErrorHandler,
			process:      c.Process,
		}
	}

	if c, ok := config.(*StepFragmenterConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepFragmenter[I]{
			stepBase: createBaseStep[I](c.Label, c.Replicas),
			process:  c.Process,
		}
	}

	if c, ok := config.(*StepResultConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepResult[I]{
			stepBase: createBaseStep[I](c.Label, c.Replicas),
			process:  c.Process,
		}
	}

	if c, ok := config.(*StepFilterConfig[I]); ok {
		if c.PassCriteria == nil {
			panic("process is required")
		}
		return &stepFilter[I]{
			stepBase:     createBaseStep[I](c.Label, c.Replicas),
			passCriteria: c.PassCriteria,
		}
	}

	if c, ok := config.(*StepBufferedConfig[I]); ok {
		if c.InputTriggeredProcess == nil && c.TimeTriggeredProcess == nil {
			panic("either time triggered or input process is required")
		}
		if c.TimeTriggeredProcess != nil && c.TimeTriggeredProcessInterval == 0 {
			panic("time triggered process interval is required to be used with time triggered process")
		}
		if c.BufferSize <= 0 {
			panic("buffer size must be greater than or equal to 0")
		}
		return &stepBuffered[I]{
			stepBase:                     createBaseStep[I](c.Label, c.Replicas),
			bufferSize:                   c.BufferSize,
			inputTriggeredProcess:        c.InputTriggeredProcess,
			timeTriggeredProcess:         c.TimeTriggeredProcess,
			timeTriggeredProcessInterval: c.TimeTriggeredProcessInterval,
			passThrough:                  c.PassThrough,
			buffer:                       make([]I, 0, c.BufferSize),
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
	return pipe
}

func createBaseStep[I any](label string, replicas uint16) stepBase[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := stepBase[I]{}
	step.label = label
	step.replicas = replicas
	return step
}
