package pipelines

// stepBase is a base struct for all steps
type stepBase[I any] struct {

	// label is an label for the step set by the user.
	label string

	// input is a channel for incoming data to the step.
	input chan I

	// ouput is a channel for outgoing data from the step.
	output chan I

	// replicas is a number of goroutines that will be running the step.
	replicas uint16

	// decrementTokensCount is a function that decrements the number of tokens in the pipeline.
	decrementTokensCount func()

	// incrementTokensCount is a function that increments the number of tokens in the pipeline.
	incrementTokensCount func()
}

func newBaseStep[I any](label string, replicas uint16) stepBase[I] {
	if replicas == 0 {
		replicas = 1
	}
	step := stepBase[I]{}
	step.label = label
	step.replicas = replicas
	return step
}

func (s *stepBase[I]) GetLabel() string {
	return s.label
}

func (s *stepBase[I]) SetInputChannel(input chan I) {
	s.input = input
}

func (s *stepBase[I]) GetInputChannel() chan I {
	return s.input
}

func (s *stepBase[I]) SetOutputChannel(output chan I) {
	s.output = output
}

func (s *stepBase[I]) GetOutputChannel() chan I {
	return s.output
}

func (s *stepBase[I]) GetReplicas() uint16 {
	return s.replicas
}

func (s *stepBase[I]) SetDecrementTokensCountHandler(handler func()) {
	s.decrementTokensCount = handler
}

func (s *stepBase[I]) SetIncrementTokensCountHandler(handler func()) {
	s.incrementTokensCount = handler
}
