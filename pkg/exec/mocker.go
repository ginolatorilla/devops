package exec

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockExecutor mocks the Executor function signature.
type MockExecutor struct {
	mock.Mock
}

// Do forwards the call to the MockExecutor and returns the MockExec.
//
// Because the mocked target is not an object, we need to use this function to handle the call.
func (m *MockExecutor) Do(ctx context.Context, cmd string, args ...string) Exec {
	var mockArgs []interface{}

	mockArgs = append(mockArgs, ctx, cmd)
	for _, arg := range args {
		mockArgs = append(mockArgs, arg)
	}

	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(Exec)
}

// Executor returns the Do function as an Executor.
func (m *MockExecutor) Executor() Executor {
	return m.Do
}
