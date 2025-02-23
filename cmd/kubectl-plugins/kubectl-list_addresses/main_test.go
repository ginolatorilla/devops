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

	t.Run("OK", func(t *testing.T) {
		t.SkipNow()
		cmd := newCommand(kplug.WithDefaultKubeApi())
		cmd.SetArgs([]string{"-A"})
		assert.NoError(cmd.Execute())
	})
}

func Test_addServicesToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
			kplug.Column{Name: "Address", Description: "The IP address"},
		)
	visitor := addServicesToTable(tbuild)
	service := coreV1.Service{
		Spec: coreV1.ServiceSpec{
			ClusterIPs:     []string{"10.0.0.10", "", "None"},
			ExternalIPs:    []string{"1.2.3.4"},
			LoadBalancerIP: "5.6.7.8",
		},
	}

	assert.NoError(t, visitor(&resource.Info{Object: &service}, nil))
}

func Test_addPodsToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
			kplug.Column{Name: "Address", Description: "The IP address"},
		)
	visitor := addPodsToTable(tbuild)
	pod := coreV1.Pod{
		Status: coreV1.PodStatus{
			PodIPs: []coreV1.PodIP{{IP: "10.0.0.10"}},
		},
	}

	assert.NoError(t, visitor(&resource.Info{Object: &pod}, nil))
}

func Test_addNodesToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.ResourceKindColumn,
			kplug.Column{Name: "Type", Description: "The type of address"},
			kplug.Column{Name: "Address", Description: "The IP address"},
		)
	visitor := addNodesToTable(tbuild)
	node := coreV1.Node{
		Status: coreV1.NodeStatus{
			Addresses: []coreV1.NodeAddress{
				{Type: coreV1.NodeInternalIP, Address: "10.0.0.10"},
				{Type: coreV1.NodeInternalDNS, Address: "example.local"},
				{Type: coreV1.NodeExternalIP, Address: "1.2.3.4"},
				{Type: coreV1.NodeExternalDNS, Address: "example.com"},
			},
		},
	}

	assert.NoError(t, visitor(&resource.Info{Object: &node}, nil))
}
