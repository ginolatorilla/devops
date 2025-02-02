package kube

import (
	"github.com/spf13/pflag"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/tools/clientcmd"
)

type ConfigFlags struct {
	genericclioptions.ConfigFlags
	AllNamespaces bool
	NoHeaders     bool
}

func NewConfigFlags() *ConfigFlags {
	return &ConfigFlags{
		ConfigFlags: *genericclioptions.NewConfigFlags(true),
	}
}

func (f *ConfigFlags) AddFlags(flags *pflag.FlagSet) {
	f.ConfigFlags.AddFlags(flags)
	flags.BoolVarP(
		&f.AllNamespaces,
		"all-namespaces",
		"A",
		false,
		"If present, list the requested object(s) across all namespaces. "+
			"Namespace in current context is ignored even if specified with --namespace.",
	)
	flags.BoolVar(
		&f.NoHeaders,
		"no-headers",
		false,
		"Don't print headers (default print headers).",
	)
}

func (f *ConfigFlags) GetEffectiveNamespace(kubeConfig clientcmd.ClientConfig) (string, error) {
	if f.AllNamespaces {
		return coreV1.NamespaceAll, nil
	}

	var namespace string
	if f.Namespace == nil {
		namespace = ""
	} else {
		namespace = *f.Namespace
	}

	if namespace == "" {
		var err error
		namespace, _, err = kubeConfig.Namespace()
		if err != nil {
			return "", err
		}
	}

	return namespace, nil
}

func (f *ConfigFlags) GetTablePrinter() printers.ResourcePrinter {
	return printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: f.AllNamespaces,
		NoHeaders:     f.NoHeaders,
	})
}
