# Database Rules

> **Note:** All projects in this repo use in-memory storage only (see `CLAUDE.md` → Project Constraints). The rules below apply if a project ever graduates to a real DB, but do not introduce external storage unless explicitly requested.


## Schema
- Store all timestamps in UTC.
- Use `created_at` and `updated_at` on every table — let the ORM/framework manage `updated_at`.
- Do not set `updated_at` manually in application code.
- Prefer soft deletes (`deleted_at`) over hard deletes for auditable entities.
- Do not filter by `deleted_at IS NULL` manually — use ORM soft-delete hooks.

## Indexes
- Analyze query patterns before adding a new query.
- Every foreign key should have an index.
- Every column used in a WHERE, ORDER BY, or JOIN should be evaluated for indexing.
- Composite indexes: column order matters — put the most selective column first.
- Document why each index exists in a comment on the migration.

## Migrations
- Migrations must be reversible (have an `up` and `down`).
- Never modify an existing migration that has already been applied — create a new one.
- Migrations must be backward-compatible with the current running version of the app.
- Avoid long-running migrations on large tables without a plan (batching, background job).

## Queries
- **Before writing any query, ask and clarify the query type**: is it a point lookup, a list/filter, an aggregation, or a write? The query type determines the index strategy and repo method signature.
- Avoid N+1 queries — use joins or batch fetching.
- Never use `SELECT *` in production queries — list columns explicitly.
- Use pagination for all list endpoints — never return unbounded result sets.
- Prefer reads from replicas for non-critical reads; writes always go to primary.

## Transactions
- Use transactions for operations that must succeed or fail together.
- Keep transactions short — do not perform external API calls inside a transaction.
- Always handle rollback on error.
