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
	"log/slog"

	coreV1 "k8s.io/api/core/v1"

	"github.com/dustin/go-humanize"
	kplug "github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

func main() {
	newCommand(kplug.WithDefaultKubeApi()).Execute()
}

func newCommand(runnerOpts ...kplug.RunnerOpts) *cobra.Command {
	runnerOpts = append(runnerOpts, kplug.WithTablePrinter(), kplug.WithAllNamespaces())
	return kplug.NewRunner(listCerts, runnerOpts...).
		ToCobraCommand(
			"kubectl-list_certs",
			"Lists all certificates in the cluster and shows when they will be effective and when they will expire",
		)
}

func listCerts(a kplug.HandlerArgs) (runtime.Object, error) {
	tableBuilder := kplug.NewTableBuilder().AdditionalColumns(
		kplug.Column{Name: "Key", Description: "The key of the certificate in the secret"},
		kplug.Column{Name: "Not_Before", Description: "The date and time when the certificate will be effective"},
		kplug.Column{Name: "Effective", Description: "How long until the certificate becomes effective"},
		kplug.Column{Name: "Not_After", Description: "The date and time when the certificate will expire"},
		kplug.Column{Name: "Expires", Description: "How long until the certificate expires"},
	)
	a.ConfigFlags.WithAll(true).WithFieldSelector("type=kubernetes.io/tls")
	a.ToResourceFinder("secrets").Do().Visit(secretsToTable(tableBuilder))
	return &tableBuilder.Table, nil
}

func secretsToTable(tableBuilder *kplug.TableBuilder) resource.VisitorFunc {
	return func(info *resource.Info, _ error) error {
		secret := kplug.As[*coreV1.Secret](info.Object)
		tlsCertKey := "tls.crt"
		cert, err := tls.X509KeyPair(secret.Data[tlsCertKey], secret.Data["tls.key"])
		if err != nil {
			slog.Warn(
				"failed to parse certificate",
				"namepsace", secret.GetNamespace(),
				"secretName", secret.GetName(),
				"key", tlsCertKey,
				"error", err)
			return nil
		}
		tableBuilder.AddRow(secret, map[string]interface{}{
			"Key":        tlsCertKey,
			"Not_Before": cert.Leaf.NotBefore,
			"Effective":  humanize.Time(cert.Leaf.NotBefore),
			"Not_After":  cert.Leaf.NotAfter,
			"Expires":    humanize.Time(cert.Leaf.NotAfter),
		})
		return nil
	}
}
