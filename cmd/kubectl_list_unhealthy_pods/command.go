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
package kubectl_list_unhealthy_pods

import (
	"fmt"
	"slices"

	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunner(listUnhealthyPods, append(runnerOpts, kubectlplugin.WithResourcePrinters())...).
		ToCobraCommand(
			"kubectl-list_unhealthy_pods",
			"Finds Kubernetes pods that are in a failed or unknown state",
		)
}

func listUnhealthyPods(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	client := a.KubeApi.CoreV1()
	pods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	unhealthyPods := filterUnhealthyPods(pods.Items)
	return podsToTable(unhealthyPods), nil
}

func filterUnhealthyPods(pods []coreV1.Pod) []coreV1.Pod {
	var filtered []coreV1.Pod
	for _, pod := range pods {
		if pod.Status.Phase == coreV1.PodFailed {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated != nil {
					pod.Status.Reason = cs.State.Terminated.Reason
				}
			}
			filtered = append(filtered, pod)
			continue
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
				filtered = append(filtered, pod)
				continue
			}
		}
	}
	return filtered
}

func podsToTable(pods []coreV1.Pod) runtime.Object {
	table := metaV1.Table{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "meta.k8s.io/v1",
			Kind:       "Table",
		},
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Status", Type: "string"},
			{Name: "Reason", Type: "string"},
		},
		Rows: make([]metaV1.TableRow, len(pods)),
	}
	for i, pod := range pods {
		table.Rows[i] = metaV1.TableRow{
			Cells: []interface{}{
				pod.Name,
				pod.Status.Phase,
				pod.Status.Reason,
			},
			Object: runtime.RawExtension{Object: &pod},
		}
	}
	return &table
}
