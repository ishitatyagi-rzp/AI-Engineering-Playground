package cli

import (
	"fmt"
	"os"
)

// Dispatch routes the parsed CLI arguments to the matching command in the registry.
// With no arguments it falls back to "help". Unknown commands print an error to
// stderr and exit with code 1 — matching Git's behaviour.
func Dispatch(reg *Registry, args []string) {
	if len(args) == 0 {
		args = []string{"help"}
	}

	cmdName := args[0]
	cmd, ok := reg.Get(cmdName)
	if !ok {
		fmt.Fprintf(os.Stderr, "gitter: '%s' is not a gitter command. See 'gitter help'.\n", cmdName)
		os.Exit(1)
	}

	if err := cmd.Run(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
