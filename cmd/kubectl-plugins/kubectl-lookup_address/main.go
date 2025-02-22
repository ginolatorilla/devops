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
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gocoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func main() {
	if err := newCommand(kubectlplugin.WithDefaultKubeApi()).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunner(lookupAddress, append(runnerOpts, kubectlplugin.WithTablePrinter())...).
		ToCobraCommand(
			"kubectl-lookup_address",
			"Finds Kubernetes resources by IP address",
			kubectlplugin.WithArgs(cobra.ExactArgs(1), []string{"ip-address"}),
		)
}

func lookupAddress(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	ipAddress := a.Args[0]
	client := a.KubeApi.CoreV1()

	tableBuilder := kubectlplugin.
		NewTableBuilder().
		AdditionalColumns(
			kubectlplugin.ResourceKindColumn,
			kubectlplugin.Column{Name: "Type", Description: "The type of address"},
		)

	if err := recordServicesWithMatchingAddress(a, client, ipAddress, tableBuilder); err != nil {
		slog.Warn("failed to list services", "error", err)
	}
	if err := recordPodsWithMatchingAddress(a, client, ipAddress, tableBuilder); err != nil {
		slog.Warn("failed to list pods", "error", err)
	}
	if err := recordNodesWithMatchingAddress(a, client, ipAddress, tableBuilder); err != nil {
		slog.Warn("failed to list nodes", "error", err)
	}
	return &tableBuilder.Table, nil
}

func recordServicesWithMatchingAddress(
	a kubectlplugin.HandlerArgs,
	client gocoreV1.CoreV1Interface,
	ipAddress string,
	tableBuilder *kubectlplugin.TableBuilder,
) error {
	services, err := client.Services(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		if slices.Contains(service.Spec.ClusterIPs, ipAddress) {
			tableBuilder.AddRow(&service, map[string]any{"Type": "ClusterIP", "Kind": "Service"})
		}
		if slices.Contains(service.Spec.ExternalIPs, ipAddress) {
			tableBuilder.AddRow(&service, map[string]any{"Type": "ExternalIP", "Kind": "Service"})
		}
		if service.Spec.LoadBalancerIP == ipAddress {
			tableBuilder.AddRow(&service, map[string]any{"Type": "LoadBalancerIP", "Kind": "Service"})
		}
	}
	return nil
}

func recordPodsWithMatchingAddress(
	a kubectlplugin.HandlerArgs,
	client gocoreV1.CoreV1Interface,
	ipAddress string,
	tableBuilder *kubectlplugin.TableBuilder,
) error {
	pods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if slices.ContainsFunc(pod.Status.PodIPs, func(p coreV1.PodIP) bool {
			return p.IP == ipAddress
		}) {
			tableBuilder.AddRow(&pod, map[string]any{"Type": "PodIP", "Kind": "Pod"})
		}
	}
	return nil
}

func recordNodesWithMatchingAddress(
	a kubectlplugin.HandlerArgs,
	client gocoreV1.CoreV1Interface,
	ipAddress string,
	tableBuilder *kubectlplugin.TableBuilder,
) error {
	nodes, err := client.Nodes().List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		for _, n := range node.Status.Addresses {
			if n.Address == ipAddress {
				tableBuilder.AddRow(&node, map[string]any{"Type": n.Type, "Kind": "Node"})
			}
		}
	}
	return nil
}
