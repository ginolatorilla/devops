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
package kubectl_list_finalizers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

func NewCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunnerV2(listResourceUsers, append(runnerOpts, kubectlplugin.WithResourcePrinters())...).
		ToCobraCommand(
			"kubectl-list_finalizers",
			"Lists all Kubernetes resources that have finalizers",
		)
}

func listResourceUsers(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	gvrs, err := getAllGroupVersionResources(a)
	if err != nil {
		return nil, fmt.Errorf("failed to get all group version resources: %w", err)
	}

	jq, err := gojq.Parse(".items[] | {kind: .kind, namespace: .metadata.namespace, name: .metadata.name, finalizers: .metadata.finalizers}")
	if err != nil {
		return nil, fmt.Errorf("failed to parse JQ query: %w", err)
	}

	table := metaV1.Table{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "meta.k8s.io/v1",
			Kind:       "Table",
		},
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Kind", Type: "string"},
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Finalizers", Type: "string"},
		},
	}

	for _, gvr := range gvrs {
		resources, err := a.DynamicApi.Resource(gvr).Namespace(a.Namespace).List(a.Cmd.Context(), metaV1.ListOptions{})
		if err != nil {
			if !errors.Is(err, fmt.Errorf("the server could not find the requested resource")) {
				slog.Warn("failed to list resources", "gvr", gvr, "error", err)
			}
			continue
		}

		queryDynamicResource(a.Cmd.Context(), jq, resources, func(rawJson []byte) {
			var object objectWithFinalizers
			json.Unmarshal(rawJson, &object)
			recordObjectsWithFinalizers(&table, object)
		})
	}
	return &table, nil
}

func getAllGroupVersionResources(a kubectlplugin.HandlerArgs) ([]schema.GroupVersionResource, error) {
	_, resources, err := a.DiscoveryApi.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server resources: %w", err)
	}
	for _, resource := range resources {
		var filteredApiResources []metaV1.APIResource
		for _, apiResource := range resource.APIResources {
			if strings.Contains(apiResource.Name, "/") {
				continue
			}
			if !slices.Contains(apiResource.Verbs, "list") && !slices.Contains(apiResource.Verbs, "get") {
				continue
			}
			filteredApiResources = append(filteredApiResources, apiResource)
		}
		resource.APIResources = filteredApiResources
	}
	gvrs, err := discovery.GroupVersionResources(resources)
	if err != nil {
		return nil, fmt.Errorf("failed to convert server API resource lists to group version resources: %w", err)
	}
	gvrSlice := make([]schema.GroupVersionResource, len(gvrs))
	for gvr := range gvrs {
		gvrSlice = append(gvrSlice, gvr)
	}
	return gvrSlice, nil
}

type jsonSerializable interface {
	MarshalJSON() ([]byte, error)
}

func queryDynamicResource(ctx context.Context, jq *gojq.Query, object jsonSerializable, handler func(raw []byte)) {
	rawJson, _ := object.MarshalJSON()
	var jsonObject map[string]any
	json.Unmarshal(rawJson, &jsonObject)

	iter := jq.RunWithContext(ctx, jsonObject)
	for {
		typeErased, ok := iter.Next()
		if !ok {
			break
		}
		rawJson, err := json.Marshal(typeErased)
		if err != nil {
			panic(fmt.Errorf("failed to marshal untyped object to JSON: %w", err))
		}
		handler(rawJson)
	}
}

type objectWithFinalizers struct {
	Kind       string   `json:"kind"`
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Finalizers []string `json:"finalizers"`
}

func recordObjectsWithFinalizers(table *metaV1.Table, object objectWithFinalizers) {
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
