package pipelines

// baseStep is a base struct for all steps
type baseStep[I any] struct {

	// id is an identifier of the step set by the user.
	id string

	// input is a channel for incoming data to the step.
	input chan I

	// replicas is a number of goroutines that will be running the step.
	replicas uint16

	// errorsQueue is a queue for errors that may be reported during the step execution.
	errorsQueue *Queue[error]
}
