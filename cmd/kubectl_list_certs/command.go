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
package kubectl_list_certs

import (
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/dustin/go-humanize"
	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCommand(runnerOpts ...kubectlplugin.RunnerOpts) *cobra.Command {
	return kubectlplugin.
		NewRunnerV2(listCerts, append(runnerOpts, kubectlplugin.WithResourcePrinters())...).
		ToCobraCommand(
			"kubectl-list_certs",
			"Lists all certificates in the cluster and shows when they will be effective and when they will expire",
		)
}

func listCerts(a kubectlplugin.HandlerArgs) (runtime.Object, error) {
	client := a.KubeApi.CoreV1()
	secrets, err := client.
		Secrets(a.Namespace).
		List(a.Cmd.Context(), metaV1.ListOptions{
			FieldSelector: "type=kubernetes.io/tls",
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	return secretsToTable(secrets.Items), nil
}

func secretsToTable(secrets []coreV1.Secret) runtime.Object {
	table := metaV1.Table{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "meta.k8s.io/v1",
			Kind:       "Table",
		},
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Key", Type: "string"},
			{Name: "NotBefore", Type: "string", Format: "date-time"},
			{Name: "Effective", Type: "string"},
			{Name: "NotAfter", Type: "string", Format: "date-time"},
			{Name: "Expires", Type: "string"},
		},
		Rows: make([]metaV1.TableRow, len(secrets)),
	}
	for i, secret := range secrets {
		tlsCertKey := "tls.crt"
		cert, err := tls.X509KeyPair(secret.Data[tlsCertKey], secret.Data["tls.key"])
		if err != nil {
			slog.Warn(
				"failed to parse certificate",
				"namepsace", secret.GetNamespace(),
				"secretName", secret.GetName(),
				"key", tlsCertKey,
				"error", err)
			continue
		}
		table.Rows[i] = metaV1.TableRow{
			Cells: []interface{}{
				secret.Name,
				tlsCertKey,
				cert.Leaf.NotBefore,
				humanize.Time(cert.Leaf.NotBefore),
				cert.Leaf.NotAfter,
				humanize.Time(cert.Leaf.NotAfter),
			},
			Object: runtime.RawExtension{Object: &secret},
		}
	}
	return &table
}
