package kubectl_list_certs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewCommand(t *testing.T) {
	caCrt, err := os.ReadFile("testdata/ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	tlsCrt, err := os.ReadFile("testdata/tls.crt")
	if err != nil {
		t.Fatal(err)
	}
	tlsKey, err := os.ReadFile("testdata/tls.key")
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)

	t.Run("WithExplicitNamespace", func(t *testing.T) {
		client := fake.NewClientset()
		_, err = client.CoreV1().Secrets("test").Create(
			context.TODO(),
			&coreV1.Secret{
				ObjectMeta: metaV1.ObjectMeta{Name: "test"},
				Data: map[string][]byte{
					coreV1.TLSCertKey:              tlsCrt,
					coreV1.TLSPrivateKeyKey:        tlsKey,
					coreV1.ServiceAccountRootCAKey: caCrt,
				},
			},
			metaV1.CreateOptions{},
		)
		if err != nil {
			t.Fatal(err)
		}

		cmd := NewCommand(client)
		cmd.SetArgs([]string{"-n", "test"})
		assert.NoError(cmd.Execute())

	})

	t.Run("NamespaceFromKubeconfig", func(t *testing.T) {
		client := fake.NewClientset()
		_, err = client.CoreV1().Secrets("test").Create(
			context.TODO(),
			&coreV1.Secret{
				ObjectMeta: metaV1.ObjectMeta{Name: "test"},
				Data: map[string][]byte{
					coreV1.TLSCertKey:              tlsCrt,
					coreV1.TLSPrivateKeyKey:        tlsKey,
					coreV1.ServiceAccountRootCAKey: caCrt,
				},
			},
			metaV1.CreateOptions{},
		)
		if err != nil {
			t.Fatal(err)
		}
		kubeConfigPath, err := filepath.Abs("testdata/kubeConfig.yaml")
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("KUBECONFIG", kubeConfigPath)
		cmd := NewCommand(client)
		assert.NoError(cmd.Execute())
	})
}
