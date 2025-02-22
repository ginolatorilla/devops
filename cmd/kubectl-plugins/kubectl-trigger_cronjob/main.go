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

	kplug "github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	batchV1 "k8s.io/api/batch/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	newCommand(kplug.WithDefaultKubeApi()).Execute()
}

func newCommand(runnerOpts ...kplug.RunnerOpts) *cobra.Command {
	runnerOpts = append(runnerOpts, kplug.WithResourcePrinters())
	return kplug.
		NewRunner(triggerCronJob, runnerOpts...).
		ToCobraCommand(
			"kubectl-trigger_cronjob",
			"Launches a new job from an existing cronjob",
		)
}

func triggerCronJob(a kplug.HandlerArgs) (runtime.Object, error) {
	cronJobName := a.Args[0]
	client := a.KubeApi.BatchV1()
	cronJob, err := client.
		CronJobs(*a.ConfigFlags.Namespace).
		Get(a.Cmd.Context(), cronJobName, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjob: %w", err)
	}
	job, err := client.
		Jobs(*a.ConfigFlags.Namespace).
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
