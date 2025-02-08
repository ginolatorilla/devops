package kubectl_list_finalizers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
)

func NewCommand(apiFactory kube.DynamicApiFactory) *cobra.Command {
	return kube.
		NewTabularRunnerWithDiscoveryApi(apiFactory, listResourceUsers).
		ToCobraCommand(
			"kubectl-list_finalizers",
			"Lists all Kubernetes resources that have finalizers",
		)
}

func listResourceUsers(a kube.HandlerArgs) (metaV1.Table, error) {
	_, resources, err := a.DiscoveryApi.ServerGroupsAndResources()
	for _, resource := range resources {
		var newApiResources []metaV1.APIResource
		for _, apiResource := range resource.APIResources {
			if strings.Contains(apiResource.Name, "/") {
				continue
			}
			if !slices.Contains(apiResource.Verbs, "list") && !slices.Contains(apiResource.Verbs, "get") {
				continue
			}
			newApiResources = append(newApiResources, apiResource)
		}
		resource.APIResources = newApiResources
	}
	if err != nil {
		return metaV1.Table{}, fmt.Errorf("failed to get server resources: %w", err)
	}
	gvrs, err := discovery.GroupVersionResources(resources)
	if err != nil {
		return metaV1.Table{}, fmt.Errorf("failed to convert server API resource lists to group version resources: %w", err)
	}

	jq, err := gojq.Parse(".items[] | {kind: .kind, namespace: .metadata.namespace, name: .metadata.name, finalizers: .metadata.finalizers}")
	if err != nil {
		return metaV1.Table{}, fmt.Errorf("failed to parse JQ query: %w", err)
	}

	table := metaV1.Table{
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Kind", Type: "string"},
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Finalizers", Type: "string"},
		},
	}

	for gvr := range gvrs {
		resources, err := a.DynamicApi.Resource(gvr).Namespace(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
		if err != nil {
			if !errors.Is(err, fmt.Errorf("the server could not find the requested resource")) {
				slog.Warn("failed to list resources", "gvr", gvr, "error", err)
			}
			continue
		}

		rawJson, _ := resources.MarshalJSON()
		var resourcesAsJson map[string]any
		json.Unmarshal(rawJson, &resourcesAsJson)

		iter := jq.RunWithContext(a.Cmd.Context(), resourcesAsJson)
		for {
			untyped, ok := iter.Next()
			if !ok {
				break
			}

			rawJson, err := json.Marshal(untyped)
			if err != nil {
				panic(fmt.Errorf("failed to marshal untyped object to JSON: %w", err))
			}
			var object struct {
				Kind       string   `json:"kind"`
				Name       string   `json:"name"`
				Namespace  string   `json:"namespace"`
				Finalizers []string `json:"finalizers"`
			}
			err = json.Unmarshal(rawJson, &object)
			if err != nil {
				panic(err)
			}

			for _, finalizer := range object.Finalizers {
				table.Rows = append(table.Rows, metaV1.TableRow{
					Cells: []interface{}{
						object.Kind,
						object.Name,
						finalizer,
					},
					Object: runtime.RawExtension{Object: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"namespace": object.Namespace,
							},
						},
					}},
				})
			}
		}
	}
	return table, nil
}
