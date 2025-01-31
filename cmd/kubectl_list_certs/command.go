package kubectl_list_certs

import (
	"crypto/tls"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewCommand() *cobra.Command {
	configFlags := genericclioptions.NewConfigFlags(true)
	var allNamespaces bool
	var noHeaders bool

	command := &cobra.Command{
		Use:   "kubectl_list_certs",
		Short: "List certificates in the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			kubeConfig := configFlags.ToRawKubeConfigLoader()
			config, err := kubeConfig.ClientConfig()
			if err != nil {
				panic(err)
			}

			var namespace string
			if !allNamespaces {
				namespace = *configFlags.Namespace
			}
			if namespace == "" {
				namespace, _, err = kubeConfig.Namespace()
				if err != nil {
					panic(err)
				}
			}

			printer := printers.NewTablePrinter(printers.PrintOptions{
				WithNamespace: allNamespaces,
				NoHeaders:     noHeaders,
			})

			client := coreV1.NewForConfigOrDie(config)
			secrets, err := client.
				Secrets(namespace).
				List(cmd.Context(), metaV1.ListOptions{
					FieldSelector: "type=kubernetes.io/tls",
				})
			if err != nil {
				panic(err)
			}

			table := metaV1.Table{
				ColumnDefinitions: []metaV1.TableColumnDefinition{
					{Name: "Name", Type: "string"},
					{Name: "Key", Type: "string"},
					{Name: "Start", Type: "string"},
					{Name: "Effective", Type: "string"},
					{Name: "Until", Type: "string"},
					{Name: "Expires", Type: "string"},
				},
				Rows: make([]metaV1.TableRow, len(secrets.Items)),
			}

			for i, secret := range secrets.Items {
				tlsCertKey := "tls.crt"
				cert, err := tls.X509KeyPair(secret.Data[tlsCertKey], secret.Data["tls.key"])
				if err != nil {
					panic(err)
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

			printer.PrintObj(&table, cmd.OutOrStdout())
		},
	}

	configFlags.AddFlags(command.Flags())
	command.Flags().BoolVarP(
		&allNamespaces,
		"all-namespaces",
		"A",
		false,
		"If present, list the requested object(s) across all namespaces. "+
			"Namespace in current context is ignored even if specified with --namespace.",
	)
	command.Flags().BoolVar(
		&noHeaders,
		"no-headers",
		false,
		"Don't print headers (default print headers).",
	)
	return command
}
