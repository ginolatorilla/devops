package kube

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type Handler func(args HandlerArgs)

func NewRunner(configFlags *ConfigFlags, apiFactory ApiFactory, handler Handler) *Runner {
	return &Runner{
		configFlags: configFlags,
		apiFactory:  apiFactory,
		handler:     handler,
	}
}

type TabularHandler func(args HandlerArgs) metaV1.Table

func NewTabularRunner(configFlags *ConfigFlags, apiFactory ApiFactory, handler TabularHandler) *Runner {
	return &Runner{
		configFlags:    configFlags,
		apiFactory:     apiFactory,
		tabularHandler: handler,
	}
}

func NewTabularRunnerWithDiscoveryApi(configFlags *ConfigFlags, apiFactory DynamicApiFactory, handler TabularHandler) *Runner {
	return &Runner{
		configFlags:       configFlags,
		dynamicApiFactory: apiFactory,
		tabularHandler:    handler,
	}
}

func (r *Runner) ToRun() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		kubeConfig := r.configFlags.ToRawKubeConfigLoader()
		namespace, err := r.configFlags.GetEffectiveNamespace(kubeConfig)
		if err != nil {
			panic(fmt.Errorf("failed to get effective namespace: %w", err))
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
			r.handler(handlerArgs)
			return
		}

		if r.tabularHandler != nil {
			table := r.tabularHandler(handlerArgs)
			if err := r.configFlags.GetTablePrinter().PrintObj(&table, cmd.OutOrStdout()); err != nil {
				panic(fmt.Errorf("failed to print table: %w", err))
			}
		}
	}
}
