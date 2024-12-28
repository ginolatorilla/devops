package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	goexec "os/exec"
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
	"bash": {
		getVersion: getBashVersion,
		constraint: ">=3.0.0",
	},
	"kubectl": {
		getVersion: getKubectlVersion,
		constraint: ">=1.29.0",
	},
	"jq": {
		getVersion: getJqVersion,
		constraint: ">=1.7.0",
	},
	"openssl": {
		getVersion: getOpensslVersion,
		constraint: ">=3.0.0",
	},
	"column": {},
}

func getBashVersion(ctx context.Context, executor exec.Executor) string {
	exec := executor(ctx, "bash", "--version")
	version := _must(exec.Output())
	zap.S().Debugf("\nCommand: %s\nOutput: %s", exec, version)
	parts := strings.Split(strings.TrimSpace(string(version)), " ")
	parts = strings.Split(parts[3], "(")
	return parts[0]
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

func getOpensslVersion(ctx context.Context, executor exec.Executor) string {
	exec := executor(ctx, "openssl", "version")
	version := _must(exec.Output())
	zap.S().Debugf("\nCommand: %s\nOutput: %s", exec, version)
	parts := strings.Split(strings.TrimSpace(string(version)), " ")
	return parts[1]
}

func newCheckRequirementsCmd(executor exec.Executor) *cobra.Command {
	var quiet bool
	var list bool
	cmd := &cobra.Command{
		Use:   "check-requirements",
		Short: "Check the requirements for the application",
		Run: func(cmd *cobra.Command, args []string) {
			if list {
				for name, r := range _requirements {
					if r.getVersion != nil {
						cmd.Printf("%s (%s)\n", name, r.constraint)
						continue
					}
					cmd.Println(name)
					continue
				}
				return
			}
			checkRequirements(cmd, executor, quiet)
		},
	}
	cmd.Flags().BoolVarP(&quiet, "quiet" /* name */, "q" /* short */, false, /* default */
		"Print only the errors" /* usage */)
	cmd.Flags().BoolVar(&list, "list" /* name */, false /* default */, "List requirements" /* usage */)
	return cmd
}

func checkRequirements(cmd *cobra.Command, executor exec.Executor, quiet bool) {
	ctx := cmd.Context()
	for name, req := range _requirements {
		if req.getVersion == nil {
			_, err := goexec.LookPath(name)
			if err != nil {
				panic(fmt.Errorf("%s is not installed", name))
			}
			if !quiet {
				cmd.Printf("✅ %s\n", name)
			}
			continue
		}

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
