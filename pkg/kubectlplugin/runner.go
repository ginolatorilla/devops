package kubectlplugin

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Runner struct {
	configFlags       *ConfigFlags
	apiFactory        ApiFactory
	dynamicApiFactory DynamicApiFactory
	handler           Handler
	tabularHandler    TabularHandler
}

type HandlerArgs struct {
	KubeApi      kubernetes.Interface
	DiscoveryApi discovery.DiscoveryInterface
	DynamicApi   dynamic.Interface
	Namespace    string
	Cmd          *cobra.Command
	Args         []string
	ConfigFlags  *ConfigFlags
}

type Handler func(args HandlerArgs) (runtime.Object, error)

func NewRunner(apiFactory ApiFactory, handler Handler) *Runner {
	return &Runner{
		configFlags: NewConfigFlagsWithResourcePrinters(),
		apiFactory:  apiFactory,
		handler:     handler,
	}
}

func NewRunnerWithoutResourcePrinters(apiFactory ApiFactory, handler Handler) *Runner {
	return &Runner{
		configFlags: NewConfigFlags(),
		apiFactory:  apiFactory,
		handler:     handler,
	}
}

func (r *Runner) ToCobraCommand(use, short string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          use,
		Short:        short,
		SilenceUsage: true,
		RunE:         r.toRunE(),
	}

	r.configFlags.AddFlags(cmd)
	return cmd
}

func (r *Runner) ToCobraCommandWithArgs(use, short string, args cobra.PositionalArgs, validArgs []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          use,
		Args:         args,
		ValidArgs:    validArgs,
		Short:        short,
		SilenceUsage: true,
		RunE:         r.toRunE(),
	}

	r.configFlags.AddFlags(cmd)
	return cmd
}

func (r *Runner) toRunE() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		kubeConfig := r.configFlags.ToRawKubeConfigLoader()
		namespace, err := r.configFlags.GetEffectiveNamespace(kubeConfig)
		if err != nil {
			return fmt.Errorf("failed to get effective namespace: %w", err)
		}

		var kubeApi kubernetes.Interface
		if r.apiFactory != nil {
			kubeApi = r.apiFactory(r.configFlags)
		}

		var discoveryApi discovery.DiscoveryInterface
		var dynamicApi dynamic.Interface
		if r.dynamicApiFactory != nil {
			discoveryApi, dynamicApi = r.dynamicApiFactory(r.configFlags)
		}

		handlerArgs := HandlerArgs{
			KubeApi:      kubeApi,
			DiscoveryApi: discoveryApi,
			DynamicApi:   dynamicApi,
			Namespace:    namespace,
			Cmd:          cmd,
			Args:         args,
			ConfigFlags:  r.configFlags,
		}

		if r.handler != nil {
			object, err := r.handler(handlerArgs)
			if err != nil {
				return err
			}
			if object == nil {
				return nil
			}

			printer, err := r.configFlags.ToPrinter()
			if err != nil {
				return fmt.Errorf("failed to get printer: %w", err)
			}
			if err := printer.PrintObj(object, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("failed to print object: %w", err)
			}
			return nil
		}

		if r.tabularHandler != nil {
			table, err := r.tabularHandler(handlerArgs)
			if table.APIVersion == "" {
				table.APIVersion = "meta.k8s.io/v1"
			}
			if table.Kind == "" {
				table.Kind = "Table"
			}

			if err != nil {
				return err
			}

			printer, err := r.configFlags.ToPrinter()
			if err != nil {
				return fmt.Errorf("failed to get printer: %w", err)
			}
			if err := printer.PrintObj(&table, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("failed to print table: %w", err)
			}
		}
		return nil
	}
}
