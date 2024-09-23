package pipelines

// ReportError is the definition of error reporting handler which may or may not be set by the user during creation.
// The first parameter is the label of the step where the error occurred and the second parameter is the error itself.
type ReportError func(string, error)

// baseStep is a base struct for all steps
type baseStep[I any] struct {

	// label is an label for the step set by the user.
	label string

	// input is a channel for incoming data to the step.
	input chan I

	// replicas is a number of goroutines that will be running the step.
	replicas uint16

	// reportError is the function called when an error occurs in the step.
	reportError ReportError

	// decrementTokensCount is a function that decrements the number of tokens in the pipeline.
	decrementTokensCount func()
}
