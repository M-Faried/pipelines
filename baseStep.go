package pipelines

type ReportError func(string, error)

// baseStep is a base struct for all steps
type baseStep[I any] struct {

	// id is an identifier of the step set by the user.
	id string

	// input is a channel for incoming data to the step.
	input chan I

	// replicas is a number of goroutines that will be running the step.
	replicas uint16

	// reportError is the function called when an error occurs in the step.
	reportError ReportError

	// decrementTokensCount is a function that decrements the number of tokens in the pipeline.
	decrementTokensCount func()
}
