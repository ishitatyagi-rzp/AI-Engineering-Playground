package commands

import (
	"fmt"

	"gitter/internal/cli"
)

// CommitCommand implements `gitter commit`.
type CommitCommand struct{}

func (c *CommitCommand) Name() string             { return "commit" }
func (c *CommitCommand) ShortDescription() string { return "Record changes to the repository" }
func (c *CommitCommand) Synopsis() string         { return "commit -m [-a] <msg>" }
func (c *CommitCommand) LongDescription() string {
	return "Create a new commit containing the current contents of the index and the given\n" +
		"log message describing the changes. The new commit is a direct child of HEAD,\n" +
		"usually the tip of the current branch, and the branch is updated to point to it."
}
func (c *CommitCommand) Options() []cli.Option {
	return []cli.Option{
		{
			Flag:        "a",
			Description: "Tell the command to automatically stage files that have been modified and deleted, but new files you have not told Git about are not affected.",
		},
		{
			Flag:        "m",
			Description: "Use the given <msg> as the commit message. If multiple -m options are given, their values are concatenated as separate paragraphs.",
		},
	}
}

// Run records staged changes as a new commit.
func (c *CommitCommand) Run(args []string) error {
	fmt.Println("not implemented")
	return nil
}
