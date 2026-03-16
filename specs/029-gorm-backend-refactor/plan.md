# Implementation Plan: GORM Backend Refactor

**Branch**: `029-gorm-backend-refactor` | **Date**: 2026-03-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/029-gorm-backend-refactor/spec.md`

## Summary

Migrate the backend data access layer from raw SQL via `pgx/v5/stdlib` to GORM ORM. Extract entity definitions into `backend/models/` and replace `backend/internal/store/` with `backend/repositories/` using GORM's query builder. All handlers and schedulers will be updated to use the new repository layer. No API or behavior changes.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: GORM (`gorm.io/gorm`), GORM PostgreSQL driver (`gorm.io/driver/postgres`), existing `golang-migrate/migrate/v4` (retained), `jackc/pgx/v5` (retained as underlying driver)
**Storage**: PostgreSQL 17 — 12 tables, schema managed by `golang-migrate` (unchanged)
**Testing**: `go test -p 1 ./...` with `testify` — tests adapted to use `*gorm.DB`
**Target Platform**: Linux/macOS server (Docker)
**Project Type**: Web service (Go backend + React frontend)
**Performance Goals**: No regression from current raw SQL performance
**Constraints**: Zero API contract changes; money stays int64 cents; migrations stay in `golang-migrate`
**Scale/Scope**: ~6,800 lines in store package across 11 entity files; 9+ handler packages to update

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development — PASS

- Repository tests will be written before implementation, following TDD.
- Existing store tests serve as the behavioral specification for each repository method.
- All existing tests must pass after migration.

### II. Security-First Design — PASS

- Pure refactor — no changes to auth, authorization, encryption, or input validation.
- GORM's parameterized queries provide equivalent SQL injection protection to `pgx` placeholders.

### III. Simplicity — VIOLATION (JUSTIFIED)

- **Violation**: Adding GORM as a new dependency contradicts "Minimal dependencies: prefer standard library solutions."
- **Justification**: User explicitly requested GORM for better organization and abstractions. See Complexity Tracking below.

## Project Structure

### Documentation (this feature)

```text
specs/029-gorm-backend-refactor/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (internal repository interfaces)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── main.go                      # Updated: GORM init, repository wiring
├── models/                      # NEW: GORM model structs
│   ├── family.go
│   ├── parent.go
│   ├── child.go
│   ├── transaction.go
│   ├── allowance_schedule.go
│   ├── interest_schedule.go
│   ├── refresh_token.go
│   ├── auth_event.go
│   ├── webhook_event.go
│   ├── savings_goal.go
│   └── goal_allocation.go
├── repositories/                # NEW: GORM-based data access
│   ├── family_repo.go
│   ├── family_repo_test.go
│   ├── parent_repo.go
│   ├── parent_repo_test.go
│   ├── child_repo.go
│   ├── child_repo_test.go
│   ├── transaction_repo.go
│   ├── transaction_repo_test.go
│   ├── schedule_repo.go
│   ├── schedule_repo_test.go
│   ├── interest_schedule_repo.go
│   ├── interest_schedule_repo_test.go
│   ├── interest_repo.go
│   ├── interest_repo_test.go
│   ├── refresh_token_repo.go
│   ├── refresh_token_repo_test.go
│   ├── auth_event_repo.go
│   ├── auth_event_repo_test.go
│   ├── webhook_event_repo.go
│   ├── webhook_event_repo_test.go
│   ├── savings_goal_repo.go
│   ├── savings_goal_repo_test.go
│   ├── goal_allocation_repo.go
│   ├── goal_allocation_repo_test.go
│   ├── db.go                    # GORM connection setup (replaces store/postgres.go)
│   └── test_helpers_test.go     # Shared test DB setup for GORM
├── internal/
│   ├── auth/                    # Updated: uses repositories
│   ├── allowance/               # Updated: uses repositories
│   ├── interest/                # Updated: uses repositories
│   ├── balance/                 # Updated: uses repositories
│   ├── family/                  # Updated: uses repositories
│   ├── settings/                # Updated: uses repositories
│   ├── goals/                   # Updated: uses repositories
│   ├── subscription/            # Updated: uses repositories
│   ├── contact/                 # Updated: uses repositories
│   ├── middleware/              # Unchanged
│   ├── config/                  # Unchanged
│   ├── testutil/                # Updated: GORM test helpers
│   └── store/                   # REMOVED after migration complete
└── go.mod                       # Updated: gorm.io/gorm, gorm.io/driver/postgres
```

**Structure Decision**: Models and repositories live at `backend/models/` and `backend/repositories/` (top-level backend packages) per user request, not under `internal/`. This makes them importable from any backend package without import cycle concerns.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Adding GORM dependency (~Simplicity principle) | User explicitly requested GORM for better organization and abstractions. GORM reduces SQL boilerplate, centralizes model definitions, and provides a consistent query builder API. | Raw SQL with `pgx` is the current approach. It works but requires manual SQL string management, manual NULL scanning, and entity definitions scattered across store files. The user wants the abstraction GORM provides. |
| Repository pattern (~Simplicity principle) | User explicitly requested `repositories/` package with repository methods for all database interactions. Provides clear separation of models from data access. | Direct store-on-DB pattern is simpler but doesn't meet the user's stated goal of better organization and abstractions. |
