# Plan: Gitter CLI Tool

## Problem Statement

Build a Git-like CLI tool called `gitter` in Go. Phase 1 covers three commands:
- `gitter help` — list all supported commands
- `gitter help <command>` — detailed help for a specific command
- `gitter init` — create an empty Gitter repository in the current directory

The tool must be extensible (easy to add new commands), match exact output formats required by the test suite, and behave analogously to Git internals.

---

## Clarification / Assumption

> `CLAUDE.md` states "no file system persistence," but `gitter init` must create a `.gitter/` directory on disk (tests verify its path). This plan treats filesystem access as a **core requirement** for gitter — equivalent to how Git works — not as an "external database." All other business-logic data (e.g., staging index, commit graph) would remain in-memory until a future command explicitly requires persistence.

---

## Approach Options

### Option A — Cobra-based CLI

Use the [github.com/spf13/cobra](https://github.com/spf13/cobra) framework, which is the de-facto standard for Go CLIs (used by Kubernetes, Docker, Hugo).

**Pros:**
- Battle-tested, widely used in production Go CLIs
- Built-in support for subcommands, flag parsing, help text templating
- Easy to extend: one file per command
- Cobra's help template can be overridden to produce git-style output

**Cons:**
- Introduces an external dependency (justified, but non-zero)
- Default help format doesn't match the required output — needs template customisation
- Cobra's auto-generated help sections (`Flags:`, `Use:`, etc.) can bleed through if templates are not fully replaced
- Slightly harder to enforce the exact section ordering required by tests (`NAME:`, `SYNOPSIS:`, `DESCRIPTION:`, `OPTIONS:`)

**Verdict:** Good choice for large CLIs, but the custom output format required by the tests makes template overriding fragile and adds cognitive overhead.

---

### Option B — urfave/cli

Use [github.com/urfave/cli](https://github.com/urfave/cli), a lighter CLI framework.

**Pros:**
- Simpler API than Cobra
- Has a built-in help template system

**Cons:**
- Same template-overriding issues as Cobra
- Less widely used; weaker community support
- Still adds an external dependency

**Verdict:** Not preferred over Cobra; doesn't solve the output-format problem either.

---

### Option C — Custom Command Registry (no external deps) ✅ Recommended

Implement a hand-rolled command registry using Go interfaces and a `map[string]Command` lookup. Each command is a struct that satisfies a `Command` interface.

```
Command interface
  Name()             string
  ShortDescription() string
  Synopsis()         string
  LongDescription()  string
  Options()          []Option
  Run(args []string) error
```

The main dispatcher:
1. Reads `os.Args`
2. Looks up the command in the registry
3. Delegates to `cmd.Run(args)`

Help is just another command in the registry that reads metadata from other commands — no magic.

**Pros:**
- Zero external dependencies
- Full control over output format — tests pass deterministically
- Adding a new command = create one file, implement the interface, register it
- Clean separation: each command owns its own metadata and execution logic
- No template overriding or framework fighting

**Cons:**
- More boilerplate vs Cobra for flag parsing (mitigated by using stdlib `flag` package per command)
- No auto-generated shell completions (not a current requirement)

**Verdict:** Best fit. Output format is tested precisely; a custom registry gives exact control. Extensibility is first-class: the registry is the single extension point.

---

### Option D — Cobra + Custom Help Command (Hybrid)

Use Cobra for subcommand routing and flag parsing, but implement `help` as a fully custom command that reads Cobra's command metadata.

**Pros:**
- Gets Cobra's flag parsing for free
- Custom help output

**Cons:**
- Two competing help mechanisms (Cobra's built-in + custom), risk of conflicts
- More complex than either pure option
- Adds external dependency

**Verdict:** Overengineered for current scope.

---

## Recommended Approach: Option C — Custom Command Registry

---

## Architecture

```
gitter/
├── cmd/
│   └── main.go                  # Entry point: parses os.Args, dispatches to registry
├── internal/
│   ├── cli/
│   │   ├── command.go           # Command interface + Option type
│   │   ├── registry.go          # Registry: stores and looks up commands
│   │   └── dispatcher.go        # Dispatcher: routes os.Args to the right command
│   └── commands/
│       ├── help.go              # `help` command implementation
│       ├── init.go              # `init` command implementation
│       ├── add.go               # `add` command stub (Phase 1: help text only)
│       └── status.go            # `status` command stub (Phase 1: help text only)
├── docs/
│   └── plan-gitter.md
├── go.mod
└── go.sum
```

### Why `internal/cli/` vs `internal/commands/`

- `cli/` owns the **framework**: interface definitions, registry, dispatcher. It never knows about specific commands.
- `commands/` owns the **implementations**: each command is self-contained. Adding a new command means adding one file here and one `Register()` call in `main.go`.

---

## Component Breakdown

### `internal/cli/command.go`
Defines the `Command` interface and the `Option` struct (name + description pairs).

```go
type Option struct {
    Flag        string
    Description string
}

type Command interface {
    Name()             string
    ShortDescription() string   // one-line summary for `gitter help` list
    Synopsis()         string   // usage line for `gitter help <cmd>`
    LongDescription()  string   // body for DESCRIPTION section
    Options()          []Option // nil/empty = no OPTIONS section printed
    Run(args []string) error
}
```

### `internal/cli/registry.go`
A `Registry` struct with `Register(Command)` and `Get(name string) (Command, bool)` and `All() []Command` (sorted by name for deterministic output).

### `internal/cli/dispatcher.go`
`Dispatch(registry, args []string)` — parses `os.Args[1:]`, handles the `help` special case, routes to the right command, or prints an "unknown command" error.

### `internal/commands/help.go`
Implements `Command`. Its `Run` method:
- With no args: prints the formatted list of all registered commands.
- With one arg (a command name): prints NAME / SYNOPSIS / DESCRIPTION / OPTIONS sections.

### `internal/commands/init.go`
Implements `Command`. Its `Run` method:
- Checks for an existing `.gitter/` directory.
- Creates `.gitter/` with sub-structure: `objects/`, `refs/heads/`, `HEAD` (pointing to `refs/heads/main`).
- Prints the appropriate message.

### `internal/commands/add.go` / `status.go`
Phase 1: implement the `Command` interface (metadata only) so they appear in `gitter help`. `Run` returns a "not yet implemented" message. This keeps the help output complete without dead code.

---

## `.gitter/` Directory Structure

Mirrors Git's `.git/` layout for future extensibility:

```
.gitter/
├── HEAD                # "ref: refs/heads/main\n"
├── objects/            # content-addressable object store (future: blobs, trees, commits)
└── refs/
    └── heads/
        └── main        # branch pointer (future: SHA of tip commit)
```

---

## Output Format Contract

### `gitter help`

```
These are common Gitter commands:
 init     Create an empty Gitter repository
 add      Add file contents to the index
 status   Show the working tree status
```

- Two-space indent before command name
- Name and description separated by spaces (column-aligned)
- Commands printed in registration order (consistent with test regex)

### `gitter help <command>`

```
NAME:
<name> - <short description>

SYNOPSIS:
gitter <synopsis>

DESCRIPTION:
<long description>

OPTIONS:
-<flag>: <description>
```

- `OPTIONS:` section only printed when `Options()` returns non-empty slice
- Blank line between sections

---

## Implementation Order

1. **Scaffold**: `go.mod`, directory structure, empty files
2. **`internal/cli/command.go`**: interface + Option type
3. **`internal/cli/registry.go`**: Registry type
4. **`internal/cli/dispatcher.go`**: Dispatcher
5. **`internal/commands/help.go`**: full help output logic
6. **`internal/commands/init.go`**: full init logic
7. **`internal/commands/add.go`** + **`status.go`**: stubs (metadata only)
8. **`cmd/main.go`**: wire everything together
9. **Tests**: unit + integration (see below)

---

## Testing Strategy

### Unit Tests

| File | What to test |
|------|--------------|
| `registry_test.go` | Register, Get (found/not found), All (order) |
| `dispatcher_test.go` | Routes to correct command, unknown command error, help dispatch |
| `help_test.go` | `gitter help` output format; each command present; `gitter help init/add/status/commit` sections |
| `init_test.go` | First-time init creates correct dirs + files; re-init returns correct message |

### Test Scenarios (Mapped to Requirements)

**Test 1 — `gitter help` lists commands**
- Run `help` with no args; assert output contains `init`, `add`, `status` with two-space indent and description.

**Test 2 — `gitter help <command>` (no options)**
- `gitter help init`: assert `NAME:`, `SYNOPSIS:`, `DESCRIPTION:` are all present.
- `gitter help add`: same.
- `gitter help status`: same.

**Test 3 — `gitter help <command>` (with options)**
- `gitter help commit`: assert `NAME:`, `SYNOPSIS:`, `DESCRIPTION:`, `OPTIONS:` all present.
- `gitter help checkout`: same.

**Test 4 — `gitter init` first time**
- Create a temp dir, run `init.Run([]string{})` with working dir set.
- Assert output matches `Initialized empty Gitter repository in <dir>/.gitter/`
- Assert `.gitter/HEAD`, `.gitter/objects/`, `.gitter/refs/heads/` exist.

**Test 5 — `gitter init` re-init**
- Run `init.Run` twice in same temp dir.
- Second run: assert output matches `Gitter repository is already initialised in <dir>/.gitter/`

---

## Open Decisions

| # | Decision | Recommendation |
|---|----------|----------------|
| 1 | Filesystem persistence for `.gitter/` | Required — VCS fundamentally needs disk. Treat as exception to in-memory constraint. |
| 2 | `add` and `status` in Phase 1 | Register as stubs with full help metadata. `Run` prints "not implemented". |
| 3 | Binary installation for integration tests | `go build -o gitter ./cmd/` then tests invoke the binary via `os/exec`. |
| 4 | Column alignment in `gitter help` | Use fixed-width padding (longest command name + 2 spaces) for clean columnar output. |

---

## Questions Before Implementation

1. **Filesystem exception confirmed?** Can `.gitter/` directory creation be treated as an explicit exception to the "no filesystem persistence" rule in `CLAUDE.md`?

2. **`add` and `status` stubs** — Phase 1 only requires help text for these. Should `Run` print "not implemented" or silently succeed (exit 0)?

3. **`gitter help commit` and `gitter help checkout`** — The test requires OPTIONS section for these two commands, but their `Run` logic is not scoped in Phase 1. Should I add them as stubs with full metadata (so help tests pass) but `Run` returning "not implemented"?

4. **Binary in PATH** — Integration tests run `gitter init` as a shell command. Should the build step (`go install`) be part of the test setup, or do you want the tests to call the Go source directly via `os/exec go run`?
