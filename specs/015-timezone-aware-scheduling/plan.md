# Implementation Plan: Timezone-Aware Scheduling

**Branch**: `015-timezone-aware-scheduling` | **Date**: 2026-02-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/015-timezone-aware-scheduling/spec.md`

## Summary

Scheduled allowance and interest payments currently fire at midnight UTC, causing payments to arrive on the wrong calendar day for non-UTC families (e.g., 7pm Tuesday instead of Wednesday for EST users). The "upcoming" date display also shows the wrong day for the same reason.

**Approach**: Modify the schedule calculation logic (`schedule_calc.go`) to compute `next_run_at` as midnight in the family's configured timezone (stored as its UTC equivalent). Modify schedulers and handlers to pass the family's `*time.Location` to these calculations. On the frontend, add a React Context for the family timezone and use it in all date formatting.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `jackc/pgx/v5`, `testify` (backend); `react-router-dom`, `lucide-react`, Vite (frontend)
**Storage**: PostgreSQL 17 — no schema changes; existing `families.timezone` column used
**Testing**: `go test -p 1 ./...` (backend), `npm test && npm run lint` (frontend)
**Target Platform**: Web application (Docker deployment)
**Project Type**: Web (backend + frontend)
**Performance Goals**: Scheduler tick interval unchanged; timezone lookups via SQL JOIN (negligible overhead)
**Constraints**: DST transitions must not skip or duplicate payments; all IANA timezones supported
**Scale/Scope**: ~6 backend files modified, ~4 frontend files modified/created, 1 new frontend context

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development — PASS
- Schedule calculation tests will be written first with timezone-parameterized cases (various IANA timezones, DST transitions)
- Scheduler integration tests verify correct firing times
- Frontend tests verify timezone-aware date formatting
- All financial calculation code paths covered

### II. Security-First Design — PASS
- No new endpoints or auth changes
- Timezone data comes from existing authenticated `GET /api/settings` endpoint
- No user input changes — timezone is already validated via `time.LoadLocation()` in settings handler
- No new data exposure

### III. Simplicity — PASS
- No new dependencies (Go's `time.Location` is stdlib; `Intl.DateTimeFormat` with `timeZone` is browser-native)
- No new database columns or tables
- Reuses existing `familyStore.GetTimezone()` and `getSettings()` API
- Single React Context (minimal abstraction) instead of prop-drilling
- Startup recalculation is a one-time idempotent loop, not a persistent migration system

### Post-Phase 1 Re-check — PASS
- New `DueAllowanceSchedule` / `DueInterestSchedule` structs are minimal wrappers (embedding + one field)
- SQL JOINs in ListDue are straightforward two-table joins
- No premature abstractions; timezone is threaded directly where needed

## Project Structure

### Documentation (this feature)

```text
specs/015-timezone-aware-scheduling/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research decisions
├── data-model.md        # Phase 1 data model
├── quickstart.md        # Phase 1 quickstart guide
├── contracts/           # Phase 1 API contract changes
│   └── api-changes.md
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── allowance/
│   │   ├── schedule_calc.go        # MODIFIED: add *time.Location param
│   │   ├── schedule_calc_test.go   # MODIFIED: timezone-aware test cases
│   │   ├── scheduler.go            # MODIFIED: familyStore dep, tz in calc, startup recalc
│   │   └── handler.go              # MODIFIED: familyStore dep, tz lookup
│   ├── interest/
│   │   ├── scheduler.go            # MODIFIED: familyStore dep, tz in calc, startup recalc
│   │   └── handler.go              # MODIFIED: familyStore dep, tz lookup
│   └── store/
│       ├── schedule.go             # MODIFIED: ListDue JOINs families for timezone
│       └── interest_schedule.go    # MODIFIED: ListDue JOINs families for timezone
├── main.go                         # MODIFIED: pass familyStore to handlers/schedulers
└── migrations/                     # NO CHANGES

frontend/
└── src/
    ├── context/
    │   └── TimezoneContext.tsx      # NEW: React context for family timezone
    ├── components/
    │   ├── TransactionsCard.tsx     # MODIFIED: use timezone context for dates
    │   ├── ChildAllowanceForm.tsx   # MODIFIED: use timezone context for next run
    │   └── InterestForm.tsx         # MODIFIED: use timezone context for next payout
    └── App.tsx (or root)            # MODIFIED: wrap with TimezoneProvider
```

**Structure Decision**: Follows existing web application structure (backend/ + frontend/). No new directories except `frontend/src/context/` for the new React Context.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
