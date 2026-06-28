// Package commands — internal test so we can call the unexported runInDir directly,
// avoiding os.Chdir which mutates global process state and breaks parallel tests.
package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdoutInternal redirects os.Stdout for the duration of fn and returns output.
func captureStdoutInternal(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestInit_firstTime_createsObjectsDir(t *testing.T) {
	dir := t.TempDir()
	cmd := &InitCommand{}

	captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if _, err := os.Stat(filepath.Join(dir, ".gitter", "objects")); os.IsNotExist(err) {
		t.Error("expected .gitter/objects to exist after init")
	}
}

func TestInit_firstTime_createsRefsHeadsDir(t *testing.T) {
	dir := t.TempDir()
	cmd := &InitCommand{}

	captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if _, err := os.Stat(filepath.Join(dir, ".gitter", "refs", "heads")); os.IsNotExist(err) {
		t.Error("expected .gitter/refs/heads to exist after init")
	}
}

func TestInit_firstTime_createsHEADWithCorrectContent(t *testing.T) {
	dir := t.TempDir()
	cmd := &InitCommand{}

	captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	headPath := filepath.Join(dir, ".gitter", "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		t.Fatalf("HEAD file not found: %v", err)
	}
	if string(data) != headFileContent {
		t.Errorf("HEAD content = %q, want %q", string(data), headFileContent)
	}
}

func TestInit_firstTime_printsInitializedMessage(t *testing.T) {
	dir := t.TempDir()
	cmd := &InitCommand{}

	out := captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Initialized empty Gitter repository") {
		t.Errorf("expected init message, got: %q", out)
	}
	// Output must include the path to the .gitter directory.
	if !strings.Contains(out, filepath.Join(dir, ".gitter")) {
		t.Errorf("expected path in init message, got: %q", out)
	}
}

func TestInit_reinit_printsAlreadyInitialisedMessage(t *testing.T) {
	dir := t.TempDir()
	cmd := &InitCommand{}

	// First init.
	captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	// Second init (re-init).
	out := captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "already initialised") {
		t.Errorf("expected 'already initialised' message on re-init, got: %q", out)
	}
}

func TestInit_reinit_doesNotModifyExistingHEAD(t *testing.T) {
	dir := t.TempDir()
	cmd := &InitCommand{}

	captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	// Overwrite HEAD with a custom ref to verify re-init won't touch it.
	headPath := filepath.Join(dir, ".gitter", "HEAD")
	customContent := "ref: refs/heads/feature\n"
	if err := os.WriteFile(headPath, []byte(customContent), 0o644); err != nil {
		t.Fatal(err)
	}

	captureStdoutInternal(t, func() {
		if err := cmd.runInDir(dir); err != nil {
			t.Fatal(err)
		}
	})

	data, err := os.ReadFile(headPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != customContent {
		t.Errorf("re-init modified HEAD: got %q, want %q", string(data), customContent)
	}
}
