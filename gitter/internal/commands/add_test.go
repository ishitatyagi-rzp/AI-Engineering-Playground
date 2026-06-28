// Package commands — internal test so we can call the unexported runInDir directly.
package commands

import (
	"os"
	"path/filepath"
	"testing"

	"gitter/internal/index"
)

// initRepo creates a minimal .gitter/ skeleton in dir so add/status commands can run.
func initRepo(t *testing.T, dir string) {
	t.Helper()
	gitterDir := filepath.Join(dir, ".gitter")
	for _, sub := range []string{"objects", filepath.Join("refs", "heads")} {
		if err := os.MkdirAll(filepath.Join(gitterDir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(gitterDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeFile creates path (with any needed parent dirs) and writes content to it.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAdd_dotArg_stagesAllFiles(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "a.txt"), "aaa")
	writeFile(t, filepath.Join(dir, "sub", "b.txt"), "bbb")

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{"."}); err != nil {
		t.Fatal(err)
	}

	idx, err := index.Load(filepath.Join(dir, ".gitter"))
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"a.txt", "sub/b.txt"} {
		if _, ok := idx.Entries[rel]; !ok {
			t.Errorf("expected %q in index after add .", rel)
		}
	}
}

func TestAdd_dotArg_skipsGitterDir(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "src.go"), "package main")

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{"."}); err != nil {
		t.Fatal(err)
	}

	idx, _ := index.Load(filepath.Join(dir, ".gitter"))
	for rel := range idx.Entries {
		if len(rel) >= len(".gitter") && rel[:len(".gitter")] == ".gitter" {
			t.Errorf("index must not contain .gitter entries, got %q", rel)
		}
	}
}

func TestAdd_globPattern_stagesMatchingFiles(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "main.go"), "package main")
	writeFile(t, filepath.Join(dir, "util.go"), "package main")
	writeFile(t, filepath.Join(dir, "readme.md"), "# hi")

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{"*.go"}); err != nil {
		t.Fatal(err)
	}

	idx, _ := index.Load(filepath.Join(dir, ".gitter"))
	if _, ok := idx.Entries["main.go"]; !ok {
		t.Error("expected main.go in index")
	}
	if _, ok := idx.Entries["util.go"]; !ok {
		t.Error("expected util.go in index")
	}
	if _, ok := idx.Entries["readme.md"]; ok {
		t.Error("readme.md must not be staged by *.go glob")
	}
}

func TestAdd_globPattern_noMatch_noError(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{"*.py"}); err != nil {
		t.Errorf("expected no error on zero matches, got: %v", err)
	}
}

func TestAdd_noArgs_noError(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{}); err != nil {
		t.Errorf("expected no error on empty args, got: %v", err)
	}
}

func TestAdd_persistsToIndexFile(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "hello.txt"), "hello")

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{"."}); err != nil {
		t.Fatal(err)
	}

	// index.json must exist on disk.
	indexPath := filepath.Join(dir, ".gitter", "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("expected .gitter/index.json to be written after add")
	}
}

func TestAdd_noRepo_returnsError(t *testing.T) {
	dir := t.TempDir() // no initRepo — no .gitter/

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{"."}); err == nil {
		t.Error("expected error when .gitter/ is missing, got nil")
	}
}

func TestAdd_literalFilePath_stagesThatFile(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "specific.txt"), "content")
	writeFile(t, filepath.Join(dir, "other.txt"), "other")

	cmd := &AddCommand{}
	if err := cmd.runInDir(dir, []string{"specific.txt"}); err != nil {
		t.Fatal(err)
	}

	idx, _ := index.Load(filepath.Join(dir, ".gitter"))
	if _, ok := idx.Entries["specific.txt"]; !ok {
		t.Error("expected specific.txt in index")
	}
	if _, ok := idx.Entries["other.txt"]; ok {
		t.Error("other.txt must not be staged")
	}
}
