package cmd

import (
	"github.com/ginolatorilla/devops/pkg/exec"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var AppName = "devops" // Name of the application
var Version = ""       // Version of the application
var CommitHash = ""    // Commit hash of the application

func Execute() {
	cmd := newRootCmd(AppName)
	cmd.AddCommand(
		newVersionCmd(Version, CommitHash),
		newCheckRequirementsCmd(exec.CommandContext),
		newTlsCertCountdownCmd(),
		newTemplateCmd(),
		newCorsTestCmd(),
	)
	_check(cmd.Execute())
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
