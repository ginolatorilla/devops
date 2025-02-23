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
	"log/slog"

	kplug "github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

func main() {
	newCommand(kplug.WithDefaultDiscoveryApi()).Execute()
}

func newCommand(opts ...kplug.RunnerOpts) *cobra.Command {
	opts = append(opts, kplug.WithAllNamespaces())
	return kplug.NewRunner(listFinalizers, opts...).
		ToCobraCommand(
			"kubectl-list_finalizers",
			"Lists all Kubernetes resources that have finalizers",
		)
}

func listFinalizers(a kplug.HandlerArgs) (runtime.Object, error) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{
				Name:        "Finalizers",
				Description: "The finalizers attached to the resource",
			})
	a.ConfigFlags.WithAll(true).WithScheme(nil)
	gvrs, err := a.GetAllServerResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server resources: %w", err)
	}
	for _, gvr := range gvrs {
		err := a.ToResourceFinder(gvr.Resource).Do().
			Visit(addResourceWithFinalizerToTable(tbuild))
		if err != nil {
			slog.Warn("failed to find resources with finalizers",
				"gvr", gvr,
				"error", err)
		}
	}
	return &tbuild.Table, nil
}

func addResourceWithFinalizerToTable(tbuild *kplug.TableBuilder) resource.VisitorFunc {
	return func(info *resource.Info, err error) error {
		uo := kplug.As[*unstructured.Unstructured](info.Object)
		finalizers := uo.GetFinalizers()
		if len(finalizers) == 0 {
			return nil
		}
		tbuild.AddRow(
			info.Object, map[string]any{
				"Kind":       uo.GetKind(),
				"Finalizers": finalizers,
			})
		return nil
	}
}
