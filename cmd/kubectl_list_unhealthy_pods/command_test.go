package kubectl_list_unhealthy_pods

import (
	"context"
	"fmt"
	"testing"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)

	t.Run("OK", func(t *testing.T) {
		client := fake.NewClientset()
		loadResources(t, client, "test")

		cmd := testable(client)
		cmd.SetArgs([]string{"-n", "test"})
		assert.NoError(cmd.Execute())
	})

	t.Run("SkipIfPodsUnreadable", func(t *testing.T) {
		client := fake.NewClientset()
		loadCannedError(t, client, "list", "pods")
		loadResources(t, client, "test")

		cmd := testable(client)
		cmd.SetArgs([]string{"-n", "test"})
		assert.NoError(cmd.Execute())
	})
}

func loadResources(t *testing.T, client kubernetes.Interface, namespace string) {
	t.Helper()

	for name, status := range map[string]coreV1.PodStatus{
		"running": {
			Phase: coreV1.PodRunning,
			ContainerStatuses: []coreV1.ContainerStatus{
				{State: coreV1.ContainerState{Running: &coreV1.ContainerStateRunning{}}},
			},
		},
		"failed": {
			Phase: coreV1.PodFailed,
			ContainerStatuses: []coreV1.ContainerStatus{
				{State: coreV1.ContainerState{
					Terminated: &coreV1.ContainerStateTerminated{
						Reason: "CrashLoopBackOff",
					},
				}},
			},
		},
		"image-pull-backoff": {
			Phase: coreV1.PodPending,
			ContainerStatuses: []coreV1.ContainerStatus{
				{State: coreV1.ContainerState{
					Waiting: &coreV1.ContainerStateWaiting{
						Reason: "ImagePullBackOff",
					},
				}},
			},
		},
		"pending-without-reason": {
			Phase: coreV1.PodPending,
			ContainerStatuses: []coreV1.ContainerStatus{
				{State: coreV1.ContainerState{
					Waiting: nil,
				}},
			},
		},
	} {
		if err := createPod(client, namespace, name, status); err != nil {
			t.Fatal(err)
		}
	}
}

func createPod(client kubernetes.Interface, namespace, name string, podStatus coreV1.PodStatus) error {
	if _, err := client.CoreV1().Pods(namespace).Create(
		context.TODO(),
		&coreV1.Pod{
			ObjectMeta: metaV1.ObjectMeta{Name: name},
			Status:     podStatus,
		},
		metaV1.CreateOptions{},
	); err != nil {
		return err
	}
	return nil
}

func loadCannedError(t *testing.T, client *fake.Clientset, verb, resource string) {
	t.Helper()

	client.PrependReactor(verb, resource, func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("canned error from test")
	})
}

func testable(client *fake.Clientset) *cobra.Command {
	return NewCommand(func(configFlags *kube.ConfigFlags) kubernetes.Interface {
		return client
	})
}
