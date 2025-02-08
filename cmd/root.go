// package cmd implements the devops CLI.
//
// # Copyright © 2025 Gino Latorilla
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
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ginolatorilla/devops/cmd/kubectl_list_addresses"
	"github.com/ginolatorilla/devops/cmd/kubectl_list_certs"
	"github.com/ginolatorilla/devops/cmd/kubectl_list_finalizers"
	"github.com/ginolatorilla/devops/cmd/kubectl_list_unhealthy_pods"
	"github.com/ginolatorilla/devops/cmd/kubectl_lookup_address"
	"github.com/ginolatorilla/devops/cmd/kubectl_trigger_cronjob"
	"github.com/ginolatorilla/devops/pkg/kubectlplugin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var AppName = "devops" // Name of the application
var Version = ""       // Version of the application
var CommitHash = ""    // Commit hash of the application

func Execute() {
	plugin := filepath.Base(os.Args[0])
	if pluginFromEnv := os.Getenv("DEVOPS_BINARY"); pluginFromEnv != "" {
		plugin = pluginFromEnv
	}

	var command *cobra.Command
	switch plugin {
	case "kubectl-list_certs":
		command = kubectl_list_certs.NewCommand(kubectlplugin.DefaultApiFactory)
	case "kubectl-lookup_address":
		command = kubectl_lookup_address.NewCommand(kubectlplugin.DefaultApiFactory)
	case "kubectl-list_unhealthy_pods":
		command = kubectl_list_unhealthy_pods.NewCommand(kubectlplugin.DefaultApiFactory)
	case "kubectl-trigger_cronjob":
		command = kubectl_trigger_cronjob.NewCommand(kubectlplugin.DefaultApiFactory)
	case "kubectl-list_addresses":
		command = kubectl_list_addresses.NewCommand(kubectlplugin.DefaultApiFactory)
	case "kubectl-list_finalizers":
		command = kubectl_list_finalizers.NewCommand(kubectlplugin.DefaultDynamicApiFactory)
	default:
		command = newRootCmd(AppName)
		command.AddCommand(
			newVersionCmd(Version, CommitHash),
			newTemplateCmd(),
			newCorsTestCmd(),
		)
	}

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// newRootCmd creates the root command.
func newRootCmd(appName string) *cobra.Command {
	var verbosity int
	cobra.OnInitialize(
		func() { setUpLogger(verbosity) },
	)
	cmd := &cobra.Command{
		Use:   appName,
		Short: "Helper tool for DevOps",
	}
	cmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose",
		"v",
		"Verbosity level. Use -v for verbose, -vv for more verbose, etc.",
	)
	return cmd
}

// setUpLogger sets up the logger based on the verbosity level.
//
// This function mimics the default logging level of Python's logger (starts at WARNING).
func setUpLogger(verbosity int) {
	lvl := zap.WarnLevel
	trace := false

	switch verbosity {
	case 0:
		lvl = zap.WarnLevel
		trace = false
	case 1:
		lvl = zap.InfoLevel
		trace = false
	case 2:
		lvl = zap.DebugLevel
		trace = false
	default:
		lvl = zap.DebugLevel
		trace = true
	}
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(lvl)
	config.DisableStacktrace = !trace
	zap.ReplaceGlobals(zap.Must(config.Build()))
}
