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
		cmd := newCommand(kplug.WithDefaultKubeApi())
		cmd.SetArgs([]string{"-A"})
		assert.NoError(cmd.Execute())
	})
}

func Test_addUnhealthyPodsToTable(t *testing.T) {
	tbuild := kplug.NewTableBuilder().
		AdditionalColumns(
			kplug.Column{Name: "Status", Description: "The status of the pod"},
			kplug.Column{Name: "Reason", Description: "The reason for the pod's status"},
		)
	visitor := addUnhealthyPodsToTable(tbuild)
	assert := assert.New(t)

	for testCase, podStatus := range map[string]coreV1.PodStatus{
		"Terminating": {
			Phase: coreV1.PodFailed,
			ContainerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{
						Terminated: &coreV1.ContainerStateTerminated{},
					},
				},
			},
		},
		"ImagePullBackoff": {
			Phase: coreV1.PodPending,
			ContainerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{
							Reason: "ImagePullBackOff",
						},
					},
				},
			},
		},
		"PodInitializing": {
			Phase: coreV1.PodPending,
			ContainerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{
							Reason: "PodInitializing",
						},
					},
				},
			},
		},
		"PendingButNoWaitingStatus": {
			Phase: coreV1.PodPending,
			ContainerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{},
				},
			},
		},
		"Unknown": {
			Phase: coreV1.PodUnknown,
		},
	} {
		t.Run(testCase, func(t *testing.T) {
			assert.NoError(visitor(&resource.Info{Object: &coreV1.Pod{Status: podStatus}}, nil))
		})
	}
}
