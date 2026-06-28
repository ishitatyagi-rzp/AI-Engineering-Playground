package commands

import (
	"fmt"
	"strings"

	"gitter/internal/cli"
)

const helpCommandName = "help"

// HelpCommand implements `gitter help` and `gitter help <command>`.
// It holds a reference to the registry so it can enumerate all commands.
type HelpCommand struct {
	registry *cli.Registry
}

// NewHelpCommand creates a HelpCommand backed by the given registry.
func NewHelpCommand(reg *cli.Registry) *HelpCommand {
	return &HelpCommand{registry: reg}
}

func (h *HelpCommand) Name() string             { return helpCommandName }
func (h *HelpCommand) ShortDescription() string { return "Show help for gitter commands" }
func (h *HelpCommand) Synopsis() string         { return "help [<command>]" }
func (h *HelpCommand) LongDescription() string {
	return "With no arguments, lists all available gitter commands with a short description.\n" +
		"With a command name, prints NAME, SYNOPSIS, DESCRIPTION, and OPTIONS for that command."
}
func (h *HelpCommand) Options() []cli.Option { return nil }

// Run prints the command list when called with no args, or detailed help for a named command.
func (h *HelpCommand) Run(args []string) error {
	if len(args) == 0 {
		h.printCommandList()
		return nil
	}
	return h.printCommandDetail(args[0])
}

// printCommandList prints all registered commands (excluding help itself) in a
// two-column format: name (left-padded, right-aligned to a common width) + description.
func (h *HelpCommand) printCommandList() {
	cmds := visibleCommands(h.registry)

	nameWidth := maxNameLen(cmds) + 3 // +3 for minimum gap between name and description

	fmt.Println("These are common Gitter commands:")
	for _, cmd := range cmds {
		padding := strings.Repeat(" ", nameWidth-len(cmd.Name()))
		fmt.Printf(" %s%s%s\n", cmd.Name(), padding, cmd.ShortDescription())
	}
}

// printCommandDetail prints the man-page-style help for a single command.
func (h *HelpCommand) printCommandDetail(name string) error {
	cmd, ok := h.registry.Get(name)
	if !ok {
		return fmt.Errorf("'%s' is not a gitter command", name)
	}

	fmt.Printf("NAME:\n%s - %s\n\n", cmd.Name(), cmd.ShortDescription())
	fmt.Printf("SYNOPSIS:\ngitter %s\n\n", cmd.Synopsis())
	fmt.Printf("DESCRIPTION:\n%s\n", cmd.LongDescription())

	if opts := cmd.Options(); len(opts) > 0 {
		fmt.Println("\nOPTIONS:")
		for _, opt := range opts {
			fmt.Printf("-%s: %s\n", opt.Flag, opt.Description)
		}
	}

	return nil
}

// visibleCommands returns all commands except the help command itself.
func visibleCommands(reg *cli.Registry) []cli.Command {
	all := reg.All()
	visible := make([]cli.Command, 0, len(all))
	for _, cmd := range all {
		if cmd.Name() != helpCommandName {
			visible = append(visible, cmd)
		}
	}
	return visible
}

// maxNameLen returns the length of the longest command name in the list.
func maxNameLen(cmds []cli.Command) int {
	max := 0
	for _, cmd := range cmds {
		if n := len(cmd.Name()); n > max {
			max = n
		}
	}
	return max
}
