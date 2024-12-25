package utils

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type Exec interface {
	Run() error
	Output() ([]byte, error)
}

type Executor func(ctx context.Context, cmd string, args ...string) Exec

type MockExec struct {
	mock.Mock
}

func NewMockExecutor(mock *MockExec) Executor {
	return func(ctx context.Context, cmd string, args ...string) Exec {
		return mock
	}
}

func (m *MockExec) Run() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockExec) Output() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
