package kubectl_trigger_cronjob

import (
	"fmt"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	batchV1 "k8s.io/api/batch/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCommand(apiFactory kube.ApiFactory) *cobra.Command {
	return kube.
		NewTabularRunner(apiFactory, triggerCronJob).
		ToCobraCommand(
			"kubectl-trigger_cronjob",
			"Launches a new job from an existing cronjob",
		)
}

func triggerCronJob(a kube.HandlerArgs) (metaV1.Table, error) {
	cronJobName := a.Args[0]
	job, err := createJobFromCronJob(a, cronJobName)
	if err != nil {
		return metaV1.Table{}, err
	}
	return jobToTable(job), nil
}

func createJobFromCronJob(a kube.HandlerArgs, cronJobName string) (*batchV1.Job, error) {
	client := a.KubeApi.BatchV1()
	cronJob, err := client.
		CronJobs(a.Namespace).
		Get(a.Cmd.Context(), cronJobName, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjob: %w", err)
	}

	job, err := client.
		Jobs(a.Namespace).
		Create(
			a.Cmd.Context(),
			&batchV1.Job{
				ObjectMeta: metaV1.ObjectMeta{GenerateName: cronJob.Name + "-"},
				Spec:       cronJob.Spec.JobTemplate.Spec,
			},
			metaV1.CreateOptions{},
		)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return job, nil
}

func jobToTable(job *batchV1.Job) metaV1.Table {
	return metaV1.Table{
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Kind", Type: "string"},
			{Name: "Name", Type: "string", Format: "name"},
		},
		Rows: []metaV1.TableRow{
			{
				Cells: []interface{}{
					"Job",
					job.GetName(),
				},
				Object: runtime.RawExtension{Object: job},
			},
		},
	}
}
