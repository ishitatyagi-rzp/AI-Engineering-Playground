package index

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const indexFileName = "index.json"

// Entry holds the staged state of a single file.
type Entry struct {
	// Hash is the hex-encoded SHA-256 of the file content at staging time.
	Hash string `json:"hash"`
	// StagedAs is the classification written in status output ("new file").
	StagedAs string `json:"staged_as"`
}

// Index is the in-memory representation of .gitter/index.json.
// Keys in Entries are slash-separated paths relative to the repo root.
type Index struct {
	Entries map[string]Entry `json:"entries"`
}

// Load reads the index from gitterDir/index.json.
// Returns an empty Index if the file does not exist yet.
func Load(gitterDir string) (*Index, error) {
	path := filepath.Join(gitterDir, indexFileName)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Index{Entries: make(map[string]Entry)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("index: read %s: %w", path, err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("index: parse %s: %w", path, err)
	}
	if idx.Entries == nil {
		idx.Entries = make(map[string]Entry)
	}
	return &idx, nil
}

// Save writes the index to gitterDir/index.json atomically.
func Save(gitterDir string, idx *Index) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("index: marshal: %w", err)
	}

	// Write to a temp file first, then rename for atomicity.
	tmpPath := filepath.Join(gitterDir, indexFileName+".tmp")
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("index: write temp file: %w", err)
	}

	dest := filepath.Join(gitterDir, indexFileName)
	if err := os.Rename(tmpPath, dest); err != nil {
		return fmt.Errorf("index: rename to %s: %w", dest, err)
	}
	return nil
}

// HashFile returns the hex-encoded SHA-256 of the file at path.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("index: open %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("index: hash %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
