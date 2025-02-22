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
	"os"
	"slices"

	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

func main() {
	if err := newCommand(kubectlplugin.WithDefaultKubeApi()).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunner(lookupAddress,
			append(runnerOpts,
				kubectlplugin.WithTablePrinter(),
				kubectlplugin.WithAllNamespaces(),
			)...).
		ToCobraCommand(
			"kubectl-lookup_address",
			"Finds Kubernetes resources by IP address",
			kubectlplugin.WithArgs(cobra.ExactArgs(1), []string{"ip-address"}),
		)
}

func lookupAddress(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	ipAddress := a.Args[0]
	tableBuilder := kubectlplugin.NewTableBuilder().
		AdditionalColumns(
			kubectlplugin.ResourceKindColumn,
			kubectlplugin.Column{Name: "Type", Description: "The type of address"},
		)
	a.ConfigFlags.WithAll(true)
	for _, d := range []struct {
		kind        string
		visitorFunc resource.VisitorFunc
	}{
		{"service", findAddressInService(tableBuilder, ipAddress)},
		{"pod", findAddressInPod(tableBuilder, ipAddress)},
		{"node", findAddressInNode(tableBuilder, ipAddress)},
	} {
		if err := a.ToResourceFinder(d.kind).Do().Visit(d.visitorFunc); err != nil {
			slog.Warn("failed to list resources", "kind", d.kind, "error", err)
		}
	}
	return &tableBuilder.Table, nil
}

func findAddressInService(tableBuilder *kubectlplugin.TableBuilder, ipAddress string) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		service := kubectlplugin.As[*coreV1.Service](info.Object)
		if slices.Contains(service.Spec.ClusterIPs, ipAddress) {
			tableBuilder.AddRow(service, map[string]any{"Type": "ClusterIP", "Kind": "Service"})
		}
		if slices.Contains(service.Spec.ExternalIPs, ipAddress) {
			tableBuilder.AddRow(service, map[string]any{"Type": "ExternalIP", "Kind": "Service"})
		}
		if service.Spec.LoadBalancerIP == ipAddress {
			tableBuilder.AddRow(service, map[string]any{"Type": "LoadBalancerIP", "Kind": "Service"})
		}
		return nil
	}
}

func findAddressInPod(tableBuilder *kubectlplugin.TableBuilder, ipAddress string) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		pod := kubectlplugin.As[*coreV1.Pod](info.Object)
		if slices.ContainsFunc(pod.Status.PodIPs, func(p coreV1.PodIP) bool {
			return p.IP == ipAddress
		}) {
			tableBuilder.AddRow(pod, map[string]any{"Type": "PodIP", "Kind": "Pod"})
		}
		return nil
	}
}

func findAddressInNode(tableBuilder *kubectlplugin.TableBuilder, ipAddress string) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		node := kubectlplugin.As[*coreV1.Node](info.Object)
		for _, n := range node.Status.Addresses {
			if n.Address == ipAddress {
				tableBuilder.AddRow(node, map[string]any{"Type": n.Type, "Kind": "Node"})
			}
		}
		return nil
	}
}
