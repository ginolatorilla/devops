package cmd

import (
	"context"
	"testing"

	"github.com/ginolatorilla/devops/pkg/exec"
	"github.com/stretchr/testify/assert"
)

func TestCheckRequirements(t *testing.T) {
	t.Parallel()
	var (
		mockExecutor exec.MockExecutor
		mockKubectl  exec.MockExec
		mockJq       exec.MockExec
	)
	mockExecutor.
		On("Do", context.Background(), "kubectl", "version", "--client", "--output", "json").
		Return(&mockKubectl)
	mockKubectl.
		On("Output").
		Return(
			[]byte(`{
				"clientVersion": {
					"gitVersion": "v1.29.0"
				}
			}`),
			nil,
		)
	mockExecutor.
		On("Do", context.Background(), "jq", "--version").
		Return(&mockJq)
	mockJq.
		On("Output").
		Return([]byte("jq-1.7"), nil)

	cmd := newCheckRequirementsCmd(mockExecutor.Executor())

	err := cmd.Execute()

	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
	mockKubectl.AssertExpectations(t)
}

func TestCheckRequirements_PanicIfNotMet(t *testing.T) {
	t.Parallel()
	var (
		mockKubectl  exec.MockExec
		mockExecutor exec.MockExecutor
	)
	mockExecutor.
		On("Do", context.Background(), "kubectl", "version", "--client", "--output", "json").
		Return(&mockKubectl)
	mockKubectl.
		On("Output").
		Return(
			[]byte(`{
				"clientVersion": {
					"gitVersion": "v0.0.0"
				}
			}`),
			nil,
		)
	cmd := newCheckRequirementsCmd(mockExecutor.Executor())

	assert.Panics(t, func() { cmd.Execute() })
	mockExecutor.AssertExpectations(t)
	mockKubectl.AssertExpectations(t)
}
