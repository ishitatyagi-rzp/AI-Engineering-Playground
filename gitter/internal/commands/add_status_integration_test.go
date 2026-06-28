// Package commands — integration tests verifying the add → status pipeline.
// These tests exercise both commands together to confirm that index.json written
// by AddCommand is correctly read back by StatusCommand.
package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_AddDot_ThenStatus_ShowsStaged(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "main.go"), "package main")
	writeFile(t, filepath.Join(dir, "README.md"), "# gitter")

	if err := (&AddCommand{}).runInDir(dir, []string{"."}); err != nil {
		t.Fatal(err)
	}

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Changes to be committed:") {
		t.Errorf("expected staged section, got:\n%s", out)
	}
	if !strings.Contains(out, "main.go") {
		t.Errorf("expected main.go in staged section, got:\n%s", out)
	}
	if !strings.Contains(out, "README.md") {
		t.Errorf("expected README.md in staged section, got:\n%s", out)
	}
	if strings.Contains(out, "Untracked files:") {
		t.Errorf("no files should be untracked after add ., got:\n%s", out)
	}
}

func TestIntegration_AddGlob_ThenStatus_ShowsOnlyMatched(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "app.go"), "package main")
	writeFile(t, filepath.Join(dir, "config.yaml"), "key: val")

	if err := (&AddCommand{}).runInDir(dir, []string{"*.go"}); err != nil {
		t.Fatal(err)
	}

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "app.go") {
		t.Errorf("expected app.go staged, got:\n%s", out)
	}
	if !strings.Contains(out, "Untracked files:") {
		t.Errorf("expected config.yaml to be untracked, got:\n%s", out)
	}
	if !strings.Contains(out, "config.yaml") {
		t.Errorf("expected config.yaml in untracked list, got:\n%s", out)
	}
}

func TestIntegration_ModifyAfterAdd_ShowsModifiedSection(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	filePath := filepath.Join(dir, "service.go")
	writeFile(t, filePath, "v1")

	if err := (&AddCommand{}).runInDir(dir, []string{"service.go"}); err != nil {
		t.Fatal(err)
	}

	// Modify the file after staging — new content differs from stored hash.
	writeFile(t, filePath, "v2 changed content")

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Changes not staged for commit:") {
		t.Errorf("expected not-staged section, got:\n%s", out)
	}
	if !strings.Contains(out, "modified:") {
		t.Errorf("expected 'modified:' label, got:\n%s", out)
	}
}

func TestIntegration_IndexPersistsAcrossCommandInstances(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "persist.txt"), "data")

	// Simulate separate process: new struct instances on each invocation.
	if err := (&AddCommand{}).runInDir(dir, []string{"."}); err != nil {
		t.Fatal(err)
	}

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if strings.Contains(out, "nothing to commit") {
		t.Errorf("staged file should appear in status, got:\n%s", out)
	}
	if !strings.Contains(out, "persist.txt") {
		t.Errorf("expected persist.txt in status output, got:\n%s", out)
	}
}

func TestIntegration_MultipleAdds_AccumulateInIndex(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, filepath.Join(dir, "first.go"), "v1")
	writeFile(t, filepath.Join(dir, "second.go"), "v2")

	// Stage files in two separate add calls.
	if err := (&AddCommand{}).runInDir(dir, []string{"first.go"}); err != nil {
		t.Fatal(err)
	}
	if err := (&AddCommand{}).runInDir(dir, []string{"second.go"}); err != nil {
		t.Fatal(err)
	}

	out := captureStdoutInternal(t, func() {
		if err := (&StatusCommand{}).runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "first.go") {
		t.Errorf("expected first.go in staged section, got:\n%s", out)
	}
	if !strings.Contains(out, "second.go") {
		t.Errorf("expected second.go in staged section, got:\n%s", out)
	}
}
