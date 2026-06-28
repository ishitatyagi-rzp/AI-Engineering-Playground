package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"gitter/internal/cli"
)

const (
	gitterDirName = ".gitter"
	headFileName  = "HEAD"
	defaultBranch = "main"
	// HEAD content follows Git's ref format, pointing at the default branch.
	headFileContent = "ref: refs/heads/main\n"
)

// InitCommand implements `gitter init`.
type InitCommand struct{}

func (c *InitCommand) Name() string             { return "init" }
func (c *InitCommand) ShortDescription() string { return "Create an empty Gitter repository" }
func (c *InitCommand) Synopsis() string         { return "init" }
func (c *InitCommand) LongDescription() string {
	return "Create an empty Gitter repository or reinitialize an existing one.\n" +
		"A .gitter/ directory is created in the current working directory containing\n" +
		"HEAD, objects/, and refs/heads/ — the minimal skeleton for a Gitter repo."
}
func (c *InitCommand) Options() []cli.Option { return nil }

// Run initialises a .gitter/ repo in the current working directory.
func (c *InitCommand) Run(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("init: get working directory: %w", err)
	}
	return c.runInDir(cwd)
}

// runInDir is the testable core: it operates on the given directory instead of
// relying on os.Getwd(). Run delegates here after resolving the working directory.
func (c *InitCommand) runInDir(dir string) error {
	gitterPath := filepath.Join(dir, gitterDirName)

	if _, err := os.Stat(gitterPath); err == nil {
		// Repository already exists — inform the user and return cleanly.
		fmt.Printf("Gitter repository is already initialised in %s/\n", gitterPath)
		return nil
	}

	if err := createRepoSkeleton(gitterPath); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	fmt.Printf("Initialized empty Gitter repository in %s/\n", gitterPath)
	return nil
}

// createRepoSkeleton creates the .gitter directory tree and the HEAD file.
func createRepoSkeleton(gitterPath string) error {
	// Directories required for the minimal repo skeleton.
	dirs := []string{
		filepath.Join(gitterPath, "objects"),
		filepath.Join(gitterPath, "refs", "heads"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	headPath := filepath.Join(gitterPath, headFileName)
	if err := os.WriteFile(headPath, []byte(headFileContent), 0o644); err != nil {
		return fmt.Errorf("write HEAD: %w", err)
	}

	return nil
}
