package kubectlplugin

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
)

type ConfigFlags struct {
	genericclioptions.ConfigFlags
	genericclioptions.PrintFlags
	genericclioptions.ResourceBuilderFlags
	NoHeaders bool
}

func NewConfigFlags() *ConfigFlags {
	rbf := &genericclioptions.ResourceBuilderFlags{}
	rbf.WithScheme(scheme.Scheme)
	return &ConfigFlags{
		ConfigFlags:          *genericclioptions.NewConfigFlags(true),
		ResourceBuilderFlags: *rbf,
	}
}

func (f *ConfigFlags) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	f.ConfigFlags.AddFlags(flags)
	f.PrintFlags.AddFlags(cmd)
	f.ResourceBuilderFlags.AddFlags(flags)
	flags.BoolVar(
		&f.NoHeaders,
		"no-headers",
		false,
		"Don't print headers (default print headers).",
	)
}

func (f *ConfigFlags) ToPrinter() (printers.ResourcePrinter, error) {
	var outputFormat string
	if f.OutputFormat != nil {
		outputFormat = *f.OutputFormat
	}

	if outputFormat == "" {
		tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
			NoHeaders:     f.NoHeaders,
			WithNamespace: f.AllNamespaces != nil && *f.AllNamespaces,
		})
		return f.TypeSetterPrinter.ToPrinter(tablePrinter), nil
	}
	return f.PrintFlags.ToPrinter()
}
