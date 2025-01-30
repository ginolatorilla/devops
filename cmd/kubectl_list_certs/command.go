package kubectl_list_certs

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl_list_certs",
		Short: "List certificates in the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			// Your code here
		},
	}

	return cmd
}
