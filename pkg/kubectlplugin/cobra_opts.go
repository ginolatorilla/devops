package kubectlplugin

import (
	"github.com/spf13/cobra"
)

type CobraOpts func(*cobra.Command)

func WithArgs(args cobra.PositionalArgs, validArgs []string) CobraOpts {
	return func(c *cobra.Command) {
		c.Args = args
		c.ValidArgs = validArgs
	}
}

func WithVersion(version string) CobraOpts {
	return func(c *cobra.Command) {
		c.Version = version
	}
}
