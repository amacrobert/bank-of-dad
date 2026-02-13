# SQLite to Postgres Migration — Claude Code Prompt

## Context

This is a side project called "Bank of Dad" — a React/TypeScript frontend with a Go backend. The backend provides an API and async workloads. The data layer currently uses SQLite (`modernc.org/sqlite` v1.44.3) with raw SQL via `database/sql`. We are doing a hard cutover to PostgreSQL 18.2 — no SQLite support should remain after this migration.

## Project Structure

- All SQL queries live in `backend/internal/store/`
- Each domain area has its own concrete store struct holding a `db *sql.DB` pointer, constructed via `NewXxxStore(db *sql.DB)`
- `main.go` calls `store.Open()` once and passes the resulting `*sql.DB` into each store constructor
- Two background schedulers also use the store: `backend/internal/allowance/scheduler.go` and `backend/internal/interest/scheduler.go`
- Tests throughout `backend/` open and close their own database connections
- The project uses `docker-compose.yaml` (existing) and has a GitHub Actions CI pipeline for linting and tests

## Migration Plan — Execute in This Order

### Phase 1: Replace the SQL driver and connection layer

1. **Remove** the `modernc.org/sqlite` dependency. **Add** `jackc/pgx/v5` and its `stdlib` adapter (`jackc/pgx/v5/stdlib`).
2. **Rewrite `store.Open()`**:
   - It currently takes a file path, opens two `*sql.DB` connections (read + write) as a SQLite concurrency workaround, and runs migrations.
   - Replace it: accept a PostgreSQL DSN string (e.g. `postgres://user:pass@localhost:5432/bankofdad?sslmode=disable`), open a **single** `*sql.DB` using `pgx` via stdlib, and return it.
   - The read/write split is no longer needed. Collapse to a single `*sql.DB` everywhere. Update all store structs, constructors, and callers (`main.go`, schedulers, tests) to use the single connection.
   - Remove all references to the old read/write connection pattern, including any `ReadDB`/`WriteDB` fields or similar.
3. **Read the DSN from the environment**: the backend should read `DATABASE_URL` from its `.env` file. Add `DATABASE_URL` to `.env.example` with a sensible default like `postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable`.

### Phase 2: Set up migrations with golang-migrate

1. **Add** `golang-migrate/migrate/v4` as a dependency, with the `postgres` and `file` source drivers.
2. **Create** a `backend/migrations/` directory.
3. **Infer the current needed schema** by reading what the store layer actually queries — the tables, columns, types, and constraints that the Go code depends on. Do NOT replay the existing SQLite migrations 1:1. Legacy tables/columns that are not referenced by any current Go code should be omitted.
4. **Write** `backend/migrations/001_initial_schema.up.sql` and `backend/migrations/001_initial_schema.down.sql` with the inferred Postgres schema. Use idiomatic Postgres types:
   - `TEXT` for strings
   - `BOOLEAN` for booleans (SQLite uses integers for these)
   - `TIMESTAMPTZ` for timestamps
   - `BIGINT` or `INTEGER` as appropriate for IDs and numeric fields
   - `NUMERIC` or `DECIMAL` for money/financial values (this is a banking app for kids)
   - `UUID` if any IDs are UUIDs
   - Use `SERIAL` or `BIGSERIAL` for auto-increment primary keys if applicable
   - Add appropriate `NOT NULL`, `DEFAULT`, foreign key, and uniqueness constraints based on what the code expects
5. **Integrate migration running** into `store.Open()` — after opening the connection, run `migrate.Up()` automatically. Handle the `migrate.ErrNoChange` case gracefully (it's not an error).
6. **Delete** all old SQLite migration files and the old `db.migrate()` function.

### Phase 3: Rewrite all SQL queries for Postgres

Go through every `.go` file in `backend/internal/store/` and rewrite all SQL queries:

1. **Replace `?` placeholders with `$1, $2, $3, ...`** — Postgres uses numbered placeholders.
2. **Replace SQLite idioms**:
   - `IFNULL()` → `COALESCE()`
   - `datetime('now')` → `NOW()`
   - `strftime(...)` → Postgres `to_char()` or `EXTRACT()`
   - `INTEGER` used as boolean → actual `BOOLEAN` columns, compare with `true`/`false` not `1`/`0`
   - `GLOB` → Postgres `LIKE` or `~` (regex)
   - `GROUP_CONCAT` → `STRING_AGG`
   - String concatenation with `||` is fine in both, no change needed
   - `AUTOINCREMENT` → handled by `SERIAL`/`BIGSERIAL` in schema
   - `INSERT OR REPLACE` / `INSERT OR IGNORE` → `INSERT ... ON CONFLICT ...`
   - `LIKE` is case-insensitive in SQLite but case-sensitive in Postgres — use `ILIKE` if case-insensitive matching is intended
3. **Review `RETURNING` clauses** — both support it, but verify syntax.
4. **Review transaction usage** — the `database/sql` `tx.Begin()` / `tx.Commit()` pattern works identically with pgx/stdlib, but verify no SQLite-specific transaction modes are set (e.g., `BEGIN IMMEDIATE`). Replace any with plain `BEGIN` (or just use `db.BeginTx()`).
5. **Review time handling** — if any code stores/retrieves times as strings (a SQLite pattern), update to use `time.Time` directly, which pgx handles natively with `TIMESTAMPTZ`.

### Phase 4: Update tests

1. **Create a shared test helper** in `backend/internal/testutil/db.go` (or similar) that:
   - Reads `DATABASE_URL` from the environment (with a sensible default for local dev like `postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable`)
   - Opens a Postgres connection
   - Runs migrations
   - **Provides test isolation**: before each test, start a transaction or truncate all tables, so tests don't leak state. A simple approach: `TRUNCATE table1, table2, ... CASCADE` in a cleanup function registered with `t.Cleanup()`.
   - Returns the `*sql.DB` for use in constructing stores
2. **Replace all per-package test helpers** — every package currently has its own copy of a helper that creates a temp SQLite file, calls `store.Open()`, and registers cleanup. Replace all of them with calls to the shared helper.
3. **Remove all SQLite temp file cleanup code** — no more `os.Remove`, WAL/SHM sidecar cleanup, `os.CreateTemp`, `t.TempDir()` database patterns.
4. **Verify all tests pass** with `go test ./...`.

### Phase 5: Docker and infrastructure

1. **Add a Postgres 18.2 service** to `docker-compose.yaml`:
   ```yaml
   postgres:
     image: postgres:18.2
     environment:
       POSTGRES_USER: bankofdad
       POSTGRES_PASSWORD: bankofdad
       POSTGRES_DB: bankofdad
     ports:
       - "5432:5432"
     volumes:
       - postgres_data:/var/lib/postgresql/data
     healthcheck:
       test: ["CMD-ONLY", "pg_isready", "-U", "bankofdad"]
       interval: 5s
       timeout: 5s
       retries: 5
   ```
2. **Add the `postgres_data` named volume** to the `volumes:` section.
3. **Remove** the old SQLite volume from docker-compose.
4. **Make the backend service depend on Postgres** with `depends_on: postgres: condition: service_healthy`.
5. **Add `DATABASE_URL`** to the backend service's environment in docker-compose, pointing to the postgres service hostname.

### Phase 6: Update CI (GitHub Actions)

1. **Add a Postgres service container** to the test job:
   ```yaml
   services:
     postgres:
       image: postgres:18.2
       env:
         POSTGRES_USER: bankofdad
         POSTGRES_PASSWORD: bankofdad
         POSTGRES_DB: bankofdad_test
       ports:
         - 5432:5432
       options: >-
         --health-cmd pg_isready
         --health-interval 10s
         --health-timeout 5s
         --health-retries 5
   ```
2. **Set `DATABASE_URL`** as an environment variable for the test step: `postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable`
3. **Remove** any SQLite-related CI setup if it exists.

### Phase 7: Cleanup

1. **Run `go mod tidy`** to remove the SQLite dependency and add pgx/migrate.
2. **Search the entire `backend/` directory** for any remaining references to `sqlite`, `SQLite`, `modernc`, the old database file path, or the read/write connection pattern. Remove them all.
3. **Verify**:
   - `docker-compose up` starts Postgres and the backend connects successfully
   - `go build ./...` compiles cleanly
   - `go test ./...` passes with a running Postgres instance
   - `go vet ./...` and any linters pass

## Important Constraints

- **No SQLite code should remain.** This is a hard cutover. Do not add a "driver toggle" or keep any SQLite compatibility.
- **No ORM.** Continue using raw SQL with `database/sql` and `pgx/stdlib`.
- **No interfaces for stores.** Keep the concrete store struct pattern — handlers take concrete types directly.
- **Infer schema from code, not from old migrations.** The old migrations contain legacy features that are no longer needed. Read what the store layer actually uses.
- **Financial values must use `NUMERIC`/`DECIMAL`, not floating point.** This is a banking app.
- Before making changes, read the relevant files to understand the current implementation. Start with `backend/internal/store/` to understand the schema and query patterns, then branch out.
