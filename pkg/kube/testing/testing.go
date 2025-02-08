package testing

import (
	"fmt"
	gotesting "testing"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func LoadCannedError(t *gotesting.T, client *fake.Clientset, verb, resource string) {
	t.Helper()

	client.PrependReactor(verb, resource, func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("canned error from test")
	})
}

func Testable(client *fake.Clientset, commandFactory func(kube.ApiFactory) *cobra.Command) *cobra.Command {
	return commandFactory(func(cf *kube.ConfigFlags) kubernetes.Interface {
		return client
	})
}
