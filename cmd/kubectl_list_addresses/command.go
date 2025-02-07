package kubectl_list_addresses

import (
	"log/slog"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gocoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewCommand(apiFactory kube.ApiFactory) *cobra.Command {
	configFlags := kube.NewConfigFlags()
	runner := kube.NewTabularRunner(configFlags, apiFactory, listAddresses)

	command := &cobra.Command{
		Use:   "kubectl-list_addresses",
		Short: "Lists all IP addresses in the cluster",
		Run:   runner.ToRun(),
	}

	configFlags.AddFlags(command.Flags())
	return command
}

func listAddresses(a kube.HandlerArgs) metaV1.Table {
	client := a.KubeApi.CoreV1()

	table := metaV1.Table{
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

	return table
}

func getServiceAddresses(a kube.HandlerArgs, client gocoreV1.CoreV1Interface, table *metaV1.Table) error {
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

func getPodAddresses(a kube.HandlerArgs, client gocoreV1.CoreV1Interface, table *metaV1.Table) error {
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

func getNodeAddresses(a kube.HandlerArgs, client gocoreV1.CoreV1Interface, table *metaV1.Table) error {
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
