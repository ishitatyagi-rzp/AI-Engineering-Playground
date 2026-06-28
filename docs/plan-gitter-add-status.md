# Plan: `gitter add` and `gitter status` Implementation

## Problem Statement

Implement the `add` and `status` commands in the existing `gitter` CLI. These two commands must work together across separate OS process invocations — `gitter add` stages files, then `gitter status` reads that staged state and reports it. This requires persistent storage between invocations.

---

## Clarification / Assumptions

1. **Index persistence via `.gitter/index.json`**: Each `gitter` invocation is a separate OS process; in-memory state cannot survive between `add` and `status`. The index must be written to disk. This is analogous to Git's `.git/index` binary file — we'll use JSON for simplicity.
2. **No commits yet**: Phase 2 does not implement `commit`, so all staged files are classified as `new file:` (never-committed). There is no HEAD tree to diff against.
3. **`gitter add .`**: Recursively stages all regular files under CWD, skipping `.gitter/` itself.
4. **`gitter add <pattern>`**: Uses `filepath.Glob` against CWD (non-recursive). If the argument is a literal file path, it stages that file. If the pattern matches a directory, walk it recursively. No error on zero matches.
5. **File modification detection**: SHA-256 hash of file content at staging time, stored in the index. If the on-disk hash differs from the stored hash at `status` time, the file appears in "Changes not staged for commit".
6. **`gitter status` run from non-repo directory**: Return an error `"not a gitter repository"`.

---

## Architecture

```
gitter/
├── cmd/
│   └── main.go                        # unchanged
├── internal/
│   ├── cli/                           # unchanged
│   ├── index/
│   │   └── index.go                   # NEW: Index type — load, save, add, status
│   └── commands/
│       ├── add.go                     # UPDATED: real implementation
│       └── status.go                  # UPDATED: real implementation
└── docs/
    └── plan-gitter-add-status.md      # this file
```

### Why a separate `internal/index/` package?

- Keeps `commands/add.go` and `commands/status.go` thin (resolve CWD + delegate).
- The `Index` type has its own serialisation concerns (`Load`/`Save`) — mixing this into a command would violate separation of concerns.
- Future `commit` command will also need to read from the index; sharing via a dedicated package is cleaner.

---

## Index JSON Schema

**File path**: `.gitter/index.json`

```json
{
  "entries": {
    "src/main.go": {
      "hash": "e3b0c44298fc1c149afbf4c8996fb924...",
      "staged_as": "new file"
    },
    "README.md": {
      "hash": "d41d8cd98f00b204e9800998ecf8427e...",
      "staged_as": "new file"
    }
  }
}
```

**Go types** in `internal/index/index.go`:

```go
type Entry struct {
    Hash     string `json:"hash"`      // SHA-256 hex of file content at stage time
    StagedAs string `json:"staged_as"` // "new file" (only value for Phase 2)
}

type Index struct {
    Entries map[string]Entry `json:"entries"` // key: path relative to repo root
}
```

- Keys are always **relative to the repo root** (where `.gitter/` lives), using forward slashes (`filepath.ToSlash` on Windows normalisation for future portability).
- `hash` is the hex-encoded SHA-256 of the file's full content at staging time.

---

## File State Machine

```
        [file on disk, not in index]
                     │
             gitter add <file>
                     │
                     ▼
        [index entry created, hash = H1]
          → "Changes to be committed: new file: <rel>"
                     │
           file content changes on disk
                     │
                     ▼
        [index entry exists, disk hash = H2 ≠ H1]
          → ALSO "Changes not staged: modified: <rel>"
```

- A file not in the index and present on disk → **Untracked**
- A file in the index (hash matches disk) → **Changes to be committed**
- A file in the index (hash differs from disk) → **Changes to be committed** AND **Changes not staged**

---

## `gitter add` Logic

### Entry point: `AddCommand.Run(args []string)`

1. Resolve CWD (`os.Getwd()`).
2. Find the repo root by looking for `.gitter/` in CWD (for Phase 2, CWD == repo root; error if not found).
3. Load the index from `.gitter/index.json` (empty index if file does not exist).
4. For each arg:
   - If arg is `"."` → `collectAll(cwd)`: `filepath.Walk(cwd)`, skip `.gitter/`, collect all regular files.
   - Otherwise → `collectGlob(cwd, arg)`: `filepath.Glob(filepath.Join(cwd, arg))`. For each match: if directory, walk recursively; otherwise add the single file. Silently succeed if zero matches.
5. For each collected file path:
   - Compute SHA-256 of file content.
   - Upsert `index.Entries[relPath] = Entry{Hash: hex, StagedAs: "new file"}`.
6. Save the updated index to `.gitter/index.json`.
7. **No output** on success.

### `collectAll(root string) ([]string, error)`

```
filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
    if info.IsDir() && info.Name() == ".gitter" {
        return filepath.SkipDir
    }
    if !info.IsDir() {
        files = append(files, path)
    }
    return nil
})
```

### `collectGlob(cwd, pattern string) ([]string, error)`

```
matches, _ := filepath.Glob(filepath.Join(cwd, pattern))
for each match:
    if directory: walk recursively (same skip-.gitter logic)
    else: append to files
```

---

## `gitter status` Logic

### Entry point: `StatusCommand.Run(args []string)`

1. Resolve CWD.
2. Locate repo root (`.gitter/` in CWD); return error if not found.
3. Load index.
4. Walk the entire working tree (skip `.gitter/`), collect all regular file paths.
5. Classify each file:

```
For each file path (relative to repo root):
    inIndex = entry exists in index
    diskHash = SHA-256 of current content
    storedHash = entry.Hash (if inIndex)

    if !inIndex:
        → untracked
    elif inIndex && diskHash == storedHash:
        → staged (to be committed)
    elif inIndex && diskHash != storedHash:
        → staged (to be committed) AND not staged (modified)
```

6. Format output (see below).

---

## `gitter status` Output Format

### Case 1 — Nothing to report

```
nothing to commit, working tree clean
```

### Case 2 — Has content

Print only non-empty sections, in this order:

```
Untracked files:
  (use "gitter add <file>..." to include in what will be committed)

	<rel-path>

Changes to be committed:
  (use "gitter restore --staged <file>..." to unstage)

	new file:   <rel-path>

Changes not staged for commit:
  (use "gitter add <file>..." to update what will be committed)

	modified:   <rel-path>

```

**Format rules** (derived from test regex patterns):
- Section header: `"Untracked files:\n"`, `"Changes to be committed:\n"`, `"Changes not staged for commit:\n"`
- Hint line: two-space indent, e.g. `"  (use \"gitter add <file>...\" to include in what will be committed)\n"`
- Blank line after the hint line.
- File entries: one tab indent, then the entry.
  - Untracked: `"\t<rel-path>\n"`
  - Staged: `"\tnew file:   <rel-path>\n"` (3 spaces after colon to align with `modified:`)
  - Modified: `"\tmodified:   <rel-path>\n"` (3 spaces after colon)
- Blank line after each section's last file entry.
- Files within each section are sorted lexicographically for deterministic output.

---

## Implementation Order

1. **`internal/index/index.go`**: `Entry`, `Index`, `Load(gitterDir)`, `Save(gitterDir)`, `HashFile(path)`.
2. **`internal/commands/add.go`**: implement `Run` — resolve repo, load index, collect files, hash, save.
3. **`internal/commands/status.go`**: implement `Run` — resolve repo, load index, walk tree, classify, format output.
4. **Unit tests** for `internal/index/`:
   - `TestLoad_missingFile_returnsEmptyIndex`
   - `TestLoad_validFile_parsesEntries`
   - `TestSave_writesValidJSON`
   - `TestHashFile_knownContent_returnsExpectedSHA256`
5. **Unit tests** for `add` command (`internal/commands/add_test.go`, `package commands`):
   - `TestAdd_dotArg_stagesAllFiles`
   - `TestAdd_dotArg_skipsGitterDir`
   - `TestAdd_globPattern_stagesMatchingFiles`
   - `TestAdd_globPattern_noMatch_noError`
   - `TestAdd_noArgs_noError`
   - `TestAdd_persistsToIndexFile`
6. **Unit tests** for `status` command (`internal/commands/status_test.go`, `package commands`):
   - `TestStatus_noRepo_returnsError`
   - `TestStatus_emptyIndex_noUntrackedFiles_printsClean`
   - `TestStatus_untrackedFiles_printsUntrackedSection`
   - `TestStatus_stagedNewFile_printsStagedSection`
   - `TestStatus_modifiedAfterStage_printsBothSections`
   - `TestStatus_allThreeSections_correctOrder`
7. **Integration tests** (`add_status_integration_test.go`, `package commands_test`):
   - `TestIntegration_AddDot_ThenStatus_ShowsStaged`
   - `TestIntegration_AddGlob_ThenStatus_ShowsOnlyMatched`
   - `TestIntegration_ModifyAfterAdd_ShowsModifiedSection`
   - `TestIntegration_NothingToCommit_AfterAllFilesStaged`

---

## Testing Strategy Details

### `internal/index/` tests
Use `t.TempDir()` to create a fake `.gitter/` dir. Call `Load`/`Save` directly. No commands invoked.

### `commands/add_test.go` (internal package)
- Introduce `runAddInDir(dir string, args []string) error` on `AddCommand` (parallel to `runInDir` pattern used in `init.go`).
- Tests call `cmd.runAddInDir(tmpDir, args)` directly — no `os.Chdir`, safe for parallel tests.

### `commands/status_test.go` (internal package)
- Introduce `runStatusInDir(dir string) error` on `StatusCommand`.
- Tests create the `.gitter/` dir + `index.json` fixture manually, then call `runStatusInDir`.

### Integration tests (external package)
- Build entire pipeline: `gitter init` → `gitter add` → `gitter status` using direct method calls (not binary invocation).
- Share one `t.TempDir()` across the full sequence to verify cross-command index persistence.

---

## Extensibility Notes for `commit`

- `commit` will need to: read the index, create a commit object in `.gitter/objects/`, clear the index (or mark entries as committed), and update `refs/heads/main`.
- The `Entry.StagedAs` field is designed to accommodate future values (`"modified"`, `"deleted"`) — only `"new file"` is used in Phase 2.
- No changes to `internal/index/` API are expected for Phase 3 — `commit` will call `Load`, iterate entries, then `Save` with cleared entries.

---

## Open Questions

| # | Question | Assumption (if unblocked) |
|---|----------|--------------------------|
| 1 | Should `gitter add` without arguments be a no-op or error? | No-op — matches Git's behaviour; tests don't require an error. |
| 2 | Should `gitter add <non-existent-literal-file>` error? | No error — consistent with "no error if no match" requirement for globs. |
| 3 | Should `gitter status` be runnable from a subdirectory of the repo? | Out of scope for Phase 2; only CWD == repo root is tested. |
| 4 | Symlinks in working tree — follow or skip? | Skip — `filepath.Walk` follows symlinks by default in Go; use `os.Lstat` check to skip. |
