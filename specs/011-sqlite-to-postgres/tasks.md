# Tasks: SQLite to PostgreSQL Migration

**Input**: Design documents from `/specs/011-sqlite-to-postgres/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Swap dependencies and update configuration to read PostgreSQL connection string.

- [ ] T001 Add `jackc/pgx/v5`, `pgx/v5/stdlib`, `golang-migrate/migrate/v4` (with `database/postgres` and `source/file` drivers) to `backend/go.mod`; remove `modernc.org/sqlite`; run `go mod tidy`
- [ ] T002 Update `backend/internal/config/config.go`: rename `DatabasePath` field to `DatabaseURL`, read from `DATABASE_URL` env var (default: `postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable`), remove `DATABASE_PATH` references
- [ ] T003 Create `.env.example` in project root with `DATABASE_URL=postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable` and existing env var placeholders (GOOGLE_CLIENT_ID, etc.)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: New connection layer, migration framework, and schema — MUST be complete before any store files can be rewritten.

- [ ] T004 Create `backend/migrations/001_initial_schema.up.sql` with the full PostgreSQL schema from `data-model.md` (7 tables, all indexes, all constraints, FK for schedule_id added via ALTER TABLE after allowance_schedules exists)
- [ ] T005 [P] Create `backend/migrations/001_initial_schema.down.sql` per `data-model.md` (DROP TABLE in reverse dependency order)
- [ ] T006 Create `backend/internal/store/postgres.go`: define `Open(dsn string) (*sql.DB, error)` that opens a connection via `pgx/v5/stdlib`, runs `golang-migrate` `Up()` (handling `ErrNoChange`), and returns a single `*sql.DB`; define `Close(db *sql.DB) error` helper; remove the old `DB` struct entirely
- [ ] T007 Update all store struct types and constructors across all store files: change `db *DB` field to `db *sql.DB`, update `NewXxxStore(db *DB)` signatures to `NewXxxStore(db *sql.DB)`, replace all `s.db.Write.Exec` / `s.db.Write.Begin` with `s.db.Exec` / `s.db.Begin`, replace all `s.db.Read.Query` / `s.db.Read.QueryRow` with `s.db.Query` / `s.db.QueryRow` — files: `child.go`, `parent.go`, `family.go`, `session.go`, `transaction.go`, `interest.go`, `interest_schedule.go`, `schedule.go`, `auth_event.go`
- [ ] T008 Update `backend/main.go`: call `store.Open(cfg.DatabaseURL)` instead of `store.Open(cfg.DatabasePath)`, receive `*sql.DB` instead of `*store.DB`, pass `*sql.DB` to all `NewXxxStore()` constructors, update `defer db.Close()` call

**Checkpoint**: Connection layer compiles against Postgres. Store files still have SQLite SQL syntax — queries will fail at runtime until Phase 3.

---

## Phase 3: User Story 1 — All Application Features Continue Working (Priority: P1)

**Goal**: Rewrite every SQL query in the store layer for PostgreSQL syntax so all existing features produce identical results.

**Independent Test**: Run `go test ./...` against a local Postgres instance — 100% of existing tests pass.

### Core store rewrites

Each task rewrites one store file: convert `?` → `$1,$2,...` placeholders, remove `parseTime()` calls (scan `time.Time` directly), pass `time.Time` values to queries instead of formatting as strings, and apply file-specific changes noted below.

- [ ] T009 [P] [US1] Rewrite `backend/internal/store/family.go` for Postgres: convert placeholders (`?` → `$1`), scan `created_at` directly into `time.Time`
- [ ] T010 [P] [US1] Rewrite `backend/internal/store/parent.go` for Postgres: convert placeholders, scan `created_at` directly into `time.Time`
- [ ] T011 [P] [US1] Rewrite `backend/internal/store/auth_event.go` for Postgres: convert placeholders (7 params in INSERT), scan `created_at` directly into `time.Time`
- [ ] T012 [US1] Rewrite `backend/internal/store/child.go` for Postgres: convert placeholders, change `is_locked` field to `bool` in `Child` struct, scan `BOOLEAN` directly, update `LockAccount` to use `TRUE`/`FALSE` instead of `1`/`0` in queries, scan timestamps directly, add `RETURNING id` to `Create` INSERT if `LastInsertId()` is used
- [ ] T013 [US1] Rewrite `backend/internal/store/session.go` for Postgres: convert placeholders, remove `parseTime()` helper function, scan `created_at`/`expires_at` directly into `time.Time`, pass `time.Time` values to queries instead of `expiresAt.UTC().Format(time.DateTime)`
- [ ] T014 [US1] Rewrite `backend/internal/store/transaction.go` for Postgres: convert placeholders, replace `tx.Exec()` + `result.LastInsertId()` with `tx.QueryRow("INSERT ... RETURNING id").Scan(&id)` in `Deposit`, `Withdraw`, and `DepositAllowance` methods, scan `created_at` directly
- [ ] T015 [US1] Rewrite `backend/internal/store/schedule.go` for Postgres: convert placeholders (up to 9 params), change `nextRunAt sql.NullString` → `sql.NullTime`, scan `created_at`/`updated_at` directly into `time.Time`, pass `time.Time` to queries instead of RFC3339 strings
- [ ] T016 [US1] Rewrite `backend/internal/store/interest.go` for Postgres: convert placeholders, replace `strftime('%Y-%m', c.last_interest_at) != ?` with `to_char(c.last_interest_at, 'YYYY-MM') != $1` in `ListDueForInterest`, pass `time.Time` for `last_interest_at` in `ApplyInterest`, scan timestamps directly
- [ ] T017 [P] [US1] Rewrite `backend/internal/store/interest_schedule.go` for Postgres: convert placeholders, change `nextRunAt sql.NullString` → `sql.NullTime`, scan timestamps directly, pass `time.Time` to queries

### Scheduler and handler updates

- [ ] T018 [P] [US1] Update `backend/internal/allowance/scheduler.go`: change `store.DB` references if any remain, verify store method calls still compile with `*sql.DB`-based stores
- [ ] T019 [P] [US1] Update `backend/internal/interest/scheduler.go`: same as T018 — verify store method calls compile
- [ ] T020 [US1] Update `backend/internal/auth/handlers.go`: if it references `store.DB` or `*store.DB`, update to `*sql.DB`; verify all handler constructors receive correct types

### Cleanup within US1

- [ ] T021 [US1] Delete `backend/internal/store/sqlite.go` (old connection layer, `DB` struct, `migrate()`, `setPragmas()`, `addColumnIfNotExists()`, `migrateTransactionsCheckConstraint()`, `migrateTransactionsInterestType()`, `migrateAllowanceUniqueChild()`)

**Checkpoint**: `go build ./...` compiles. All store operations use Postgres SQL. Run against a local Postgres instance to verify queries execute correctly.

---

## Phase 4: User Story 2 — Reliable Automated Testing (Priority: P2)

**Goal**: Create a shared test helper and migrate all test files so the full test suite runs against Postgres with proper isolation.

**Independent Test**: `TEST_DATABASE_URL=... go test ./...` — all tests pass, no test-to-test state leakage.

### Shared test helper

- [ ] T022 [US2] Create `backend/internal/testutil/db.go`: exported `SetupTestDB(t *testing.T) *sql.DB` that reads `TEST_DATABASE_URL` (default `postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable`), opens connection via `store.Open()`, registers `t.Cleanup()` that truncates all tables (`TRUNCATE families, parents, children, sessions, auth_events, transactions, allowance_schedules, interest_schedules CASCADE`) and closes connection; include helper functions `CreateTestFamily`, `CreateTestParent`, `CreateTestChild`, `SetRequestContext` to eliminate duplication across test packages

### Store package test updates

- [ ] T023 [P] [US2] Update `backend/internal/store/session_test.go`: replace `testDB()` helper with `testutil.SetupTestDB(t)`, remove `os.Remove`/`t.TempDir()` cleanup, remove `filepath` import
- [ ] T024 [P] [US2] Update `backend/internal/store/auth_event_test.go`: replace `setupTestDB()` with `testutil.SetupTestDB(t)`, remove temp file cleanup
- [ ] T025 [P] [US2] Update all other store test files (`backend/internal/store/family_test.go`, `child_test.go`, `parent_test.go`, `transaction_test.go`, `schedule_test.go`, `interest_test.go`, `interest_schedule_test.go`): replace `testDB()` calls with `testutil.SetupTestDB(t)`, update any direct SQL that uses `?` placeholders to `$1`, update any `db.Write.Exec`/`db.Read.QueryRow` to `db.Exec`/`db.QueryRow`, fix boolean comparisons in test assertions if `is_locked` changed to `bool`

### Handler test updates

- [ ] T026 [P] [US2] Update `backend/internal/balance/handler_test.go`: replace `setupTestDB()` with `testutil.SetupTestDB(t)`, replace local `createTestFamily`/`createTestParent`/`createTestChild`/`setRequestContext` helpers with `testutil` equivalents, remove `os.CreateTemp`/`os.Remove` cleanup
- [ ] T027 [P] [US2] Update `backend/internal/allowance/handler_test.go` and `backend/internal/allowance/scheduler_test.go`: same changes as T026
- [ ] T028 [P] [US2] Update `backend/internal/interest/handler_test.go` and `backend/internal/interest/scheduler_test.go`: same changes as T026

### Test file cleanup

- [ ] T029 [US2] Delete `backend/internal/store/sqlite_test.go` (tests WAL mode, PRAGMA foreign_keys — SQLite-specific)
- [ ] T030 [US2] Create `backend/internal/store/postgres_test.go`: test that `store.Open(dsn)` connects successfully, runs migrations without error, and returns a usable `*sql.DB`; test that calling `Open()` twice is idempotent (migrations already applied)

**Checkpoint**: `go test ./...` passes with a running Postgres instance. All tests use the shared helper. No SQLite temp file patterns remain.

---

## Phase 5: User Story 3 — Seamless Local Development and CI (Priority: P3)

**Goal**: Developers can start the full stack with Docker Compose and CI runs tests against Postgres automatically.

**Independent Test**: `docker-compose up` starts Postgres + backend successfully; CI workflow runs tests with Postgres service container.

- [ ] T031 [US3] Update `docker-compose.yaml`: add `postgres` service (image `postgres:17`, env `POSTGRES_USER/PASSWORD/DB=bankofdad`, port `5432:5432`, healthcheck with `pg_isready`, volume `postgres_data`), add `postgres_data` named volume, remove old `db-data` volume, update backend service to `depends_on: postgres: condition: service_healthy`, replace `DATABASE_PATH` env var with `DATABASE_URL=postgres://bankofdad:bankofdad@postgres:5432/bankofdad?sslmode=disable`, remove old SQLite volume mount
- [ ] T032 [P] [US3] Update `backend/Dockerfile`: remove `RUN mkdir -p /data`, remove `ENV DATABASE_PATH`, remove `VOLUME ["/data"]`, add `ENV DATABASE_URL` placeholder
- [ ] T033 [US3] Update `.github/workflows/ci.yml`: add `services: postgres:` block to backend job (image `postgres:17`, env `POSTGRES_USER/PASSWORD/DB=bankofdad_test`, port `5432:5432`, health options), add `DATABASE_URL` and `TEST_DATABASE_URL` env vars to test step pointing to `localhost:5432/bankofdad_test`, ensure `go test ./...` runs after Postgres is healthy

**Checkpoint**: `docker-compose up` starts cleanly. CI backend job includes Postgres service.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Remove all remaining SQLite references, verify everything compiles and passes.

- [ ] T034 Run `go mod tidy` in `backend/` to remove `modernc.org/sqlite` from `go.sum` and verify `go.mod` is clean
- [ ] T035 Search entire `backend/` directory for remaining references to `sqlite`, `SQLite`, `modernc`, `DATABASE_PATH`, `bankodad.db`, `parseTime`, `.db-wal`, `.db-shm`, `db.Write`, `db.Read`, `*store.DB` — remove or update any found
- [ ] T036 Verify `go build ./...` compiles cleanly in `backend/`
- [ ] T037 Verify `go vet ./...` passes in `backend/`
- [ ] T038 Verify `go test ./...` passes against running Postgres with `TEST_DATABASE_URL` set
- [ ] T039 Verify `docker-compose up` starts Postgres, runs migrations, and backend serves requests
- [ ] T040 Run quickstart.md validation: fresh clone → `docker-compose up -d postgres` → `createdb bankofdad_test` → `go test ./...` → all steps succeed as documented

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2. Core migration — all store files must be rewritten before tests can pass
- **US2 (Phase 4)**: Depends on US1 completion (store files must have valid Postgres SQL before tests can run)
- **US3 (Phase 5)**: Depends on Phase 2 (docker-compose/CI only need the connection layer). Can proceed in parallel with US1/US2 for infrastructure tasks, but final verification needs US1+US2 complete
- **Polish (Phase 6)**: Depends on US1 + US2 + US3

### User Story Dependencies

- **US1 (P1)**: Depends only on Foundational (Phase 2). This is the critical path — all other stories depend on it.
- **US2 (P2)**: Depends on US1. Tests cannot run against Postgres until store queries are rewritten.
- **US3 (P3)**: Docker/CI tasks (T031-T033) can start after Phase 2, but final verification needs US1+US2.

### Within US1 (Store Rewrites)

- T009, T010, T011 (family, parent, auth_event) — simple files, can run in parallel [P]
- T012 (child.go) — boolean handling, should go early since other tests create children
- T013 (session.go) — removes parseTime(), which other files also call; do this early
- T014 (transaction.go) — RETURNING clause; depends on child.go existing
- T015, T016, T017 (schedule, interest, interest_schedule) — can run in parallel after session.go parseTime removal
- T018, T019 (schedulers) — depend on store type changes being complete
- T020 (auth handlers) — depends on store type changes
- T021 (delete sqlite.go) — MUST be last in US1

### Parallel Opportunities

```
Phase 2: T004 and T005 can run in parallel (SQL files vs Go code)
US1:     T009, T010, T011 in parallel (family, parent, auth_event)
US1:     T015, T016, T017 in parallel (schedule, interest, interest_schedule)
US1:     T018, T019 in parallel (allowance scheduler, interest scheduler)
US2:     T023, T024, T025 in parallel (store test files)
US2:     T026, T027, T028 in parallel (handler test files)
US3:     T031, T032 in parallel (docker-compose, Dockerfile)
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T008)
3. Complete Phase 3: US1 store rewrites (T009-T021)
4. **STOP and VALIDATE**: `go build ./...` compiles, manual smoke test against Postgres
5. This is the minimum viable migration — app works against Postgres

### Incremental Delivery

1. Setup + Foundational → Connection layer ready
2. US1 → All features work against Postgres (MVP!)
3. US2 → Full test suite passes against Postgres
4. US3 → Docker and CI infrastructure complete
5. Polish → Clean codebase, no SQLite remnants

### Practical Note

Due to compilation constraints, Phase 2 (T007 — store struct changes) and Phase 3 (T009-T017 — query rewrites) are tightly coupled. The code will not compile between these phases. In practice, an implementer may prefer to rewrite each store file completely (struct change + query rewrite) in a single pass per file, treating T007 and T009-T017 as one logical unit.
