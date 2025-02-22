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
		NewRunner(listUnhealthyPods,
			append(runnerOpts,
				kubectlplugin.WithTablePrinter(),
				kubectlplugin.WithAllNamespaces(),
			)...).
		ToCobraCommand(
			"kubectl-list_unhealthy_pods",
			"Finds Kubernetes pods that are in a failed or unknown state",
		)
}

func listUnhealthyPods(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	tableBuilder := kubectlplugin.NewTableBuilder().AdditionalColumns(
		kubectlplugin.Column{Name: "Status", Description: "The status of the pod"},
		kubectlplugin.Column{Name: "Reason", Description: "The reason for the pod's status"},
	)

	a.ConfigFlags.WithAll(true)
	if err := a.ToResourceFinder("pods").Do().Visit(getUnhealthyPods(tableBuilder)); err != nil {
		return nil, fmt.Errorf("failed to list unhealthy pods: %w", err)
	}
	return &tableBuilder.Table, nil
}

func getUnhealthyPods(tableBuilder *kubectlplugin.TableBuilder) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		pod := kubectlplugin.As[*coreV1.Pod](info.Object)
		if pod.Status.Phase == coreV1.PodFailed {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated != nil {
					pod.Status.Reason = cs.State.Terminated.Reason
				}
			}
			tableBuilder.AddRow(pod, map[string]any{
				"Status": pod.Status.Phase,
				"Reason": pod.Status.Reason,
			})
			return nil
		}

		if pod.Status.Phase == coreV1.PodPending {
			if slices.ContainsFunc(
				pod.Status.ContainerStatuses,
				func(cs coreV1.ContainerStatus) bool {
					if cs.State.Waiting != nil {
						pod.Status.Reason = cs.State.Waiting.Reason
						return cs.State.Waiting.Reason != "ContainerCreating" && cs.State.Waiting.Reason != "PodInitializing"
					}
					return false
				},
			) {
				tableBuilder.AddRow(pod, map[string]any{
					"Status": pod.Status.Phase,
					"Reason": pod.Status.Reason,
				})
				return nil
			}
		}
		return nil
	}
}
