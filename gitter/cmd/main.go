package main

import (
	"os"

	"gitter/internal/cli"
	"gitter/internal/commands"
)

func main() {
	reg := cli.NewRegistry()

	// help must be registered first so its listing reflects insertion order.
	helpCmd := commands.NewHelpCommand(reg)
	reg.Register(helpCmd)

	// Register commands in the order they should appear in `gitter help`.
	reg.Register(&commands.InitCommand{})
	reg.Register(&commands.AddCommand{})
	reg.Register(&commands.StatusCommand{})
	reg.Register(&commands.CommitCommand{})
	reg.Register(&commands.CheckoutCommand{})

	cli.Dispatch(reg, os.Args[1:])
}
