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
package kubectl_list_addresses

import (
	"log/slog"

	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gocoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunnerV2(listAddresses, append(runnerOpts, kubectlplugin.WithResourcePrinters())...).
		ToCobraCommand(
			"kubectl-list_addresses",
			"Lists all IP addresses in the cluster",
		)
}

func listAddresses(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	client := a.KubeApi.CoreV1()

	table := metaV1.Table{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "meta.k8s.io/v1",
			Kind:       "Table",
		},
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Kind", Type: "string"},
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Type", Type: "string"},
			{Name: "Address", Type: "string"},
		},
	}

	if err := getServiceAddresses(a, client, &table); err != nil {
		slog.Warn("failed to list services", "error", err)
	}

	if err := getPodAddresses(a, client, &table); err != nil {
		slog.Warn("failed to list pods", "error", err)
	}

	if err := getNodeAddresses(a, client, &table); err != nil {
		slog.Warn("failed to list nodes", "error", err)
	}

	return &table, nil
}

func getServiceAddresses(a kubectlplugin.HandlerArgs, client gocoreV1.CoreV1Interface, table *metaV1.Table) error {
	services, err := client.Services(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}

	for _, service := range services.Items {
		for _, ip := range service.Spec.ClusterIPs {
			if ip == "" || ip == "None" {
				continue
			}
			table.Rows = append(
				table.Rows,
				newTableRow("Service", service.Name, "ClusterIP", ip, &service),
			)
		}

		for _, ip := range service.Spec.ExternalIPs {
			table.Rows = append(
				table.Rows,
				newTableRow("Service", service.Name, "ExternalIP", ip, &service),
			)
		}

		if service.Spec.LoadBalancerIP != "" {
			table.Rows = append(
				table.Rows,
				newTableRow("Service", service.Name, "LoadBalancerIP", service.Spec.LoadBalancerIP, &service),
			)
		}
	}
	return nil
}

func getPodAddresses(a kubectlplugin.HandlerArgs, client gocoreV1.CoreV1Interface, table *metaV1.Table) error {
	pods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		for _, ip := range pod.Status.PodIPs {
			table.Rows = append(
				table.Rows,
				newTableRow("Pod", pod.Name, "PodIP", ip.IP, &pod),
			)
		}
	}
	return nil
}

func getNodeAddresses(a kubectlplugin.HandlerArgs, client gocoreV1.CoreV1Interface, table *metaV1.Table) error {
	nodes, err := client.Nodes().List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type != coreV1.NodeInternalIP && address.Type != coreV1.NodeExternalIP {
				continue
			}
			table.Rows = append(
				table.Rows,
				newTableRow("Node", node.Name, string(address.Type), address.Address, &node),
			)
		}
	}
	return nil
}

func newTableRow(kind, name, addressType, address string, object runtime.Object) metaV1.TableRow {
	return metaV1.TableRow{
		Cells:  []interface{}{kind, name, addressType, address},
		Object: runtime.RawExtension{Object: object},
	}
}
