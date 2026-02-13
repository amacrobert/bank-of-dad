# Implementation Plan: SQLite to PostgreSQL Migration

**Branch**: `011-sqlite-to-postgres` | **Date**: 2026-02-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/011-sqlite-to-postgres/spec.md`
**Further context**: `./sqlite-to-postgres-migration-abstract.md`

## Summary

Hard cutover from SQLite (`modernc.org/sqlite`) to PostgreSQL 17 (`jackc/pgx/v5`) with `golang-migrate` for schema management. The migration touches every file in `backend/internal/store/`, collapses the read/write connection split to a single `*sql.DB`, rewrites all SQL queries for Postgres syntax, replaces per-package test helpers with a shared Postgres-backed helper, and updates Docker Compose + CI to include a Postgres service.

## Technical Context

**Language/Version**: Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend — unchanged)
**Primary Dependencies**: `jackc/pgx/v5` + `pgx/v5/stdlib` (new), `golang-migrate/migrate/v4` (new), remove `modernc.org/sqlite`
**Storage**: PostgreSQL 17 (replacing SQLite)
**Testing**: `go test` + `testify`, Postgres service container in CI
**Target Platform**: Linux/macOS (Docker), GitHub Actions (CI)
**Project Type**: Web application (backend + frontend)
**Performance Goals**: No regression from current behavior; test suite < 60s
**Constraints**: No SQLite code remaining; no ORM; no store interfaces; financial values use exact integer types
**Scale/Scope**: ~10 store files, ~6 test files, docker-compose, CI workflow, backend Dockerfile

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | All existing tests will be migrated to Postgres; shared test helper ensures isolation. No new features = no new TDD cycles required, but test infrastructure itself will be verified by running the full suite. |
| II. Security-First Design | PASS | No security surface changes. `DATABASE_URL` follows existing env-var pattern. Credentials in docker-compose are local dev only. |
| III. Simplicity | PASS | Removing the read/write split reduces complexity. `golang-migrate` is a minimal, well-established migration tool. Single `*sql.DB` is simpler than dual connections. No new abstractions introduced. |

No violations. No Complexity Tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/011-sqlite-to-postgres/
├── plan.md              # This file
├── research.md          # Phase 0: Design decisions and rationale
├── data-model.md        # Phase 1: PostgreSQL schema
├── quickstart.md        # Phase 1: Developer setup guide
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
backend/
├── main.go                              # Update: store.Open(dsn), single *sql.DB
├── internal/
│   ├── config/
│   │   └── config.go                    # Update: DATABASE_URL replaces DATABASE_PATH
│   ├── store/
│   │   ├── postgres.go                  # New: replaces sqlite.go (connection + Open())
│   │   ├── child.go                     # Update: Postgres SQL syntax
│   │   ├── parent.go                    # Update: Postgres SQL syntax
│   │   ├── family.go                    # Update: Postgres SQL syntax
│   │   ├── session.go                   # Update: Postgres SQL syntax, time.Time scanning
│   │   ├── transaction.go              # Update: RETURNING instead of LastInsertId
│   │   ├── interest.go                  # Update: strftime → EXTRACT/to_char
│   │   ├── interest_schedule.go         # Update: Postgres SQL syntax
│   │   ├── schedule.go                  # Update: Postgres SQL syntax, time.Time scanning
│   │   ├── auth_event.go               # Update: Postgres SQL syntax
│   │   ├── sqlite.go                    # DELETE
│   │   ├── sqlite_test.go              # DELETE (replaced by postgres_test.go)
│   │   ├── postgres_test.go            # New: connection/migration tests
│   │   ├── session_test.go             # Update: use shared test helper
│   │   └── auth_event_test.go          # Update: use shared test helper
│   ├── testutil/
│   │   └── db.go                        # New: shared test helper
│   ├── allowance/
│   │   ├── scheduler.go                 # Update: single *sql.DB
│   │   └── handler_test.go              # Update: use shared test helper
│   ├── interest/
│   │   ├── scheduler.go                 # Update: single *sql.DB
│   │   └── handler_test.go              # Update: use shared test helper
│   └── balance/
│       └── handler_test.go              # Update: use shared test helper
├── migrations/
│   ├── 001_initial_schema.up.sql        # New: Postgres DDL
│   └── 001_initial_schema.down.sql      # New: Postgres DDL rollback
├── Dockerfile                           # Update: remove SQLite data dir
├── go.mod                               # Update: swap dependencies
└── go.sum                               # Update: regenerated

docker-compose.yaml                      # Update: add postgres service
.github/workflows/ci.yml                # Update: add postgres service container
.env.example                             # New: DATABASE_URL default
```

**Structure Decision**: Existing web application structure preserved. Only backend files change. Frontend is untouched.

## Phases Overview

### Phase 1: Driver and connection layer
Replace `modernc.org/sqlite` with `jackc/pgx/v5`, collapse `DB{Read, Write}` to single `*sql.DB`, update config to read `DATABASE_URL`.

### Phase 2: Migration framework
Add `golang-migrate/migrate/v4`, create `backend/migrations/001_initial_schema.{up,down}.sql` with Postgres DDL, integrate `migrate.Up()` into `store.Open()`.

### Phase 3: SQL query rewrite
Convert all store files: `?` → `$N` placeholders, `parseTime()` → direct `time.Time` scanning, `LastInsertId()` → `RETURNING`, `strftime()` → `EXTRACT()`/`to_char()`, boolean handling.

### Phase 4: Test infrastructure
Create shared test helper in `backend/internal/testutil/db.go`, replace all per-package helpers, remove SQLite temp file cleanup.

### Phase 5: Docker and CI
Add Postgres service to `docker-compose.yaml`, add Postgres service container to GitHub Actions, update backend Dockerfile, create `.env.example`.

### Phase 6: Cleanup and verification
Remove all SQLite references, `go mod tidy`, verify build/test/lint/vet.
