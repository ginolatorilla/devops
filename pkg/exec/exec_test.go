package exec_test

import (
	"context"
	"testing"

	goexec "os/exec"

	"github.com/ginolatorilla/devops/pkg/exec"
	"github.com/stretchr/testify/assert"
)

func TestMockExecutor_SingleCommand(t *testing.T) {
	t.Parallel()

	var (
		mockExecutor exec.MockExecutor
		mockExec     exec.MockExec
	)
	mockExecutor.On("Do", context.TODO(), "echo", "Hello, World").
		Return(&mockExec)
	mockExec.On("Output").
		Return([]byte("hi from test"), nil)

	executor := mockExecutor.Executor()
	out, err := executor(context.TODO(), "echo", "Hello, World").Output()

	assert.NoError(t, err)
	assert.Equal(t, "hi from test", string(out))
	mockExecutor.AssertExpectations(t)
	mockExec.AssertExpectations(t)
}

func TestMockExecutor_MultipleCommands(t *testing.T) {
	t.Parallel()

	var (
		mockExecutor exec.MockExecutor
		mockExec1    exec.MockExec
		mockExec2    exec.MockExec
	)
	mockExecutor.On("Do", context.TODO(), "echo", "Hello, World").
		Return(&mockExec1)
	mockExec1.On("Output").
		Return([]byte("hi from test"), nil)

	mockExecutor.On("Do", context.TODO(), "echo", "Good morning").
		Return(&mockExec2)
	mockExec2.On("Output").
		Return([]byte("good morning from test"), nil)

	executor := mockExecutor.Executor()
	out, err := executor(context.TODO(), "echo", "Hello, World").Output()

	assert.NoError(t, err)
	assert.Equal(t, "hi from test", string(out))
	mockExecutor.AssertCalled(t, "Do", context.TODO(), "echo", "Hello, World")
	mockExec1.AssertCalled(t, "Output")

	out, err = executor(context.TODO(), "echo", "Good morning").Output()

	assert.NoError(t, err)
	assert.Equal(t, "good morning from test", string(out))
	mockExecutor.AssertExpectations(t)
	mockExec2.AssertExpectations(t)
}

func TestMockExecutor_NonZeroExit(t *testing.T) {
	t.Parallel()

	var (
		mockExecutor exec.MockExecutor
		mockExec     exec.MockExec
	)
	mockExecutor.On("Do", context.TODO(), "something").
		Return(&mockExec)
	mockExec.On("Run").
		Return(&goexec.ExitError{})

	executor := mockExecutor.Executor()
	err := executor(context.TODO(), "something").Run()

	assert.Error(t, err)
	assert.Equal(t, &goexec.ExitError{}, err)
	mockExecutor.AssertExpectations(t)
	mockExec.AssertExpectations(t)
}
