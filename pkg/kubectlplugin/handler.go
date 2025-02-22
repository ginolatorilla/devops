package kubectlplugin

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

type HandlerArgs struct {
	Runner
	Namespace string
	Cmd       *cobra.Command
	Args      []string
}

type Handler func(args HandlerArgs) (runtime.Object, error)

func (a HandlerArgs) ToResourceBuilder() *resource.Builder {
	return resource.NewBuilder(a.ConfigFlags).
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		AllNamespaces(*a.ConfigFlags.AllNamespaces).
		NamespaceParam(*a.ConfigFlags.Namespace).
		ContinueOnError()
}
