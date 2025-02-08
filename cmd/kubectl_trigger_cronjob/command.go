package kubectl_trigger_cronjob

import (
	"fmt"
	"log/slog"

	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	batchV1 "k8s.io/api/batch/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
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
	client := a.KubeApi.BatchV1()

	cronJob, err := client.CronJobs(a.Namespace).Get(a.Cmd.Context(), cronJobName, metaV1.GetOptions{})
	if err != nil {
		return metaV1.Table{}, fmt.Errorf("failed to get cronjob: %w", err)
	}

	job, err := client.Jobs(a.Namespace).Create(a.Cmd.Context(), &batchV1.Job{
		ObjectMeta: metaV1.ObjectMeta{
			GenerateName: cronJob.Name + "-",
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}, metaV1.CreateOptions{})
	if err != nil {
		return metaV1.Table{}, fmt.Errorf("failed to create job: %w", err)
	}

	gvk, err := apiutil.GVKForObject(job, scheme.Scheme)
	if err != nil {
		slog.Warn("cannot get group/version/kind of resource", "object", job)
	}

	return metaV1.Table{
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Kind", Type: "string"},
			{Name: "Name", Type: "string", Format: "name"},
		},
		Rows: []metaV1.TableRow{
			{
				Cells: []interface{}{
					gvk.Kind,
					job.GetName(),
				},
			},
		},
	}, nil
}
