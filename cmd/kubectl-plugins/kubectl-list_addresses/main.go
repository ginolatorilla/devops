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

	devopscmd "github.com/ginolatorilla/devops/cmd/devops/cmd"
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
	return kplug.NewRunner(listAddresses, opts...).
		ToCobraCommand(
			"kubectl-list_addresses",
			"Lists all IP addresses in the cluster",
			kplug.WithVersion(devopscmd.Version+"-"+devopscmd.CommitHash),
		)
}

func listAddresses(a kplug.HandlerArgs) (runtime.Object, error) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
			kplug.Column{Name: "Address", Description: "The IP address"},
		)
	a.ConfigFlags.WithAll(true)
	for _, d := range []struct {
		kind        string
		visitorFunc resource.VisitorFunc
	}{
		{"service", addServicesToTable(tbuild)},
		{"pod", addPodsToTable(tbuild)},
		{"node", addNodesToTable(tbuild)},
	} {
		err := a.ToResourceFinder(d.kind).
			Do().
			Visit(d.visitorFunc)
		if err != nil {
			slog.Warn("failed to list resources", "kind", d.kind, "error", err)
		}
	}
	return &tbuild.Table, nil
}

func addServicesToTable(tableBuilder *kplug.TableBuilder) resource.VisitorFunc {
	return func(info *resource.Info, err error) error {
		service := kplug.As[*coreV1.Service](info.Object)
		for _, ip := range service.Spec.ClusterIPs {
			if ip == "" || ip == "None" {
				continue
			}
			tableBuilder.AddRow(service, map[string]any{
				"Kind":    "Service",
				"Type":    "ClusterIP",
				"Address": ip,
			})
		}
		for _, ip := range service.Spec.ExternalIPs {
			tableBuilder.AddRow(service, map[string]any{
				"Kind":    "Service",
				"Type":    "ExternalIP",
				"Address": ip,
			})
		}
		if service.Spec.LoadBalancerIP != "" {
			tableBuilder.AddRow(service, map[string]any{
				"Kind":    "Service",
				"Type":    "LoadBalancerIP",
				"Address": service.Spec.LoadBalancerIP,
			})
		}
		return nil
	}
}

func addPodsToTable(tbuild *kplug.TableBuilder) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		pod := kplug.As[*coreV1.Pod](info.Object)
		for _, ip := range pod.Status.PodIPs {
			tbuild.AddRow(pod, map[string]any{
				"Kind":    "Pod",
				"Type":    "PodIP",
				"Address": ip.IP,
			})
		}
		return nil
	}
}

func addNodesToTable(tableBuilder *kplug.TableBuilder) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		node := kplug.As[*coreV1.Node](info.Object)
		for _, address := range node.Status.Addresses {
			if address.Type != coreV1.NodeInternalIP &&
				address.Type != coreV1.NodeExternalIP {
				continue
			}
			tableBuilder.AddRow(node, map[string]any{
				"Kind":    "Node",
				"Type":    address.Type,
				"Address": address.Address,
			})
		}
		return nil
	}
}
