package pipelines

type stepIntermediate[I any] struct {
	stepBase[I]
	// ouput is a channel for outgoing data from the step.
	output chan I
}

func (s *stepIntermediate[I]) setOuputChannel(output chan I) {
	s.output = output
}

func (s *stepIntermediate[I]) getOuputChannel() chan I {
	return s.output
}

func (s *stepBase[I]) setIncrementTokensCountHandler(handler func()) {
	s.incrementTokensCount = handler
}
