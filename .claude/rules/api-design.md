# API Design Rules

## Contract First
- Define the API contract (request/response shape, status codes, error format) before writing any implementation.
- Document every endpoint: method, path, auth requirement, request body, response body, error codes.

## Validation
- Validate all inputs at the handler/entry layer — not deep inside service code.
- Reject early: return a 400 immediately for malformed or missing required fields.
- Validate types, ranges, and formats at the boundary.
- Do not re-validate the same input at multiple layers.

## Response Design
- Use consistent response envelopes: `{ "data": {}, "error": null }` or similar.
- Always include an error code (machine-readable) alongside an error message (human-readable).
- Return the correct HTTP status code — never return 200 with an error body.

## Versioning
- Version APIs from the start: `/v1/some-resource`
- Never break existing contracts — add new fields, never remove or rename existing ones.
- Follow Hyrum's Law: once an API is used, everything about it becomes a contract.

## Idempotency
- Write operations (POST, DELETE) should be idempotent where possible.
- Use idempotency keys for operations that must not be executed twice (payments, emails).

## Naming Conventions
- Use nouns for resources: `/users`, `/orders` — not `/createUser`, `/getOrder`
- Use HTTP verbs to express intent: GET (read), POST (create), PUT/PATCH (update), DELETE (remove)
- Use plural nouns for collections.
