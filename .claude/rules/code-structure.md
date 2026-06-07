# Code Structure Rules

## Directory Structure

Use the following Go-idiomatic layout as a baseline. Adapt as needed per project size and complexity.

**Project root folder name must match the actual project name** . Never use placeholder names like `my-go-app`.

```
<project-name>/
├── cmd/
│   └── server/
│       └── main.go       # Application entry point
├── internal/             # Private application code
│   ├── errors/
│   │   └── errors.go     # Domain-specific, custom error definitions
│   ├── handler/
│   │   └── user.go       # HTTP/gRPC transport layer
│   ├── repo/
│   │   └── user.go       # Database/storage abstraction
│   └── service/
│       └── user.go       # Core business logic
├── pkg/                  # Public reusable code (optional)
│   └── validator/
│       └── validator.go  # Safe to share with external projects
├── utils/                # Generic helpers (pure functions only)
│   └── string_util.go
├── go.mod
└── go.sum
```

- Use `internal/` for all application-specific code — the Go compiler prevents external imports.
- `pkg/` is optional; only use it for code genuinely reusable across projects.
- `utils/` should contain only pure, stateless helper functions (no I/O, no dependencies).

## Layering
- **Handler/Server layer**: only unmarshal input → call service → marshal response. No logic.
- **Service layer**: all business logic lives here. No DB calls directly.
- **Repository layer**: all data access. No business logic.
- Never skip a layer or mix concerns across layers.

## Repository Structure
- Create one repository interface per DB table — do not combine multiple tables into a single repo.
- Each repository interface is defined alongside its mock in a `mocks/` directory.
- In unit tests, inject the mock repository instead of the real one — never hit the DB in unit tests.
- Example: `users` table → `UserRepository` interface + `mocks/MockUserRepository`.

## Naming
- Functions should be named for what they do, not how they do it.
- No stuttering: `user.UserID` → `user.ID`
- Use constants for all string literals — no magic strings inline.
- Interface names for single-method interfaces: verb + "er" (e.g., `Reader`, `Validator`).

## Functions
- Each function does one thing. If you need "and" to describe it, split it.
- Ideal length: under 50 lines. Hard limit: 100 lines.
- Pass only what a function needs — not entire structs when one field suffices.
- **Every function that performs I/O (DB, HTTP, cache) must accept `ctx context.Context` as its first argument** — named `ctx` in Go, `c` in frameworks like Gin. This enables request cancellation, timeouts, and tracing.
- Repository interface methods must always have `ctx` as the first parameter: `GetByID(ctx context.Context, id string) (*Model, error)`

## Clean Code
- Use switch statements over repeated if-else for multi-case dispatch.
- Remove all unused imports and variables before committing.
- Never leave dead code — delete it, don't comment it out.
- Define methods on the entity they belong to, not as standalone functions elsewhere.
- Use the builder pattern for structs with more than 4 fields.

## Abstractions
- Do not create helpers or utilities for one-time operations.
- Do not design for hypothetical future requirements.
- Minimum complexity for the current task only.
