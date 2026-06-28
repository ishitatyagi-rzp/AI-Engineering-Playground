package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_missingFile_returnsEmptyIndex(t *testing.T) {
	dir := t.TempDir()

	idx, err := Load(dir)
	if err != nil {
		t.Fatalf("expected no error for missing index file, got: %v", err)
	}
	if len(idx.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(idx.Entries))
	}
}

func TestLoad_validFile_parsesEntries(t *testing.T) {
	dir := t.TempDir()

	raw := `{"entries":{"src/main.go":{"hash":"abc123","staged_as":"new file"}}}`
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	idx, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry, ok := idx.Entries["src/main.go"]
	if !ok {
		t.Fatal("expected entry for src/main.go")
	}
	if entry.Hash != "abc123" {
		t.Errorf("hash = %q, want %q", entry.Hash, "abc123")
	}
	if entry.StagedAs != "new file" {
		t.Errorf("staged_as = %q, want %q", entry.StagedAs, "new file")
	}
}

func TestSave_writesValidJSON(t *testing.T) {
	dir := t.TempDir()

	idx := &Index{Entries: map[string]Entry{
		"README.md": {Hash: "deadbeef", StagedAs: "new file"},
	}}

	if err := Save(dir, idx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	entry, ok := loaded.Entries["README.md"]
	if !ok {
		t.Fatal("expected README.md entry after save+load round-trip")
	}
	if entry.Hash != "deadbeef" {
		t.Errorf("round-trip hash = %q, want %q", entry.Hash, "deadbeef")
	}
}

func TestHashFile_knownContent_returnsExpectedSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")

	// SHA-256 of "hello\n" is well-known.
	content := "hello\n"
	wantHash := "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := HashFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != wantHash {
		t.Errorf("hash = %q, want %q", got, wantHash)
	}
}

func TestHashFile_emptyFile_returnsEmptySHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")

	// SHA-256 of empty string.
	wantHash := "e3b0c44298fc1c149afbf4c8996fb924" +
		"27ae41e4649b934ca495991b7852b855"

	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := HashFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != wantHash {
		t.Errorf("hash = %q, want %q", got, wantHash)
	}
}
