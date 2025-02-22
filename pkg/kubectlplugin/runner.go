package kubectlplugin

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Runner struct {
	ConfigFlags  *ConfigFlags
	handler      Handler
	KubeApi      kubernetes.Interface
	DiscoveryApi discovery.DiscoveryInterface
	DynamicApi   dynamic.Interface
}

type HandlerArgs struct {
	Runner
	Namespace string
	Cmd       *cobra.Command
	Args      []string
}

type Handler func(args HandlerArgs) (runtime.Object, error)

func NewRunner(handler Handler, opts ...RunnerOpts) *Runner {
	r := &Runner{
		ConfigFlags: NewConfigFlags(),
		handler:     handler,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Runner) ToCobraCommand(use, short string, opts ...CobraOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:          use,
		Short:        short,
		SilenceUsage: true,
		RunE:         r.cobraRunE(),
	}
	for _, opt := range opts {
		opt(cmd)
	}
	r.ConfigFlags.AddFlags(cmd)
	return cmd
}

func (r *Runner) cobraRunE() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		namespace, err := r.getNamespace()
		if err != nil {
			return err
		}

		if r.handler == nil {
			panic("handler cannot be nil")
		}

		return r.runHandlerAndPrint(
			cmd.OutOrStdout(),
			HandlerArgs{
				Runner:    *r,
				Namespace: namespace,
				Cmd:       cmd,
				Args:      args,
			},
		)
	}
}

func (r *Runner) getNamespace() (string, error) {
	kubeConfig := r.ConfigFlags.ToRawKubeConfigLoader()
	namespace, err := r.ConfigFlags.GetEffectiveNamespace(kubeConfig)
	if err != nil {
		return "", fmt.Errorf("failed to get effective namespace: %w", err)
	}
	return namespace, nil
}

func (r *Runner) runHandlerAndPrint(out io.Writer, args HandlerArgs) error {
	object, err := r.handler(args)
	if err != nil {
		return err
	}
	if object == nil {
		return nil
	}

	printer, err := r.ConfigFlags.ToPrinter()
	if err != nil {
		return fmt.Errorf("failed to get printer: %w", err)
	}
	if err := printer.PrintObj(object, out); err != nil {
		return fmt.Errorf("failed to print object: %w", err)
	}
	return nil
}
