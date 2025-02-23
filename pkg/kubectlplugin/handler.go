package kubectlplugin

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type HandlerArgs struct {
	Runner
	Cmd  *cobra.Command
	Args []string
}

type Handler func(args HandlerArgs) (runtime.Object, error)

func (a HandlerArgs) ToResourceFinder(resources ...string) genericclioptions.ResourceFinder {
	return a.ConfigFlags.ToBuilder(a.ConfigFlags, resources)
}
