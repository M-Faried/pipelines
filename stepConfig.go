package pipelines

type IStepConfig[I any] interface{}

type StepConfig[I any] struct {
	Label        string
	Replicas     uint16
	ErrorHandler ErrorHandler
	Process      StepProcess[I]
}

type StepFragmenterConfig[I any] struct {
	Label        string
	Replicas     uint16
	ErrorHandler ErrorHandler
	Process      StepFragmenterProcess[I]
}

type StepResultConfig[I any] struct {
	Label        string
	Replicas     uint16
	ErrorHandler ErrorHandler
	Process      StepResultProcess[I]
}

func createBaseStep[I any](label string, replicas uint16, errorHandler ErrorHandler) step[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := step[I]{}
	step.label = label
	step.replicas = replicas
	step.errorHandler = errorHandler
	return step
}

func NewStep[I any](config IStepConfig[I]) IStep[I] {

	if c, ok := config.(*StepConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepStandard[I]{
			step:    createBaseStep[I](c.Label, c.Replicas, c.ErrorHandler),
			process: c.Process,
		}
	}

	if c, ok := config.(*StepFragmenterConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepFragmenter[I]{
			step:    createBaseStep[I](c.Label, c.Replicas, c.ErrorHandler),
			process: c.Process,
		}
	}

	if c, ok := config.(*StepResultConfig[I]); ok {
		if c.Process == nil {
			panic("process is required")
		}
		return &stepResult[I]{
			step:    createBaseStep[I](c.Label, c.Replicas, c.ErrorHandler),
			process: c.Process,
		}
	}

	return nil
}
