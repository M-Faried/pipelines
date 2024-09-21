package pipelines

type baseStep[I any] struct {
	id          string
	input       chan I
	errorsQueue *Queue[error]
	replicas    uint8
}

func (s *baseStep[I]) GetID() string {
	return s.id
}
