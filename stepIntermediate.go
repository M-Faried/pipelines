package pipelines

type stepIntermediate[I any] struct {
	stepBase[I]
	// ouput is a channel for outgoing data from the step.
	output chan I
	// incrementTokensCount is a function that increments the number of tokens in the pipeline.
	incrementTokensCount func()
}

func (s *stepIntermediate[I]) setOuputChannel(output chan I) {
	s.output = output
}

func (s *stepIntermediate[I]) getOuputChannel() chan I {
	return s.output
}

func (s *stepIntermediate[I]) setIncrementTokensCountHandler(handler func()) {
	s.incrementTokensCount = handler
}
