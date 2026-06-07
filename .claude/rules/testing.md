# Testing Rules

## TDD Workflow
- Write the test first. Watch it fail (red).
- Write the minimum code to make it pass (green).
- Refactor while keeping tests green.
- Never write implementation before a failing test exists.

## Test Types Required
- Write **Unit Tests (UT)** for all service and repository logic — fast, isolated, no I/O.
- Write **slit tests** for all API endpoints — every endpoint is treated as P0, so slit coverage is mandatory.
- No integration or E2E tests are required beyond slit tests.

## Test Quality
- Test both success and failure paths — every PR must have both.
- Use table-driven tests for multiple input scenarios.
- Test names must describe behavior: `TestMethodName_ErrorType_ReturnsError`
- DAMP over DRY in tests — readable duplication beats abstraction.

## Mocking
- Use mock libraries (gomock, testify/mock) — no hand-rolled stubs.
- Mock at interface boundaries only.
- Never put mock types in the same file as the interface definition.
- Mocks go in a dedicated `mocks/` or `mock/` directory.

## What NOT to do
- Do not test implementation details — test behavior.
- Do not add config overrides in environment-specific config files for tests.
- Do not use `time.Sleep` in tests — use channels or explicit synchronization.
