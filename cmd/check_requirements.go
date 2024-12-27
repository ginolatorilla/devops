package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/ginolatorilla/devops/pkg/exec"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type requirement struct {
	getVersion func(ctx context.Context, exe exec.Executor) string
	constraint string
}

var _requirements = map[string]requirement{
	"kubectl": {
		getVersion: getKubectlVersion,
		constraint: ">=1.29.0",
	},
	"jq": {
		getVersion: getJqVersion,
		constraint: ">=1.7.0",
	},
}

func getKubectlVersion(ctx context.Context, executor exec.Executor) string {
	exec := executor(ctx, "kubectl", "version", "--client", "--output", "json")
	version := _must(exec.Output())
	zap.S().Debugf(`Command: "%s", Output: "%s"`, exec, version)
	type kubectlVersion struct {
		ClientVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}
	var kv kubectlVersion
	_check(json.Unmarshal(version, &kv))
	return kv.ClientVersion.GitVersion
}

func getJqVersion(ctx context.Context, executor exec.Executor) string {
	exec := executor(ctx, "jq", "--version")
	version := _must(exec.Output())
	zap.S().Debugf("\nCommand: %s\nOutput: %s", exec, version)
	parts := strings.Split(strings.TrimSpace(string(version)), "-")
	return parts[1]
}

func newCheckRequirementsCmd(executor exec.Executor) *cobra.Command {
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

func checkRequirements(cmd *cobra.Command, executor exec.Executor, quiet bool) {
	ctx := cmd.Context()
	for name, req := range _requirements {
		actualVersion := semver.MustParse(req.getVersion(ctx, executor))
		constraint := _must(semver.NewConstraint(req.constraint))
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
