# Security Rules

## Credentials & Secrets
- Never hardcode credentials, API keys, tokens, or passwords in source code.
- Never commit secrets to git — use environment variables or a secrets manager.
- Rotate any secret that was accidentally committed, even briefly.

## Input Validation (OWASP Top 10)
- Sanitize and validate all user-supplied input before use.
- Prevent SQL injection: use parameterized queries or an ORM — never string-concatenate SQL.
- Prevent XSS: escape output in templates; use Content-Security-Policy headers.
- Prevent command injection: never pass user input to shell commands.

## Authentication & Authorization
- Always verify identity before serving protected resources.
- Check authorization on every request — do not rely on client-side enforcement.
- Use short-lived tokens; implement token expiry and refresh.
- Never store plain-text passwords — always hash with bcrypt or argon2.

## Data
- Encrypt sensitive data at rest (PII, financial data).
- Use HTTPS/TLS for all data in transit.
- Apply the principle of least privilege — only request and store data you need.

## Dependencies
- Justify every new dependency added.
- Pin dependency versions.
- Audit for known vulnerabilities before adding.

## Debug Leftovers
- Remove all debug logs, hardcoded test values, and commented-out code before merging.
- Never log sensitive data (passwords, tokens, card numbers).
