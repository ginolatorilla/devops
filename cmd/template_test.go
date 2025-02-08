package cmd

import (
	"bytes"
	"fmt"
	"os"
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

func TestTemplateParseX509(t *testing.T) {
	t.Parallel()

	crtRaw, err := os.ReadFile("testdata/tls.crt")
	if err != nil {
		t.Fatal(err)
	}

	cmd := newTemplateCmd()
	cmd.SetIn(bytes.NewBufferString(fmt.Sprintf("{{ index (`%s` | x509) `NotAfter` }}", string(crtRaw))))
	var out bytes.Buffer
	cmd.SetOut(&out)

	assert.NoError(t, cmd.Execute())
	assert.Equal(t, "2026-02-01 06:31:00 +0000 UTC", out.String())
}
