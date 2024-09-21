package pipelines

func NewPipelineWithResultHandler[I any](channelSize uint16, resultStep *ResultStep[I], steps ...*Step[I]) IPipeline[I] {
	return &pipeline[I]{
		steps:       steps,
		resultStep:  resultStep,
		channelSize: channelSize,
	}
}

func NewPipeline[I any](channelSize uint16, steps ...*Step[I]) IPipeline[I] {
	return &pipeline[I]{
		steps:       steps,
		channelSize: channelSize,
	}
}
