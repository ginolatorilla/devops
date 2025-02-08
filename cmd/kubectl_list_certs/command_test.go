package kubectl_list_certs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)

	t.Run("WithExplicitNamespace", func(t *testing.T) {
		client := fake.NewClientset()
		loadTLSCertsFromPath(t, client, "testdata/ok", "test", "test")

		cmd := testable(client)
		cmd.SetArgs([]string{"-n", "test"})
		assert.NoError(cmd.Execute())
	})

	t.Run("NamespaceFromKubeconfig", func(t *testing.T) {
		client := fake.NewClientset()
		loadTLSCertsFromPath(t, client, "testdata/ok", "test", "test")
		setKubeConfigEnv(t, "testdata/ok/kubeConfig.yaml")

		cmd := testable(client)
		assert.NoError(cmd.Execute())
	})

	t.Run("DefaultNamespace", func(t *testing.T) {
		client := fake.NewClientset()
		loadTLSCertsFromPath(t, client, "testdata/ok", "default", "test")
		setKubeConfigEnv(t, "testdata/missing-namespace/kubeConfig.yaml")

		cmd := testable(client)
		assert.NoError(cmd.Execute())
	})

	t.Run("ErrorIfUnableToListSecrets", func(t *testing.T) {
		client := fake.NewClientset()
		client.PrependReactor("list", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &coreV1.SecretList{}, fmt.Errorf("canned error from test")
		})

		cmd := testable(client)
		assert.Error(cmd.Execute())
	})

	t.Run("ErrorIfEmptyKubeConfig", func(t *testing.T) {
		client := fake.NewClientset()
		setKubeConfigEnv(t, "testdata/invalid/kubeConfig.yaml")

		cmd := testable(client)
		assert.Error(cmd.Execute())
	})

	t.Run("SkipIfCertificateIsInvalid", func(t *testing.T) {
		client := fake.NewClientset()
		loadTLSCertsFromPath(t, client, "testdata/broken-certs", "default", "test")

		cmd := testable(client)
		assert.NoError(cmd.Execute())
	})
}

func testable(client *fake.Clientset) *cobra.Command {
	return NewCommand(func(configFlags *kube.ConfigFlags) kubernetes.Interface {
		return client
	})
}

func loadTLSCertsFromPath(t *testing.T, client *fake.Clientset, certsDir, namespace, secretName string) {
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

	_, err = client.CoreV1().Secrets(namespace).Create(
		context.TODO(),
		&coreV1.Secret{
			ObjectMeta: metaV1.ObjectMeta{Name: secretName},
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
}

func setKubeConfigEnv(t *testing.T, kubeConfigPath string) {
	t.Helper()
	kubeConfigPath, err := filepath.Abs(kubeConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", kubeConfigPath)
}
