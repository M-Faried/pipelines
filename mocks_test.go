package pipelines

// mockErrorHandler is a mock implementation of the error handler
type mockErrorHandler struct {
	called bool
	label  string
	err    error
}

func (m *mockErrorHandler) Handle(label string, err error) {
	m.called = true
	m.label = label
	m.err = err
}

type mockResultProcessHandler[I any] struct {
	called bool
	input  I
}

func (m *mockResultProcessHandler[I]) Handle(input I) error {
	m.called = true
	m.input = input
	return nil
}

type mockDecrementTokensHandler struct {
	called bool
	value  int
}

func (m *mockDecrementTokensHandler) Handle() {
	m.called = true
	m.value--
}

type mockIncrementTokensHandler struct {
	called bool
	value  int
}

func (m *mockIncrementTokensHandler) Handle() {
	m.called = true
	m.value++
}
