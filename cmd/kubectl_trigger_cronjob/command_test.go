package kubectl_trigger_cronjob

import (
	"context"
	"testing"

	kubetesting "github.com/ginolatorilla/devops/pkg/kube/testing"
	"github.com/stretchr/testify/assert"
	batchV1 "k8s.io/api/batch/v1"
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

		cmd := kubetesting.Testable(client, NewCommand)
		cmd.SetArgs([]string{"-n", "test", "test-cronjob"})
		assert.NoError(cmd.Execute())
	})

	t.Run("ErrorIfCronjobNotFound", func(t *testing.T) {
		client := fake.NewClientset()
		loadResources(t, client, "test")

		cmd := kubetesting.Testable(client, NewCommand)
		cmd.SetArgs([]string{"-n", "test", "missing-cronjob"})
		assert.Error(cmd.Execute())
	})

	t.Run("ErrorIfFailedToCreateJob", func(t *testing.T) {
		client := fake.NewClientset()
		kubetesting.LoadCannedError(t, client, "create", "jobs")
		loadResources(t, client, "test")

		cmd := kubetesting.Testable(client, NewCommand)
		cmd.SetArgs([]string{"-n", "test", "test-cronjob"})
		assert.Error(cmd.Execute())
	})
}

func loadResources(t *testing.T, client kubernetes.Interface, namespace string) {
	t.Helper()

	if _, err := client.BatchV1().CronJobs(namespace).Create(
		context.TODO(),
		&batchV1.CronJob{
			ObjectMeta: metaV1.ObjectMeta{Name: "test-cronjob"},
			Spec: batchV1.CronJobSpec{
				JobTemplate: batchV1.JobTemplateSpec{
					Spec: batchV1.JobSpec{
						Template: coreV1.PodTemplateSpec{
							Spec: coreV1.PodSpec{
								Containers: []coreV1.Container{
									{Name: "test", Image: "test"},
								},
							},
						},
					},
				},
			},
		},
		metaV1.CreateOptions{},
	); err != nil {
		t.Fatal(err)
	}
}
