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
	"log/slog"
	"slices"

	kplug "github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

func main() {
	newCommand(kplug.WithDefaultKubeApi()).Execute()
}

func newCommand(opts ...kplug.RunnerOpts) *cobra.Command {
	opts = append(opts, kplug.WithTablePrinter(), kplug.WithAllNamespaces())
	return kplug.NewRunner(lookupAddress, opts...).
		ToCobraCommand(
			"kubectl-lookup_address",
			"Finds Kubernetes resources by IP address",
			kplug.WithArgs(cobra.ExactArgs(1), []string{"ip-address"}),
		)
}

func lookupAddress(a kplug.HandlerArgs) (runtime.Object, error) {
	ipAddr := a.Args[0]
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
		)
	a.ConfigFlags.WithAll(true)
	for _, d := range []struct {
		kind        string
		visitorFunc resource.VisitorFunc
	}{
		{"service", addServiceWithAddressToTable(tbuild, ipAddr)},
		{"pod", addPodWithAddressToTable(tbuild, ipAddr)},
		{"node", addNodeWithAddressToTable(tbuild, ipAddr)},
	} {
		if err := a.ToResourceFinder(d.kind).Do().Visit(d.visitorFunc); err != nil {
			slog.Warn("failed to list resources", "kind", d.kind, "error", err)
		}
	}
	return &tbuild.Table, nil
}

func addServiceWithAddressToTable(tbuild *kplug.TableBuilder, ipAddress string) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		service := kplug.As[*coreV1.Service](info.Object)
		if slices.Contains(service.Spec.ClusterIPs, ipAddress) {
			tbuild.AddRow(service, map[string]any{"Type": "ClusterIP", "Kind": "Service"})
		}
		if slices.Contains(service.Spec.ExternalIPs, ipAddress) {
			tbuild.AddRow(service, map[string]any{"Type": "ExternalIP", "Kind": "Service"})
		}
		if service.Spec.LoadBalancerIP == ipAddress {
			tbuild.AddRow(service, map[string]any{"Type": "LoadBalancerIP", "Kind": "Service"})
		}
		return nil
	}
}

func addPodWithAddressToTable(tbuild *kplug.TableBuilder, ipAddress string) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		pod := kplug.As[*coreV1.Pod](info.Object)
		if slices.ContainsFunc(pod.Status.PodIPs, func(p coreV1.PodIP) bool {
			return p.IP == ipAddress
		}) {
			tbuild.AddRow(pod, map[string]any{"Type": "PodIP", "Kind": "Pod"})
		}
		return nil
	}
}

func addNodeWithAddressToTable(tbuild *kplug.TableBuilder, ipAddress string) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		node := kplug.As[*coreV1.Node](info.Object)
		for _, n := range node.Status.Addresses {
			if n.Address == ipAddress {
				tbuild.AddRow(node, map[string]any{"Type": n.Type, "Kind": "Node"})
			}
		}
		return nil
	}
}
