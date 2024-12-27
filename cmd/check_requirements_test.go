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
					"gitVersion": "v1.29.0"
				}
			}`),
			nil,
		)
	cmd := newCheckRequirementsCmd(mockExecutor.Executor())

	err := cmd.Execute()

	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
	mockKubectl.AssertExpectations(t)
}
