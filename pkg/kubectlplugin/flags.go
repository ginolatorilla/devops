package kubectlplugin

import (
	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

type ConfigFlags struct {
	genericclioptions.ConfigFlags
	genericclioptions.PrintFlags
	AllNamespaces bool
	NoHeaders     bool
}

func NewConfigFlags() *ConfigFlags {
	pf := genericclioptions.NewPrintFlags("")
	pf.TypeSetterPrinter = printers.NewTypeSetter(scheme.Scheme)
	return &ConfigFlags{
		ConfigFlags: *genericclioptions.NewConfigFlags(true),
		PrintFlags:  *pf,
	}
}

func (f *ConfigFlags) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	f.ConfigFlags.AddFlags(flags)
	f.PrintFlags.AddFlags(cmd)
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

func (f *ConfigFlags) ToPrinter() (printers.ResourcePrinter, error) {
	var outputFormat string
	if f.OutputFormat != nil {
		outputFormat = *f.OutputFormat
	}

	if outputFormat == "" {
		tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
			NoHeaders:     f.NoHeaders,
			WithNamespace: f.AllNamespaces,
		})
		return f.TypeSetterPrinter.ToPrinter(tablePrinter), nil
	}
	return f.PrintFlags.ToPrinter()
}
