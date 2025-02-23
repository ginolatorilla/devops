package main

import (
	"testing"

	kplug "github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)

	// for _, ip := range []string{
	// 	"1.1.1.1",
	// 	"2.2.2.2",
	// 	"3.3.3.3",
	// 	"4.4.4.4",
	// 	"5.5.5.5",
	// 	"6.6.6.6",
	// } {

	t.Run("OK", func(t *testing.T) {
		cmd := newCommand(kplug.WithDefaultKubeApi())
		cmd.SetArgs([]string{"-A", "172.16.0.1"})
		assert.NoError(cmd.Execute())
	})
	// }

	// t.Run("ErrorIfNotEnoughArgs", func(t *testing.T) {
	// 	client := fake.NewClientset()
	// 	loadResources(t, client, "test")

	// 	cmd := newCommand(kubectlplugin.WithKubeApi(client))
	// 	cmd.SetArgs([]string{})
	// 	assert.Error(cmd.Execute())
	// })

	// for _, resource := range []string{"services", "pods", "nodes"} {
	// 	t.Run(
	// 		fmt.Sprintf("SkipIf%sUnreadable", cases.Title(language.English).String(resource)),
	// 		func(t *testing.T) {
	// 			client := fake.NewClientset()
	// 			kubetesting.LoadCannedError(t, client, "list", resource)
	// 			loadResources(t, client, "test")

	// 			cmd := newCommand(kubectlplugin.WithKubeApi(client))
	// 			cmd.SetArgs([]string{"-n", "test", "1.1.1.1"})
	// 			assert.NoError(cmd.Execute())
	// 		})
	// }
}

func Test_addServiceWithAddressToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
		)
	visitor := addServiceWithAddressToTable(tbuild, "10.0.0.10")
	service := &coreV1.Service{
		Spec: coreV1.ServiceSpec{
			ClusterIPs:     []string{"10.0.0.10"},
			ExternalIPs:    []string{"10.0.0.10"},
			LoadBalancerIP: "10.0.0.10",
		},
	}
	assert.NoError(t, visitor(&resource.Info{Object: service}, nil))
}

func Test_addPodWithAddressToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
		)
	visitor := addPodWithAddressToTable(tbuild, "10.0.0.10")
	pod := &coreV1.Pod{
		Status: coreV1.PodStatus{
			PodIPs: []coreV1.PodIP{
				{IP: "10.0.0.10"},
			},
		},
	}
	assert.NoError(t, visitor(&resource.Info{Object: pod}, nil))
}

func Test_addNodeWithAddressToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
		)
	visitor := addNodeWithAddressToTable(tbuild, "10.0.0.10")
	node := &coreV1.Node{
		Status: coreV1.NodeStatus{
			Addresses: []coreV1.NodeAddress{
				{
					Type:    coreV1.NodeInternalIP,
					Address: "10.0.0.10",
				},
			},
		},
	}
	assert.NoError(t, visitor(&resource.Info{Object: node}, nil))
}
