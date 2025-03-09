package js

import (
	"fmt"
	"syscall/js"
)

func CheckArgs(args []js.Value, argTypes ...js.Type) error {
	if len(argTypes) != len(args) {
		return fmt.Errorf("expected %d arguments", len(argTypes))
	}
	for i, arg := range args {
		if arg.Type() != argTypes[i] {
			return fmt.Errorf("expected argument %d (%q) to be of type %s", i+1, arg.String(), argTypes[i].String())
		}
	}
	return nil
}
