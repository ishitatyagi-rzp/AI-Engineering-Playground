package commands

import (
	"fmt"

	"gitter/internal/cli"
)

// CheckoutCommand implements `gitter checkout`.
type CheckoutCommand struct{}

func (c *CheckoutCommand) Name() string { return "checkout" }
func (c *CheckoutCommand) ShortDescription() string {
	return "Switch branches or restore working tree files"
}
func (c *CheckoutCommand) Synopsis() string { return "checkout [-b] <branch>" }
func (c *CheckoutCommand) LongDescription() string {
	return "Switch to another branch and update the working tree to match. If -b is given,\n" +
		"create the branch before switching. The working tree files are updated to match\n" +
		"the version in the target branch."
}
func (c *CheckoutCommand) Options() []cli.Option {
	return []cli.Option{
		{
			Flag:        "b",
			Description: "Create a new branch and switch to it immediately.",
		},
	}
}

// Run switches the current branch or restores working tree files.
func (c *CheckoutCommand) Run(args []string) error {
	fmt.Println("not implemented")
	return nil
}
