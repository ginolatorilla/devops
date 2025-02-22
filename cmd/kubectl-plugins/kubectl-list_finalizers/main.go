// Copyright © 2025 Gino Latorilla
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package main

import (
	"fmt"
	"os"

	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

func main() {
	if err := newCommand(kubectlplugin.WithDefaultDiscoveryApi()).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunner(findResourcesWithFinalizers, append(runnerOpts,
			kubectlplugin.WithTablePrinter(),
			kubectlplugin.WithAllNamespaces(),
		)...).
		ToCobraCommand(
			"kubectl-list_finalizers",
			"Lists all Kubernetes resources that have finalizers",
		)
}

func findResourcesWithFinalizers(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	tableBuilder := kubectlplugin.NewTableBuilder().AdditionalColumns(
		kubectlplugin.ResourceKindColumn,
		kubectlplugin.Column{Name: "Finalizers", Description: "The finalizers attached to the resource"},
	)
	apiResourceList, err := a.DiscoveryApi.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server preferred resources: %w", err)
	}
	a.ConfigFlags.WithAll(true).WithScheme(nil)
	for _, ar := range apiResourceList {
		for _, r := range ar.APIResources {
			if err := a.ToResourceFinder(r.Name).Do().Visit(func(i *resource.Info, err error) error {
				uo := kubectlplugin.As[*unstructured.Unstructured](i.Object)
				finalizers := uo.GetFinalizers()
				if len(finalizers) == 0 {
					return nil
				}
				tableBuilder.AddRow(
					i.Object, map[string]any{
						"Kind":       r.Kind,
						"Finalizers": finalizers,
					})
				return nil
			}); err != nil {
				return nil, fmt.Errorf("failed to find resources with finalizers: %w", err)
			}
		}
	}
	return &tableBuilder.Table, nil
}
