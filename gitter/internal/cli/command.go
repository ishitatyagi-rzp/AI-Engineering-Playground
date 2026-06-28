package cli

// Option represents a single command-line flag and its description.
type Option struct {
	Flag        string
	Description string
}

// Command is the interface every gitter subcommand must satisfy.
// Adding a new command: implement this interface, then call Registry.Register in main.go.
type Command interface {
	// Name returns the subcommand name used on the CLI (e.g. "init").
	Name() string

	// ShortDescription returns the one-line summary shown in `gitter help` list.
	ShortDescription() string

	// Synopsis returns the usage line shown under SYNOPSIS in `gitter help <cmd>`.
	Synopsis() string

	// LongDescription returns the body shown under DESCRIPTION in `gitter help <cmd>`.
	LongDescription() string

	// Options returns the flags shown under OPTIONS in `gitter help <cmd>`.
	// Return nil or empty slice if the command has no options.
	Options() []Option

	// Run executes the command with the remaining CLI arguments after the command name.
	Run(args []string) error
}
