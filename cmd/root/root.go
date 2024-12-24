package root

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCommand(appName string) *cobra.Command {
	var verbosity int

	cobra.OnInitialize(
		func() { setUpLogger(verbosity) },
	)

	cmd := &cobra.Command{
		Use:   appName,
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
