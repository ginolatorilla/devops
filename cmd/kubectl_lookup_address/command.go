package kubectl_lookup_address

import (
	"log/slog"
	"slices"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	services, err := client.Services(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list services", "error", err)
	}

	var objects []runtime.Object
	for _, service := range services.Items {
		if slices.Contains(service.Spec.ClusterIPs, ipAddress) {
			objects = append(objects, &service)
		}
		if slices.Contains(service.Spec.ExternalIPs, ipAddress) {
			objects = append(objects, &service)
		}
		if service.Spec.LoadBalancerIP == ipAddress {
			objects = append(objects, &service)
		}
	}

	pods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list pods", "error", err)
	}
	for _, pod := range pods.Items {
		if slices.ContainsFunc(pod.Status.PodIPs, func(p coreV1.PodIP) bool {
			return p.IP == ipAddress
		}) {
			objects = append(objects, &pod)
		}
	}

	nodes, err := client.Nodes().List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list nodes", "error", err)
	}
	for _, node := range nodes.Items {
		if slices.ContainsFunc(node.Status.Addresses, func(n coreV1.NodeAddress) bool {
			return n.Address == ipAddress
		}) {
			objects = append(objects, &node)
		}
	}

	if len(objects) == 0 {
		slog.Warn("not found", "address", ipAddress)
		return metaV1.Table{}, nil
	}

	table := metaV1.Table{
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Kind", Type: "string"},
			{Name: "Name", Type: "string", Format: "name"},
		},
		Rows: make([]metaV1.TableRow, len(objects)),
	}
	for i, object := range objects {
		gvk, err := apiutil.GVKForObject(object, scheme.Scheme)
		if err != nil {
			slog.Warn("cannot get group/version/kind of resource", "object", object)
			continue
		}
		table.Rows[i] = metaV1.TableRow{
			Cells: []interface{}{
				gvk.Kind,
				object.(metaV1.Object).GetName(),
			},
			Object: runtime.RawExtension{Object: object},
		}
	}
	return table, nil
}
