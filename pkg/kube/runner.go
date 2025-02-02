package kube

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Runner struct {
	configFlags *ConfigFlags
	apiFactory  ApiFactory
	mainFunc    MainFunc
}

type MainFunc func(kubeApi kubernetes.Interface, namespace string, cmd *cobra.Command, args []string, configFlags *ConfigFlags) metaV1.Table

func NewRunner(
	configFlags *ConfigFlags,
	apiFactory ApiFactory,
	mainFunc MainFunc,
) *Runner {
	return &Runner{
		configFlags: configFlags,
		apiFactory:  apiFactory,
		mainFunc:    mainFunc,
	}
}

func (r *Runner) ToRun() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		kubeConfig := r.configFlags.ToRawKubeConfigLoader()
		namespace, err := r.configFlags.GetEffectiveNamespace(kubeConfig)
		if err != nil {
			panic(fmt.Errorf("failed to get effective namespace: %w", err))
		}

		table := r.mainFunc(r.apiFactory(r.configFlags), namespace, cmd, args, r.configFlags)

		if err := r.configFlags.GetTablePrinter().PrintObj(&table, cmd.OutOrStdout()); err != nil {
			panic(fmt.Errorf("failed to print table: %w", err))
		}
	}
}
