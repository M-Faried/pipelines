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

	if c, ok := config.(*StepFilterConfig[I]); ok {
		if c.Criteria == nil {
			panic("process is required")
		}
		return &stepFilter[I]{
			stepBase: createBaseStep[I](c.Label, c.Replicas, nil),
			criteria: c.Criteria,
		}
	}

	if c, ok := config.(*StepMonitorConfig[I]); ok {
		if c.NotifyCriteria == nil && c.CheckInterval == 0 {
			panic("either notify criteria or check interval should be set for the monitor step")
		}
		return &stepMonitor[I]{
			stepBase:       createBaseStep[I](c.Label, c.Replicas, nil),
			notifyCriteria: c.NotifyCriteria,
			notify:         c.Notify,
			checkInterval:  c.CheckInterval,
			buffer:         make([]I, 0),
		}
	}

	if c, ok := config.(*StepAggregatorConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		if c.ThresholdCriteria == nil && c.AggregationInterval == 0 {
			panic("either threshold or time duration must be set for the aggregation step")
		}
		return &stepAggregator[I]{
			stepBase:            createBaseStep[I](c.Label, c.Replicas, c.ErrorHandler),
			process:             c.Process,
			thresholdCriteria:   c.ThresholdCriteria,
			aggregationInterval: c.AggregationInterval,
			buffer:              make([]I, 0),
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
