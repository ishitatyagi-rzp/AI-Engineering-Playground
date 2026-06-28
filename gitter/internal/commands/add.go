package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gitter/internal/cli"
	"gitter/internal/index"
)

// AddCommand implements `gitter add`.
type AddCommand struct{}

func (c *AddCommand) Name() string             { return "add" }
func (c *AddCommand) ShortDescription() string { return "Add file contents to the index" }
func (c *AddCommand) Synopsis() string         { return "add <pathspec>..." }
func (c *AddCommand) LongDescription() string {
	return "Update the index using the current content found in the working tree, to prepare\n" +
		"the content staged for the next commit. File paths matching the given pathspec\n" +
		"are added to the staging area."
}
func (c *AddCommand) Options() []cli.Option { return nil }

// Run stages files specified by pathspec.
func (c *AddCommand) Run(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("add: get working directory: %w", err)
	}
	return c.runInDir(cwd, args)
}

// runInDir is the testable core — operates on dir instead of relying on os.Getwd().
func (c *AddCommand) runInDir(dir string, args []string) error {
	gitterDir := filepath.Join(dir, gitterDirName)
	if _, err := os.Stat(gitterDir); os.IsNotExist(err) {
		return fmt.Errorf("add: not a gitter repository (no .gitter directory found)")
	}

	idx, err := index.Load(gitterDir)
	if err != nil {
		return fmt.Errorf("add: load index: %w", err)
	}

	// No args is a no-op.
	if len(args) == 0 {
		return nil
	}

	for _, arg := range args {
		files, err := collectFiles(dir, arg)
		if err != nil {
			return fmt.Errorf("add: collect files for %q: %w", arg, err)
		}
		for _, absPath := range files {
			if err := stageFile(dir, absPath, idx); err != nil {
				return err
			}
		}
	}

	return index.Save(gitterDir, idx)
}

// collectFiles returns absolute paths for the given pathspec under root.
// "." collects all files recursively. Otherwise filepath.Glob is used;
// directories matched by the glob are walked recursively. Silent no-op on zero matches.
func collectFiles(root, pathspec string) ([]string, error) {
	if pathspec == "." {
		return walkDir(root)
	}

	matches, err := filepath.Glob(filepath.Join(root, pathspec))
	if err != nil {
		// Glob only errors on malformed patterns; treat as no match.
		return nil, nil //nolint:nilerr
	}

	var files []string
	for _, match := range matches {
		info, err := os.Lstat(match)
		if err != nil {
			continue
		}
		if info.IsDir() {
			walked, err := walkDir(match)
			if err != nil {
				return nil, err
			}
			files = append(files, walked...)
		} else if info.Mode().IsRegular() {
			files = append(files, match)
		}
	}
	return files, nil
}

// walkDir recursively collects regular files under dir, skipping .gitter/.
func walkDir(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == gitterDirName {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// stageFile hashes absPath and upserts an entry in idx with a path relative to repoRoot.
func stageFile(repoRoot, absPath string, idx *index.Index) error {
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return fmt.Errorf("add: resolve relative path for %s: %w", absPath, err)
	}
	// Normalise to forward slashes for cross-platform consistency.
	relPath = filepath.ToSlash(relPath)

	hash, err := index.HashFile(absPath)
	if err != nil {
		return fmt.Errorf("add: hash file %s: %w", absPath, err)
	}

	idx.Entries[relPath] = index.Entry{Hash: hash, StagedAs: "new file"}
	return nil
}
