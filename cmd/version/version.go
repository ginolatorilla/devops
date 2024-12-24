package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// NewCommand creates the version command.
func NewCommand(version, commitHash string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of the application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version:   ", version)
			fmt.Println("Commit:    ", commitHash)
			fmt.Println("Go Version:", runtime.Version())
			fmt.Println("Compiler:  ", runtime.Compiler)
			fmt.Println("Platform:  ", runtime.GOOS, runtime.GOARCH)
		},
	}
	return cmd
}
