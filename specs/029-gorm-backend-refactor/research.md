# Research: GORM Backend Refactor

**Date**: 2026-03-16 | **Feature**: 029-gorm-backend-refactor

## R1: GORM + pgx/v5 Compatibility

**Decision**: Use `gorm.io/driver/postgres` which supports pgx as the underlying driver. GORM wraps `*sql.DB` internally, so the existing pgx/v5/stdlib registration continues to work.

**Rationale**: GORM's postgres driver uses `database/sql` under the hood. Since `pgx/v5/stdlib` registers itself as a `database/sql` driver, GORM can use the same DSN and connection pool. No driver conflict.

**Alternatives considered**:
- Pure pgx without stdlib adapter — rejected because GORM requires `database/sql` interface
- Switching to lib/pq — rejected, pgx is more performant and already in use

## R2: GORM Model Design for Nullable Fields

**Decision**: Use pointer types (`*string`, `*int64`, `*time.Time`) in GORM model structs for nullable columns. GORM handles pointer nil ↔ SQL NULL mapping automatically.

**Rationale**: The current store uses `sql.NullString`/`sql.NullInt64`/`sql.NullTime` for scanning, then manually converts to pointer types in the domain struct. GORM eliminates this two-step process — pointer fields are scanned directly. This reduces boilerplate significantly.

**Alternatives considered**:
- Keep `sql.Null*` types in GORM models — works but defeats the purpose of GORM's auto-mapping and adds unnecessary verbosity
- Use zero values for nullable — rejected because zero and NULL have different semantics (e.g., empty string vs no avatar)

## R3: Money Representation in GORM

**Decision**: Continue using `int64` for all money fields (balance_cents, amount_cents, target_amount_cents, current_amount_cents). GORM maps `int64` to PostgreSQL `BIGINT` directly.

**Rationale**: No change needed. GORM handles int64 ↔ BIGINT natively. The field names in models will use Go conventions (e.g., `BalanceCents int64`) with GORM column tags mapping to snake_case DB columns.

**Alternatives considered**:
- Custom Money type with GORM Scanner/Valuer interface — over-engineering for this use case
- `decimal` package — rejected per existing project convention of int64 cents

## R4: Migration Strategy (golang-migrate Preservation)

**Decision**: GORM's AutoMigrate is explicitly forbidden. Schema management stays with `golang-migrate`. GORM models must match the schema defined in `backend/migrations/`. The `Open()` function in the new `repositories/db.go` will still run migrations via `golang-migrate` before returning the `*gorm.DB` instance.

**Rationale**: AutoMigrate can't handle complex migrations (data transformations, index changes, column renames). The existing migration workflow is battle-tested. GORM is used only for runtime query building.

**Alternatives considered**:
- GORM AutoMigrate for dev, golang-migrate for prod — rejected because schema drift between environments causes bugs
- Atlas or goose — rejected, golang-migrate is already in use and works well

## R5: GORM Transaction Support

**Decision**: Use GORM's `db.Transaction(func(tx *gorm.DB) error)` closure pattern for atomic operations. Repository methods that need atomicity will accept a `*gorm.DB` (which could be a transaction).

**Rationale**: GORM's transaction closure automatically handles commit/rollback. This is cleaner than the current `tx, err := db.Begin(); defer tx.Rollback()` pattern. For operations spanning multiple repositories, the caller passes a transaction `*gorm.DB` to each repository method.

**Alternatives considered**:
- Unit of Work pattern — over-engineering for this project's complexity level
- Transaction manager service — adds unnecessary abstraction layer

## R6: Complex Query Handling

**Decision**: Use GORM's query builder for standard CRUD. Use `db.Raw()` and `db.Exec()` for complex queries (e.g., monthly transaction summaries, interest calculations with date logic, slug suggestion queries).

**Rationale**: Not all current SQL queries translate cleanly to GORM's builder (e.g., `to_char()` aggregations, `COALESCE`, complex `GROUP BY`). Forcing these into GORM's API would produce less readable code than the raw SQL. GORM's `Raw()` still provides parameter binding safety.

**Alternatives considered**:
- Force everything through GORM builder — rejected for readability
- Keep a separate raw SQL layer — rejected for consistency; `db.Raw()` within repositories keeps all data access in one place

## R7: Error Handling Translation

**Decision**: Map GORM error patterns to match existing store behavior:
- `gorm.ErrRecordNotFound` → return `nil, nil` (not found is not an error)
- `gorm.ErrDuplicatedKey` or unique constraint errors → wrap with user-friendly message
- All other errors → wrap with context using `fmt.Errorf`

**Rationale**: Existing handlers expect nil return for not-found and specific error messages for duplicates. GORM uses sentinel errors, so translation is straightforward.

**Alternatives considered**:
- Custom error types — over-engineering; string wrapping matches current pattern
- Return GORM errors directly — breaks handler expectations

## R8: Test Migration Strategy

**Decision**: Create `repositories/test_helpers_test.go` with a `testDB()` function that returns `*gorm.DB` connected to `bankofdad_test`. Each test file mirrors the existing store test patterns but uses GORM. TRUNCATE statements use `db.Exec()` with `RESTART IDENTITY CASCADE`.

**Rationale**: Tests must use the real database (per constitution: no mocks for data layer). The `testDB()` helper provides a consistent connection. Tests run with `-p 1` due to shared test DB.

**Alternatives considered**:
- SQLite for test speed — rejected, PostgreSQL-specific features (SERIAL, TIMESTAMPTZ, to_char) would break
- Per-test database — too slow and complex for this project scale

## R9: Handler Migration Approach

**Decision**: Migrate handlers one package at a time. Each handler's constructor changes from accepting `*store.XxxStore` to `*repositories.XxxRepo`. Method names and signatures on the repository should match the store methods to minimize handler changes.

**Rationale**: Keeping method names consistent (e.g., `GetByID`, `Create`, `ListByFamily`) means handler code changes are primarily import path and type name updates. This reduces risk and makes the migration mechanical.

**Alternatives considered**:
- Interface-based approach (handlers depend on interfaces) — adds abstraction layer not requested by user; could be done later if needed
- Big-bang migration — too risky; incremental per-handler is safer

## R10: Package Placement (Top-Level vs Internal)

**Decision**: Place `models/` and `repositories/` at `backend/models/` and `backend/repositories/` (top-level), not under `backend/internal/`.

**Rationale**: User explicitly requested these paths. Top-level placement also avoids import cycle issues — `internal/` packages importing each other can create cycles, while top-level packages are imported by `internal/` packages one-directionally. Models and repositories are foundational packages that many other packages depend on.

**Alternatives considered**:
- `backend/internal/models/` and `backend/internal/repositories/` — would work but doesn't match user request and could create import cycle challenges
