package kubectl_list_certs

import (
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/dustin/go-humanize"
	"github.com/ginolatorilla/devops/pkg/kube"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func NewCommand(apiFactory kube.ApiFactory) *cobra.Command {
	configFlags := kube.NewConfigFlags()
	runner := kube.NewRunner(configFlags, apiFactory, listCerts)

	command := &cobra.Command{
		Use:   "kubectl_list_certs",
		Short: "List certificates in the cluster",
		RunE:  runner.ToRunE(),
	}

	configFlags.AddFlags(command.Flags())
	return command
}

func listCerts(kubeApi kubernetes.Interface, namespace string, cmd *cobra.Command, configFlags *kube.ConfigFlags) (metaV1.Table, error) {
	client := kubeApi.CoreV1()
	secrets, err := client.
		Secrets(namespace).
		List(cmd.Context(), metaV1.ListOptions{
			FieldSelector: "type=kubernetes.io/tls",
		})
	if err != nil {
		return metaV1.Table{}, fmt.Errorf("failed to list secrets: %w", err)
	}

	table := metaV1.Table{
		ColumnDefinitions: []metaV1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Key", Type: "string"},
			{Name: "NotBefore", Type: "string", Format: "date-time"},
			{Name: "Effective", Type: "string"},
			{Name: "NotAfter", Type: "string", Format: "date-time"},
			{Name: "Expires", Type: "string"},
		},
		Rows: make([]metaV1.TableRow, len(secrets.Items)),
	}

	for i, secret := range secrets.Items {
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
	return table, nil
}
