package kubectlplugin

import (
	"fmt"
	"io"
	"log/slog"
	"slices"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

type Runner struct {
	ConfigFlags  *ConfigFlags
	handler      Handler
	KubeApi      kubernetes.Interface
	DiscoveryApi discovery.DiscoveryInterface
	DynamicApi   dynamic.Interface
}

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

func (r *Runner) GetAllServerResources() ([]schema.GroupVersionResource, error) {
	agrs, err := restmapper.GetAPIGroupResources(r.DiscoveryApi)
	if err != nil {
		return nil, fmt.Errorf("failed to get API group resources: %w", err)
	}
	result := make([]schema.GroupVersionResource, 0)
	for _, agr := range agrs {
		group := agr.Group.GroupVersionKind().Group
		version := agr.Group.PreferredVersion.Version
		for _, resource := range agr.VersionedResources[version] {
			if !slices.Contains(resource.Verbs, "list") {
				slog.Debug("skipping unlistable resource", "resource", resource.Name)
				continue
			}
			result = append(result, schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: resource.Name,
			})
		}
	}
	return result, nil
}

func (r *Runner) cobraRunE() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if r.handler == nil {
			panic("handler cannot be nil")
		}

		return r.runHandlerAndPrint(
			cmd.OutOrStdout(),
			HandlerArgs{
				Runner: *r,
				Cmd:    cmd,
				Args:   args,
			},
		)
	}
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
