package kubectl_list_unhealthy_pods

import (
	"fmt"
	"slices"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCommand(apiFactory kube.ApiFactory) *cobra.Command {
	return kube.
		NewTabularRunner(apiFactory, listUnhealthyPods).
		ToCobraCommand(
			"kubectl-list_unhealthy_pods",
			"Finds Kubernetes pods that are in a failed or unknown state",
		)
}

func listUnhealthyPods(a kube.HandlerArgs) (metaV1.Table, error) {
	client := a.KubeApi.CoreV1()
	pods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return metaV1.Table{}, fmt.Errorf("failed to list pods: %w", err)
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

func podsToTable(pods []coreV1.Pod) metaV1.Table {
	table := metaV1.Table{
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
	return table
}
