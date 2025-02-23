package main

import (
	"os"
	"path/filepath"
	"testing"

	kplug "github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)
	t.Run("OK", func(t *testing.T) {
		cmd := newCommand(kplug.WithDefaultKubeApi())
		cmd.SetArgs([]string{"-A"})
		assert.NoError(cmd.Execute())
	})
}

func Test_addSecretsWithTLSCertsToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder()
	visitor := addSecretsWithTLSCertsToTable(tbuild)
	secret := loadTLSCertsFromPath(t, "testdata/broken-certs")

	assert.NoError(t, visitor(&resource.Info{Object: secret}, nil))
	assert.Empty(t, tbuild.Table.Rows)
}

func loadTLSCertsFromPath(t *testing.T, certsDir string) *coreV1.Secret {
	t.Helper()

	caCrt, err := os.ReadFile(filepath.Join(certsDir, "ca.crt"))
	if err != nil {
		t.Fatal(err)
	}
	tlsCrt, err := os.ReadFile(filepath.Join(certsDir, "tls.crt"))
	if err != nil {
		t.Fatal(err)
	}
	tlsKey, err := os.ReadFile(filepath.Join(certsDir, "tls.key"))
	if err != nil {
		t.Fatal(err)
	}

	return &coreV1.Secret{
		Data: map[string][]byte{
			coreV1.TLSCertKey:              tlsCrt,
			coreV1.TLSPrivateKeyKey:        tlsKey,
			coreV1.ServiceAccountRootCAKey: caCrt,
		},
		Type: coreV1.SecretTypeTLS,
	}
}
