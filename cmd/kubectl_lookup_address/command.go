package kubectl_lookup_address

import (
	"log/slog"
	"slices"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gocoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewCommand(apiFactory kube.ApiFactory) *cobra.Command {
	return kube.
		NewTabularRunner(apiFactory, lookupAddress).
		ToCobraCommandWithArgs(
			"kubectl-lookup_address",
			"Finds Kubernetes resources by IP address",
			cobra.ExactArgs(1),
			[]string{"ip-address"},
		)
}

func lookupAddress(a kube.HandlerArgs) (metaV1.Table, error) {
	ipAddress := a.Args[0]
	client := a.KubeApi.CoreV1()

	table := metaV1.Table{
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Kind", Type: "string"},
			{Name: "Name", Type: "string", Format: "name"},
		},
		Rows: []metaV1.TableRow{},
	}

	if err := recordServicesWithMatchingAddress(a, client, ipAddress, &table); err != nil {
		slog.Warn("failed to list services", "error", err)
	}
	if err := recordPodsWithMatchingAddress(a, client, ipAddress, &table); err != nil {
		slog.Warn("failed to list pods", "error", err)
	}
	if err := recordNodesWithMatchingAddress(a, client, ipAddress, &table); err != nil {
		slog.Warn("failed to list nodes", "error", err)
	}
	return table, nil
}

func recordServicesWithMatchingAddress(a kube.HandlerArgs, client gocoreV1.CoreV1Interface, ipAddress string, table *metaV1.Table) error {
	services, err := client.Services(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		if slices.Contains(service.Spec.ClusterIPs, ipAddress) {
			table.Rows = append(table.Rows, newTableRow("Service", service.Name, &service))
		}
		if slices.Contains(service.Spec.ExternalIPs, ipAddress) {
			table.Rows = append(table.Rows, newTableRow("Service", service.Name, &service))
		}
		if service.Spec.LoadBalancerIP == ipAddress {
			table.Rows = append(table.Rows, newTableRow("Service", service.Name, &service))
		}
	}
	return nil
}

func recordPodsWithMatchingAddress(a kube.HandlerArgs, client gocoreV1.CoreV1Interface, ipAddress string, table *metaV1.Table) error {
	pods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if slices.ContainsFunc(pod.Status.PodIPs, func(p coreV1.PodIP) bool {
			return p.IP == ipAddress
		}) {
			table.Rows = append(table.Rows, newTableRow("Pod", pod.Name, &pod))
		}
	}
	return nil
}

func recordNodesWithMatchingAddress(a kube.HandlerArgs, client gocoreV1.CoreV1Interface, ipAddress string, table *metaV1.Table) error {
	nodes, err := client.Nodes().List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		if slices.ContainsFunc(node.Status.Addresses, func(n coreV1.NodeAddress) bool {
			return n.Address == ipAddress
		}) {
			table.Rows = append(table.Rows, newTableRow("Node", node.Name, &node))
		}
	}
	return nil
}

func newTableRow(kind, name string, object runtime.Object) metaV1.TableRow {
	return metaV1.TableRow{
		Cells:  []interface{}{kind, name},
		Object: runtime.RawExtension{Object: object},
	}
}
