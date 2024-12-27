package exec

import "github.com/stretchr/testify/mock"

// MockExec is a mock implementation of the Exec interface.
//
// Use this mock to setup the "backend" behaviour of the command.
type MockExec struct {
	mock.Mock
}

func (m *MockExec) Run() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockExec) Output() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
