package kubectl_list_finalizers

import (
	"encoding/json"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
)

func NewCommand(apiFactory kube.DynamicApiFactory) *cobra.Command {
	configFlags := kube.NewConfigFlags()
	runner := kube.NewTabularRunnerWithDiscoveryApi(configFlags, apiFactory, listResourceUsers)

	command := &cobra.Command{
		Use:   "kubectl-list_finalizers",
		Short: "Lists all Kubernetes resources that have finalizers",
		Run:   runner.ToRun(),
	}

	configFlags.AddFlags(command.Flags())
	return command
}

func listResourceUsers(a kube.HandlerArgs) metaV1.Table {
	_, resources, err := a.DiscoveryApi.ServerGroupsAndResources()
	if err != nil {
		panic(err)
	}
	gvrs, err := discovery.GroupVersionResources(resources)
	if err != nil {
		panic(err)
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
			continue
		}

		jq, err := gojq.Parse(".items[] | {kind: .kind, namespace: .metadata.namespace, name: .metadata.name, finalizers: .metadata.finalizers}")
		if err != nil {
			panic(err)
		}
		raw, _ := resources.MarshalJSON()
		var v map[string]any
		json.Unmarshal(raw, &v)

		iter := jq.RunWithContext(a.Cmd.Context(), v)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}

			raw, err := json.Marshal(v)
			if err != nil {
				panic(err)
			}
			var v2 struct {
				Kind       string   `json:"kind"`
				Name       string   `json:"name"`
				Namespace  string   `json:"namespace"`
				Finalizers []string `json:"finalizers"`
			}
			err = json.Unmarshal(raw, &v2)
			if err != nil {
				panic(err)
			}

			for _, finalizer := range v2.Finalizers {
				table.Rows = append(table.Rows, metaV1.TableRow{
					Cells: []interface{}{
						v2.Kind,
						v2.Name,
						finalizer,
					},
					Object: runtime.RawExtension{Object: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"namespace": v2.Namespace,
							},
						},
					}},
				})
			}
		}
	}
	return table
}
