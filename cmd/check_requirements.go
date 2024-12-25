package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/semver"
	u "github.com/ginolatorilla/devops/pkg/utils"
	"github.com/spf13/cobra"
)

var requirements = map[string]requirement{
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

type requirement struct {
	args       []string
	getVersion func(ctx context.Context, e u.Executor) string
	constraint string
}

func newCheckRequirementsCmd(executor u.Executor) *cobra.Command {
	return &cobra.Command{
		Use:   "check-requirements",
		Short: "Check the requirements for the application",
		Run: func(cmd *cobra.Command, args []string) {
			for name, req := range requirements {
				actualVersion := semver.MustParse(req.getVersion(cmd.Context(), executor))
				constraint := u.Must(semver.NewConstraint(req.constraint))
				if !constraint.Check(actualVersion) {
					panic(fmt.Errorf("%s version %s does not meet requirement %s",
						name, actualVersion, req.constraint,
					))
				}
			}
		},
	}
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
