// Package commands — internal test so we can call the unexported runInDir directly.
package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"gitter/internal/index"
)

func TestStatus_noRepo_returnsError(t *testing.T) {
	dir := t.TempDir()

	cmd := &StatusCommand{}
	if err := cmd.runInDir(dir); err == nil {
		t.Error("expected error when .gitter/ is missing, got nil")
	}
}

func TestStatus_emptyIndex_noUntrackedFiles_printsClean(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	// No files on disk, empty index.

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "nothing to commit, working tree clean") {
		t.Errorf("expected clean message, got: %q", out)
	}
}

func TestStatus_untrackedFiles_printsUntrackedSection(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "untracked.txt"), "data")

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Untracked files:") {
		t.Errorf("expected Untracked files: section, got: %q", out)
	}
	if !strings.Contains(out, "untracked.txt") {
		t.Errorf("expected untracked.txt in output, got: %q", out)
	}
}

func TestStatus_stagedNewFile_printsStagedSection(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "staged.go"), "package main")

	// Stage the file.
	if err := (&AddCommand{}).runInDir(dir, []string{"staged.go"}); err != nil {
		t.Fatal(err)
	}

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Changes to be committed:") {
		t.Errorf("expected Changes to be committed: section, got: %q", out)
	}
	if !strings.Contains(out, "new file:") {
		t.Errorf("expected 'new file:' label, got: %q", out)
	}
	if !strings.Contains(out, "staged.go") {
		t.Errorf("expected staged.go in output, got: %q", out)
	}
}

func TestStatus_stagedFile_doesNotAppearAsUntracked(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "file.go"), "package main")

	if err := (&AddCommand{}).runInDir(dir, []string{"file.go"}); err != nil {
		t.Fatal(err)
	}

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if strings.Contains(out, "Untracked files:") {
		t.Errorf("staged file must not appear as untracked, got: %q", out)
	}
}

func TestStatus_modifiedAfterStage_printsBothSections(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	filePath := filepath.Join(dir, "mod.go")
	writeFile(t, filePath, "v1")

	// Stage original content.
	if err := (&AddCommand{}).runInDir(dir, []string{"mod.go"}); err != nil {
		t.Fatal(err)
	}

	// Modify the file after staging.
	writeFile(t, filePath, "v2 — different content")

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Changes to be committed:") {
		t.Errorf("expected staged section, got: %q", out)
	}
	if !strings.Contains(out, "Changes not staged for commit:") {
		t.Errorf("expected not-staged section, got: %q", out)
	}
	if !strings.Contains(out, "modified:") {
		t.Errorf("expected 'modified:' label, got: %q", out)
	}
}

func TestStatus_allThreeSections_correctOrder(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	// staged.go — will be staged then modified (appears in committed + not staged)
	writeFile(t, filepath.Join(dir, "staged.go"), "v1")
	if err := (&AddCommand{}).runInDir(dir, []string{"staged.go"}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "staged.go"), "v2")

	// clean.go — staged and not modified (appears only in committed)
	writeFile(t, filepath.Join(dir, "clean.go"), "clean")
	if err := (&AddCommand{}).runInDir(dir, []string{"clean.go"}); err != nil {
		t.Fatal(err)
	}

	// untracked.txt — not in index
	writeFile(t, filepath.Join(dir, "untracked.txt"), "hi")

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	idxUntracked := strings.Index(out, "Untracked files:")
	idxToCommit := strings.Index(out, "Changes to be committed:")
	idxNotStaged := strings.Index(out, "Changes not staged for commit:")

	if idxUntracked < 0 || idxToCommit < 0 || idxNotStaged < 0 {
		t.Fatalf("one or more expected sections missing:\n%s", out)
	}
	// Sections must appear in the order: Untracked → To be committed → Not staged.
	if !(idxUntracked < idxToCommit && idxToCommit < idxNotStaged) {
		t.Errorf("section order wrong — untracked=%d, toCommit=%d, notStaged=%d\n%s",
			idxUntracked, idxToCommit, idxNotStaged, out)
	}
}

func TestStatus_cleanAfterAddWithNoChanges_printsClean(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "file.txt"), "content")

	if err := (&AddCommand{}).runInDir(dir, []string{"file.txt"}); err != nil {
		t.Fatal(err)
	}

	// Manually mark the index entry as already-committed by clearing the index.
	// We simulate "nothing new" by writing an empty index.
	emptyIdx := &index.Index{Entries: make(map[string]index.Entry)}
	if err := index.Save(filepath.Join(dir, ".gitter"), emptyIdx); err != nil {
		t.Fatal(err)
	}
	// Now the file is on disk but index is empty → untracked.
	// For the "clean" path we need both: file staged and content unchanged.
	// Re-stage it.
	if err := (&AddCommand{}).runInDir(dir, []string{"file.txt"}); err != nil {
		t.Fatal(err)
	}

	// Hash matches — only "Changes to be committed" should show.
	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if strings.Contains(out, "Changes not staged for commit:") {
		t.Errorf("should not show not-staged section when content unchanged, got:\n%s", out)
	}
}
