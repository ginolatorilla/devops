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

type MainFunc func(kubeApi kubernetes.Interface, namespace string, cmd *cobra.Command, configFlags *ConfigFlags) (metaV1.Table, error)

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

func (r *Runner) ToRunE() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		kubeConfig := r.configFlags.ToRawKubeConfigLoader()
		namespace, err := r.configFlags.GetEffectiveNamespace(kubeConfig)
		if err != nil {
			return fmt.Errorf("failed to get effective namespace: %w", err)
		}

		table, err := r.mainFunc(r.apiFactory(r.configFlags), namespace, cmd, r.configFlags)
		if err != nil {
			return err
		}

		return r.configFlags.
			GetTablePrinter().
			PrintObj(&table, cmd.OutOrStdout())
	}
}
