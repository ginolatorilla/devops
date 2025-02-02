package kubectl_list_addresses

import (
	"fmt"
	"log/slog"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	services, err := client.Services(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to list services: %w", err))
	}

	gvk, err := apiutil.GVKForObject(&services.Items[0], scheme.Scheme)
	if err != nil {
		slog.Warn("cannot get group/version/kind of resource", "object", &services.Items[0])
	}
	for _, service := range services.Items {
		for _, ip := range service.Spec.ClusterIPs {
			if ip == "" || ip == "None" {
				continue
			}
			table.Rows = append(table.Rows, metaV1.TableRow{
				Cells: []interface{}{
					gvk.Kind,
					service.Name,
					"ClusterIP",
					ip,
				},
				Object: runtime.RawExtension{Object: &service},
			})
		}

		for _, ip := range service.Spec.ExternalIPs {
			table.Rows = append(table.Rows, metaV1.TableRow{
				Cells: []interface{}{
					gvk.Kind,
					service.Name,
					"ExternalIP",
					ip,
				},
				Object: runtime.RawExtension{Object: &service},
			})
		}

		if service.Spec.LoadBalancerIP != "" {
			table.Rows = append(table.Rows, metaV1.TableRow{
				Cells: []interface{}{
					gvk.Kind,
					service.Name,
					"LoadBalancerIP",
					service.Spec.LoadBalancerIP,
				},
				Object: runtime.RawExtension{Object: &service},
			})
		}
	}

	pods, err := client.Pods(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list pods", "error", err)
	}
	gvk, err = apiutil.GVKForObject(&pods.Items[0], scheme.Scheme)
	if err != nil {
		slog.Warn("cannot get group/version/kind of resource", "object", &pods.Items[0])
	}
	for _, pod := range pods.Items {
		for _, ip := range pod.Status.PodIPs {
			table.Rows = append(table.Rows, metaV1.TableRow{
				Cells: []interface{}{
					gvk.Kind,
					pod.Name,
					"PodIP",
					ip.IP,
				},
				Object: runtime.RawExtension{Object: &pod},
			})
		}
	}

	nodes, err := client.Nodes().List(a.Cmd.Context(), metaV1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list nodes", "error", err)
	}
	gvk, err = apiutil.GVKForObject(&nodes.Items[0], scheme.Scheme)
	if err != nil {
		slog.Warn("cannot get group/version/kind of resource", "object", &nodes.Items[0])
	}
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == "Hostname" {
				continue
			}
			table.Rows = append(table.Rows, metaV1.TableRow{
				Cells: []interface{}{
					gvk.Kind,
					node.Name,
					address.Type,
					address.Address,
				},
				Object: runtime.RawExtension{Object: &node},
			})
		}
	}

	return table
}
