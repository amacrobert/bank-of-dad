# Research: SQLite to PostgreSQL Migration

## R-001: Money Column Type — BIGINT vs NUMERIC

**Decision**: Use `BIGINT` for all `*_cents` columns (`amount_cents`, `balance_cents`).

**Rationale**: The Go code uses `int64` throughout for money, representing amounts in cents. `BIGINT` maps directly to Go's `int64` with zero conversion overhead. It provides exact integer precision — there is no floating-point risk because the values are integers. Using `NUMERIC`/`DECIMAL` would require changes to the Go scanning layer (different driver type mapping) for no actual precision benefit, since the values are always whole cents.

**Alternatives considered**:
- `NUMERIC(19,0)` — Adds scanning complexity. pgx would return `pgtype.Numeric` instead of `int64`, requiring explicit conversion. No precision benefit for integer cents.
- `INTEGER` — Sufficient range for this app, but `BIGINT` matches Go's `int64` and is future-proof.

## R-002: Boolean Columns — Postgres BOOLEAN vs INTEGER

**Decision**: Use `BOOLEAN` for `is_locked`. Map Go's `int64`/`int` comparisons (`== 0`, `== 1`) to native boolean scanning.

**Rationale**: Postgres has a native `BOOLEAN` type. The current SQLite schema uses `INTEGER NOT NULL DEFAULT 0` for `is_locked`. In Go, the `Child` struct already stores this as an `int` or is compared against 0/1. We should update the Go struct to use `bool` and scan directly, or keep `int` and let pgx handle the conversion (pgx can scan Postgres `BOOLEAN` into Go `bool`). The cleaner approach is to update the Go struct field to `bool`.

**Alternatives considered**:
- Keep `INTEGER` in Postgres — Works but isn't idiomatic Postgres. Misses the opportunity to use native type safety.

## R-003: Timestamp Handling — Direct time.Time vs String Parsing

**Decision**: Remove `parseTime()` entirely. Scan timestamps directly into `time.Time` (or `sql.NullTime` for nullable columns). Pass `time.Time` values directly to queries instead of formatting as strings.

**Rationale**: SQLite stores timestamps as text, requiring the `parseTime()` helper that tries multiple formats. pgx with `TIMESTAMPTZ` columns handles `time.Time` natively — no string formatting or parsing needed. This eliminates an entire class of bugs (format mismatches) and simplifies the code.

**Impact**: Every store file that calls `parseTime()` or formats `time.Time` as a string needs updating. All `var createdAt string` + `parseTime(createdAt)` patterns become `var createdAt time.Time` with direct scanning.

## R-004: Auto-increment IDs — SERIAL vs BIGSERIAL

**Decision**: Use `SERIAL` (4-byte) for all auto-increment primary keys.

**Rationale**: The Go code uses `int64` for IDs, which can hold either `SERIAL` (int32 max ~2B) or `BIGSERIAL` (int64) values. For a family banking app, `SERIAL` provides more than enough range. `BIGSERIAL` is unnecessary overhead.

**Alternatives considered**:
- `BIGSERIAL` — Overkill for this app's scale. Go's `int64` works with `SERIAL` (values fit).
- `UUID` — Not used anywhere in the current codebase. Would require changing all ID types.

## R-005: LastInsertId() → RETURNING

**Decision**: Replace all `result.LastInsertId()` calls with `INSERT ... RETURNING id` and `QueryRow().Scan()`.

**Rationale**: pgx's stdlib adapter does not support `LastInsertId()` (Postgres limitation). The idiomatic Postgres approach is `INSERT ... RETURNING id`, which is also more efficient (single round trip).

**Impact**: `transaction.go` (Deposit, Withdraw, DepositAllowance) all use `tx.Exec()` + `result.LastInsertId()`. These become `tx.QueryRow(...).Scan(&id)` with `RETURNING id` appended to the INSERT.

## R-006: strftime() in Interest Calculation

**Decision**: Replace `strftime('%Y-%m', last_interest_at) != strftime('%Y-%m', 'now')` with `EXTRACT(YEAR FROM last_interest_at) != EXTRACT(YEAR FROM NOW()) OR EXTRACT(MONTH FROM last_interest_at) != EXTRACT(MONTH FROM NOW())`, or equivalently `to_char(last_interest_at, 'YYYY-MM') != to_char(NOW(), 'YYYY-MM')`.

**Rationale**: `strftime()` is SQLite-specific. Postgres provides `EXTRACT()` and `to_char()`. The `to_char()` approach is closest to the original intent (string comparison of year-month).

## R-007: Connection Architecture — Single Pool

**Decision**: Replace `DB{Write *sql.DB, Read *sql.DB}` with a single `*sql.DB`.

**Rationale**: The read/write split exists solely because SQLite only allows one writer at a time. Postgres handles concurrent reads and writes natively through MVCC. A single connection pool with configurable `MaxOpenConns` is simpler and idiomatic.

**Impact**: The `DB` struct is removed. All store structs change from `db *DB` to `db *sql.DB`. All `s.db.Write.Exec()` and `s.db.Read.Query()` become `s.db.Exec()` and `s.db.Query()`. Every store constructor signature changes from `NewXxxStore(db *DB)` to `NewXxxStore(db *sql.DB)`.

## R-008: Migration Framework — golang-migrate

**Decision**: Use `golang-migrate/migrate/v4` with file-based migration source and Postgres driver.

**Rationale**: Lightweight, well-maintained, and purpose-built for exactly this use case. No ORM overhead. Migrations are plain SQL files. Integrates easily into `store.Open()`. Handles `ErrNoChange` gracefully.

**Dependencies added**:
- `github.com/golang-migrate/migrate/v4`
- `github.com/golang-migrate/migrate/v4/database/postgres`
- `github.com/golang-migrate/migrate/v4/source/file`

## R-009: Test Isolation Strategy

**Decision**: Shared test helper in `backend/internal/testutil/db.go` that connects to a test Postgres database and truncates all tables in `t.Cleanup()`.

**Rationale**: Truncation is simpler than transaction-based isolation (which would interfere with tests that explicitly test transaction behavior). All test packages currently have duplicated setup helpers — centralizing eliminates ~6 copies of nearly identical code.

**Test database**: `bankofdad_test` database, connected via `TEST_DATABASE_URL` env var (defaults to `postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable`).

## R-010: Postgres Version

**Decision**: PostgreSQL 17 (latest stable).

**Rationale**: The migration abstract suggests 18.2, but PostgreSQL 18 is not yet released as of February 2026. PostgreSQL 17 is the current latest stable release. Using an unreleased version would break Docker image pulls.

## R-011: Placeholder Conversion Strategy

**Decision**: Manual replacement of `?` with `$1, $2, $3, ...` in each query, with careful attention to parameter ordering.

**Rationale**: No automated tool reliably handles this conversion across multi-line SQL strings in Go. The codebase has ~50 queries total across 10 files — manual conversion is tractable and safer.

**Methodology**: For each query, count parameters left-to-right and replace `?` with `$N` sequentially. Double-check that the Go arguments in `Query()`/`Exec()` calls match the new numbering.
