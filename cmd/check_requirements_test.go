package cmd

import (
	"testing"

	u "github.com/ginolatorilla/devops/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCheckRequirements(t *testing.T) {
	t.Parallel()
	mock := u.MockExec{}
	mock.On("Output").
		Return(
			[]byte(`{
				"clientVersion": {
					"gitVersion": "v1.29.0"
				}
			}`),
			nil,
		)
	cmd := newCheckRequirementsCmd(u.NewMockExecutor(&mock))

	err := cmd.Execute()

	assert.NoError(t, err)
	mock.AssertCalled(t, "Output")
}
