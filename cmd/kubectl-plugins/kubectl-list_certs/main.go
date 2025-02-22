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
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
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
		NewRunner(listCerts, append(runnerOpts, kubectlplugin.WithResourcePrinters())...).
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
	tableBuilder := kubectlplugin.NewTableBuilder().AdditionalColumns(
		kubectlplugin.Column{Name: "Key", Description: "The key of the certificate in the secret"},
		kubectlplugin.Column{Name: "Not_Before", Description: "The date and time when the certificate will be effective"},
		kubectlplugin.Column{Name: "Effective", Description: "How long until the certificate becomes effective"},
		kubectlplugin.Column{Name: "Not_After", Description: "The date and time when the certificate will expire"},
		kubectlplugin.Column{Name: "Expires", Description: "How long until the certificate expires"},
	)
	for _, secret := range secrets {
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
		tableBuilder.AddRow(&secret, map[string]interface{}{
			"Key":        tlsCertKey,
			"Not_Before": cert.Leaf.NotBefore,
			"Effective":  humanize.Time(cert.Leaf.NotBefore),
			"Not_After":  cert.Leaf.NotAfter,
			"Expires":    humanize.Time(cert.Leaf.NotAfter),
		})
	}
	return &tableBuilder.Table
}
