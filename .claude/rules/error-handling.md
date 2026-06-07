# Error Handling Rules

## Logging
- **Log at the exact node where the error first occurs** — the function that directly receives the error from a dependency (DB call, HTTP call, etc.) must log it immediately.
- Use structured logging with key-value pairs: `logger.Error("msg", "key", value, "error", err)`
- Never use `fmt.Errorf` as a substitute for logging — log AND wrap.
- Log failed reasons on business operation failures (e.g., order failed, payment rejected).

## Wrapping
- After logging at the origin, **wrap the error with context at every upstream caller level**: `fmt.Errorf("create user: %w", err)`
- Each wrapping layer adds its own operation name so the error chain reads like a call stack.
- Preserve the original error using `%w` so callers can use `errors.Is` / `errors.As`.
- Example chain: repo logs `"db insert failed"` → service wraps `"create user: %w"` → handler wraps `"handle POST /users: %w"`.

## Propagation
- Never silently swallow errors — if you call a function and ignore the error, add a log.
- If an error is expected and non-fatal, document why it is being ignored.

## External exposure
- Do not expose internal error messages to external callers (HTTP clients, API consumers).
- Map internal errors to appropriate external codes (4xx for client errors, 5xx for server).
- Return generic messages for unexpected internal errors — log the detail internally.

## Domain errors
- Define typed error constants for domain-specific failures.
- Do not use raw strings for error comparison — use sentinel errors or error types.
