package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	kubetesting "github.com/ginolatorilla/devops/pkg/kubectlplugin/testing"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)

	t.Run("OK", func(t *testing.T) {
		client := fake.NewClientset()
		loadResources(t, client, "test")

		cmd := newCommand(kubectlplugin.WithKubeApi(client))
		cmd.SetArgs([]string{"-n", "test"})
		assert.NoError(cmd.Execute())
	})

	for _, resource := range []string{"services", "pods", "nodes"} {
		t.Run(
			fmt.Sprintf("SkipIf%sUnreadable", cases.Title(language.English).String(resource)),
			func(t *testing.T) {
				client := fake.NewClientset()
				kubetesting.LoadCannedError(t, client, "list", resource)
				loadResources(t, client, "test")

				cmd := newCommand(kubectlplugin.WithKubeApi(client))
				cmd.SetArgs([]string{"-n", "test"})
				assert.NoError(cmd.Execute())
			})
	}
}

func loadResources(t *testing.T, client kubernetes.Interface, namespace string) {
	t.Helper()

	for name, spec := range map[string]coreV1.ServiceSpec{
		"empty-cluster-ip": {ClusterIPs: []string{}},
		"none-cluster-ip":  {ClusterIPs: []string{"None"}},
		"cluster-ip":       {ClusterIPs: []string{"1.2.3.4"}},
		"external-ip":      {ExternalIPs: []string{"1.2.3.4"}},
		"loadbalancer-ip":  {LoadBalancerIP: "1.2.3.4"},
	} {
		if err := createService(client, namespace, name, spec); err != nil {
			t.Fatal(err)
		}
	}

	if err := createPod(client, namespace, "test", "1.2.3.4"); err != nil {
		t.Fatal(err)
	}

	for name, nodeAddress := range map[string]coreV1.NodeAddress{
		"hostname":     {Type: coreV1.NodeHostName, Address: "example.com"},
		"internal-ip":  {Type: coreV1.NodeInternalIP, Address: "1.2.3.4"},
		"external-ip":  {Type: coreV1.NodeExternalIP, Address: "1.2.3.4"},
		"internal-dns": {Type: coreV1.NodeInternalDNS, Address: "example.com"},
		"external-dns": {Type: coreV1.NodeExternalDNS, Address: "example.com"},
	} {
		if err := createNode(client, name, nodeAddress); err != nil {
			t.Fatal(err)
		}
	}
}

func createService(client kubernetes.Interface, namespace, name string, spec coreV1.ServiceSpec) error {
	if _, err := client.CoreV1().Services(namespace).Create(
		context.TODO(),
		&coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{Name: name},
			Spec:       spec,
		},
		metaV1.CreateOptions{},
	); err != nil {
		return err
	}
	return nil
}

func createPod(client kubernetes.Interface, namespace, name, podIP string) error {
	if _, err := client.CoreV1().Pods(namespace).Create(
		context.TODO(),
		&coreV1.Pod{
			ObjectMeta: metaV1.ObjectMeta{Name: name},
			Status:     coreV1.PodStatus{PodIPs: []coreV1.PodIP{{IP: podIP}}},
		},
		metaV1.CreateOptions{},
	); err != nil {
		return err
	}
	return nil
}

func createNode(client kubernetes.Interface, name string, nodeAddress coreV1.NodeAddress) error {
	if _, err := client.CoreV1().Nodes().Create(
		context.TODO(),
		&coreV1.Node{
			ObjectMeta: metaV1.ObjectMeta{Name: name},
			Status:     coreV1.NodeStatus{Addresses: []coreV1.NodeAddress{nodeAddress}},
		},
		metaV1.CreateOptions{},
	); err != nil {
		return err
	}
	return nil
}
