// package exec defines an interface for the os/exec package and mocks for testing.
//
// The mock types in this package are based on github.com/stretchr/testify/mock.
package exec

import (
	"context"
	"os/exec"
)

// Exec is an interface for the exec.Cmd type from the os/exec package.
type Exec interface {
	Run() error              // Run starts the command. It should return an exec.ExitError if the underlying command fails.
	Output() ([]byte, error) // Output is similar to Run, but it returns the output of the command.
}

// Executor is the function signature of exec.Command.
type Executor func(ctx context.Context, cmd string, args ...string) Exec

func Command(ctx context.Context, cmd string, args ...string) Exec {
	return exec.Command(cmd, args...)
}

func CommandContext(ctx context.Context, cmd string, args ...string) Exec {
	return exec.CommandContext(ctx, cmd, args...)
}
