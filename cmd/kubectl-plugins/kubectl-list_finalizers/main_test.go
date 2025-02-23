package main

import (
	"testing"

	kplug "github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/stretchr/testify/assert"
)

const Recurse = true
const DontEnforceNamespace = false

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)

	t.Run("OK", func(t *testing.T) {
		cmd := newCommand(kplug.WithDefaultDiscoveryApi())
		cmd.SetArgs([]string{"-A"})
		assert.NoError(cmd.Execute())
	})
}
