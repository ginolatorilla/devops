package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockExec(t *testing.T) {
	m := MockExec{}
	m.On("Output").Return([]byte("hello"), nil)

	r, err := m.Output()

	assert.NoError(t, err)
	assert.Equal(t, "hello", string(r))
	m.AssertCalled(t, "Output")
}
