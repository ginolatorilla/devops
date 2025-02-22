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
package main

import (
	"fmt"
	"os"

	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	batchV1 "k8s.io/api/batch/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	if err := newCommand(kubectlplugin.WithDefaultKubeApi()).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunner(triggerCronJob, append(runnerOpts, kubectlplugin.WithResourcePrinters())...).
		ToCobraCommand(
			"kubectl-trigger_cronjob",
			"Launches a new job from an existing cronjob",
		)
}

func triggerCronJob(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	cronJobName := a.Args[0]
	return createJobFromCronJob(a, cronJobName)
}

func createJobFromCronJob(a kubectlplugin.HandlerArgs, cronJobName string) (*batchV1.Job, error) {
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
