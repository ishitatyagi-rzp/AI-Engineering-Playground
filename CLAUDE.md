# dev-playground — Claude Configuration

This repo is used for system design problem implementations driven by AI.

## Rules

Detailed rules are in `.claude/rules/`. Follow all of them on every task.

- `.claude/rules/code-structure.md` — layering, naming, clean code
- `.claude/rules/testing.md` — TDD, test pyramid, mocking
- `.claude/rules/error-handling.md` — logging, wrapping, no silent swallows
- `.claude/rules/api-design.md` — contract-first, validation, versioning
- `.claude/rules/security.md` — input validation, no credentials in code
- `.claude/rules/database.md` — indexes, timestamps, query patterns

## Project Constraints

### Storage
- All projects in this repo use **in-memory storage only** — no external databases, no Redis, no file system persistence.
- Implement storage as a simple in-memory struct (e.g., `map` protected by a `sync.RWMutex`).
- Repository interfaces must still be defined — swap real DB for an in-memory implementation behind the same interface.
- Do not pull in any DB drivers, ORMs, or cache client dependencies.

## AI Development Principles

### 1. Never assume — always ask
If anything is unclear (requirements, edge cases, design choices), stop and ask.
Do not fill in the blanks yourself.
Keep asking and iterating until every requirement is unambiguous before writing any code.
Do not make assumptions to fill gaps — ask until there are no open questions.

### 2. Plan before code
Before writing any implementation:
1. Ask all clarifying questions upfront. Do not proceed until every question is answered.
2. Create a physical plan document at `docs/plan-<feature-name>.md` covering:
   - Problem statement
   - HLD (entities, API contracts, data flow)
   - Component breakdown (what each layer does)
   - Implementation order (step-by-step)
   - Open decisions or constraints
3. Wait for explicit user confirmation on the plan before writing any code.

Never start coding without a confirmed plan doc and explicit user sign-off.

### 3. Spec before code
Before writing any component, describe:
- What the component does
- Its inputs and outputs
- Edge cases and constraints

### 4. Read before edit
Always read existing files before modifying them.
Never assume what a file contains.

### 5. Incremental slices
Build one thin vertical slice at a time.
Each slice should be independently testable and runnable.
Do not scaffold the entire system upfront.

### 6. Tests are proof
"Seems right" is not enough.
Every feature must have a passing test before it is considered done.
Follow red → green → refactor. 

### 7. Small diffs
Keep each change to ~100 lines or less.
Large diffs are hard to review and hide bugs.
Prefer multiple small commits over one large one.

### 8. No gold-plating
Do not add features, abstractions, or error handling beyond what is asked.
Three similar lines of code is better than a premature abstraction.
Build for the current requirement, not imagined future ones.

### 9. Verify at boundaries
Validate all inputs at system entry points (HTTP handlers, CLI args, queue consumers).
Trust internal code — do not re-validate inside service or repository layers.

### 10. Atomic commits
Each commit should represent one logical change.
Commit message format: `type(scope): short description`
Types: feat, fix, refactor, test, docs, chore

### 11. Comments
Add a single-line comment on top of every method/function describing what it does.
Keep it to one sentence — what it does, not how.
Example: `// CreateUser validates input and persists a new user record to the database.`

Add short inline comments inside functions where the logic is not immediately obvious:
- Why a particular approach was chosen
- What a block of logic is doing at a high level
- Any non-obvious constraint or edge case being handled

Keep inline comments brief (one line). Do not comment every line — only where it adds clarity.

### 13. Context propagation
Every function that does I/O (DB query, HTTP call, cache read/write) must accept `context.Context` as its first argument.
- In Go: name it `ctx` → `func (r *repo) GetByID(ctx context.Context, id string) (*Model, error)`
- In Gin handlers: the Gin context `c *gin.Context` carries this — pass `c.Request.Context()` when calling service/repo.
- Never write a repo interface method without `ctx` — it cannot be cancelled, traced, or timed out.

### 14. Post-iteration summary
After every code change prompted by the user, provide:
1. **What changed** — files and functions modified, and why
2. **Code walkthrough** — walk through the full implementation end-to-end: entry point → handler → service → repo, explaining what each layer does and how they connect
3. **How to run** — exact commands to run the server/app and execute all test cases

Example format:
```
## Walkthrough
- `cmd/server/main.go` — starts the HTTP server, wires dependencies
- `internal/handler/user.go` — receives request, validates input, calls service
- `internal/service/user.go` — applies business logic, calls repo
- `internal/repo/user.go` — reads/writes from in-memory store

## Run
cd <project-name>
go run cmd/server/main.go

## Tests
cd <project-name>
go test ./...                    # summary output
go test ./... -count=1 -v        # detailed per-test output (bypasses cache)
```

**Important:** This repo is a mono-repo. Each project has its own `go.mod`. Always prefix run/test commands with `cd <project-name>`. Never instruct the user to run Go commands from the repo root.

## Quick Commands

- "Implement X" → Read spec rules first, write tests first, then implement
- "Add an endpoint" → Follow api-design.md, validate at handler, logic in service
- "Fix a bug" → Read the file, understand root cause, fix minimally
- "Review this" → Check against all rule files before responding
