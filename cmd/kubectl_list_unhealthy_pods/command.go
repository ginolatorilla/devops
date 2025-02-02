package kubectl_list_unhealthy_pods

import (
	"log/slog"
	"slices"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCommand(apiFactory kube.ApiFactory) *cobra.Command {
	configFlags := kube.NewConfigFlags()
	runner := kube.NewTabularRunner(configFlags, apiFactory, listUnhealthyPods)

	command := &cobra.Command{
		Use:   "kubectl-list_unhealthy_pods",
		Short: "Finds Kubernetes pods that are in a failed or unknown state",
		Run:   runner.ToRun(),
	}

	configFlags.AddFlags(command.Flags())
	return command
}

func listUnhealthyPods(a kube.HandlerArgs) metaV1.Table {
	client := a.KubeApi.CoreV1()

	var pods []coreV1.Pod
	allPods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	for _, pod := range allPods.Items {
		if pod.Status.Phase == coreV1.PodFailed {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated != nil {
					pod.Status.Reason = cs.State.Terminated.Reason
				}
			}
			pods = append(pods, pod)
			continue
		}
		if pod.Status.Phase == coreV1.PodPending {
			if slices.ContainsFunc(pod.Status.ContainerStatuses, func(cs coreV1.ContainerStatus) bool {
				if cs.State.Waiting != nil {
					pod.Status.Reason = cs.State.Waiting.Reason
					return cs.State.Waiting.Reason != "ContainerCreating" && cs.State.Waiting.Reason != "PodInitializing"
				}
				return false
			}) {
				pods = append(pods, pod)
				continue
			}
		}
	}

	if err != nil {
		slog.Warn("failed to list pods", "error", err)
	}

	if len(pods) == 0 {
		return metaV1.Table{}
	}

	table := metaV1.Table{
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Status", Type: "string"},
			{Name: "Reason", Type: "string"},
		},
		Rows: make([]metaV1.TableRow, len(pods)),
	}
	for i, pod := range pods {
		if err != nil {
			slog.Warn("cannot get group/version/kind of resource", "object", pod)
			continue
		}
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
