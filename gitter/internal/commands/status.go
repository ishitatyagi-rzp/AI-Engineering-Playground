package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gitter/internal/cli"
	"gitter/internal/index"
)

// StatusCommand implements `gitter status`.
type StatusCommand struct{}

func (c *StatusCommand) Name() string             { return "status" }
func (c *StatusCommand) ShortDescription() string { return "Show the working tree status" }
func (c *StatusCommand) Synopsis() string         { return "status" }
func (c *StatusCommand) LongDescription() string {
	return "Display paths that have differences between the index and HEAD, paths that\n" +
		"have differences between the working tree and the index, and paths in the\n" +
		"working tree that are not tracked by Gitter."
}
func (c *StatusCommand) Options() []cli.Option { return nil }

// Run shows the current working tree status.
func (c *StatusCommand) Run(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("status: get working directory: %w", err)
	}
	return c.runInDir(cwd)
}

// runInDir is the testable core — operates on dir instead of relying on os.Getwd().
func (c *StatusCommand) runInDir(dir string) error {
	gitterDir := filepath.Join(dir, gitterDirName)
	if _, err := os.Stat(gitterDir); os.IsNotExist(err) {
		return fmt.Errorf("status: not a gitter repository (no .gitter directory found)")
	}

	idx, err := index.Load(gitterDir)
	if err != nil {
		return fmt.Errorf("status: load index: %w", err)
	}

	untracked, staged, notStaged, err := classifyFiles(dir, idx)
	if err != nil {
		return fmt.Errorf("status: classify files: %w", err)
	}

	if len(untracked) == 0 && len(staged) == 0 && len(notStaged) == 0 {
		fmt.Println("nothing to commit, working tree clean")
		return nil
	}

	printStatus(untracked, staged, notStaged)
	return nil
}

// classifyFiles walks dir and sorts each file into one of three buckets.
func classifyFiles(repoRoot string, idx *index.Index) (untracked, staged, notStaged []string, err error) {
	// Collect all regular files currently on disk.
	diskFiles := make(map[string]struct{})
	walkErr := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && d.Name() == gitterDirName {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Type().IsRegular() {
			rel, _ := filepath.Rel(repoRoot, path)
			diskFiles[filepath.ToSlash(rel)] = struct{}{}
		}
		return nil
	})
	if walkErr != nil {
		return nil, nil, nil, walkErr
	}

	// Files in the index.
	for relPath, entry := range idx.Entries {
		absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
		diskHash, hashErr := index.HashFile(absPath)
		if hashErr != nil {
			// File was deleted after staging — treat as not staged (deleted).
			notStaged = append(notStaged, relPath)
			continue
		}

		staged = append(staged, relPath)
		if diskHash != entry.Hash {
			// Content changed on disk since staging.
			notStaged = append(notStaged, relPath)
		}
	}

	// Files on disk not in the index.
	for relPath := range diskFiles {
		if _, inIndex := idx.Entries[relPath]; !inIndex {
			untracked = append(untracked, relPath)
		}
	}

	sort.Strings(untracked)
	sort.Strings(staged)
	sort.Strings(notStaged)
	return untracked, staged, notStaged, nil
}

// printStatus formats and writes the status output to stdout.
func printStatus(untracked, staged, notStaged []string) {
	var sb strings.Builder

	if len(untracked) > 0 {
		sb.WriteString("Untracked files:\n")
		sb.WriteString("  (use \"gitter add <file>...\" to include in what will be committed)\n")
		sb.WriteString("\n")
		for _, f := range untracked {
			sb.WriteString("\t" + f + "\n")
		}
		sb.WriteString("\n")
	}

	if len(staged) > 0 {
		sb.WriteString("Changes to be committed:\n")
		sb.WriteString("  (use \"gitter restore --staged <file>...\" to unstage)\n")
		sb.WriteString("\n")
		for _, f := range staged {
			sb.WriteString("\tnew file:   " + f + "\n")
		}
		sb.WriteString("\n")
	}

	if len(notStaged) > 0 {
		sb.WriteString("Changes not staged for commit:\n")
		sb.WriteString("  (use \"gitter add <file>...\" to update what will be committed)\n")
		sb.WriteString("\n")
		for _, f := range notStaged {
			sb.WriteString("\tmodified:   " + f + "\n")
		}
		sb.WriteString("\n")
	}

	fmt.Print(sb.String())
}
