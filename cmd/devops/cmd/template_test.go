package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	t.Parallel()
	cmd := newTemplateCmd()
	cmd.SetIn(bytes.NewBufferString(`{{ "Hello, World!" }}`))
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", out.String())
}
