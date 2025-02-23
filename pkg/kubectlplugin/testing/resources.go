package testing

import (
	"context"
	"fmt"
	gotesting "testing"

	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

const Recurse = true
const DontEnforceNamespace = false

func LoadManifestToCluster(t *gotesting.T, manifestPath string) {
	config, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	require.NoError(t, err, "Could not get Kubernetes config. A live cluster is required to run this test.")

	discoveryApi := discovery.NewDiscoveryClientForConfigOrDie(config)
	groupResources, err := restmapper.GetAPIGroupResources(discoveryApi)
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	client := dynamic.NewForConfigOrDie(config)
	require.NoError(t, err, "Could not load Kubernetes config. A live cluster is required to run this test.")
	err = resource.NewLocalBuilder().
		Unstructured().
		ContinueOnError().
		FilenameParam(DontEnforceNamespace, &resource.FilenameOptions{Recursive: Recurse, Filenames: []string{manifestPath}}).
		Flatten().
		Do().
		Visit(func(info *resource.Info, _ error) error {
			object, ok := info.Object.(*unstructured.Unstructured)
			require.True(t, ok, "info.Object is not of type unstructured.Unstructured")
			rm, _ := mapper.RESTMapping(object.GroupVersionKind().GroupKind())
			_, err := client.Resource(rm.Resource).
				Namespace(object.GetNamespace()).
				Create(context.Background(), object, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create resource %s/%s: %w", info.Namespace, info.Name, err)
			}
			return nil
		})
	require.NoError(t, err, "Failed to apply manifest to cluster")
}
