package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/semver"
	u "github.com/ginolatorilla/devops/pkg/utils"
	"github.com/spf13/cobra"
)

type requirement struct {
	args       []string
	getVersion func(ctx context.Context, e u.Executor) string
	constraint string
}

var _requirements = map[string]requirement{
	"kubectl": {
		args: []string{
			"version",
			"--client",
			"--output", "json",
		},
		getVersion: getKubectlVersion,
		constraint: ">=1.29.0",
	},
}

func getKubectlVersion(ctx context.Context, executor u.Executor) string {
	exec := executor(ctx, "kubectl", "version", "--client", "--output", "json")
	version := u.Must(exec.Output())
	type kubectlVersion struct {
		ClientVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}
	var kv kubectlVersion
	u.Check(json.Unmarshal(version, &kv))
	return kv.ClientVersion.GitVersion
}

func newCheckRequirementsCmd(executor u.Executor) *cobra.Command {
	var quiet bool
	cmd := &cobra.Command{
		Use:   "check-requirements",
		Short: "Check the requirements for the application",
		Run: func(cmd *cobra.Command, args []string) {
			checkRequirements(cmd, executor, quiet)
		},
	}
	cmd.Flags().BoolVarP(&quiet, "quiet" /* name */, "q" /* short */, false, /* default */
		"Print only the errors" /* usage */)
	return cmd
}

func checkRequirements(cmd *cobra.Command, executor u.Executor, quiet bool) {
	ctx := cmd.Context()
	for name, req := range _requirements {
		actualVersion := semver.MustParse(req.getVersion(ctx, executor))
		constraint := u.Must(semver.NewConstraint(req.constraint))
		if !constraint.Check(actualVersion) {
			panic(fmt.Errorf("%s version %s does not meet requirement %s",
				name, actualVersion, req.constraint,
			))
		}
		if !quiet {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.Printf("✅ %s (%s)\n", name, actualVersion)
		}
	}
}
