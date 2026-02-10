# Implementation Plan: Account Management Enhancements

**Branch**: `006-account-management-enhancements` | **Date**: 2026-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-account-management-enhancements/spec.md`

## Summary

Consolidate child management into a single unified view. Parents can view transaction history, configure interest rates (pre-populated), manage a single allowance, and set interest accrual schedules — all from the ManageChild form. The interest accrual system is extended from fixed monthly to flexible scheduling (weekly/biweekly/monthly). Children see their interest rate and next payment date on their dashboard. The standalone allowance section is removed from the parent dashboard.

## Technical Context

**Language/Version**: Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `modernc.org/sqlite`, `testify`, react-router-dom, Vite
**Storage**: SQLite with WAL mode (separate read/write connections)
**Testing**: `go test` with `testify` (backend), `tsc --noEmit` + `vite build` (frontend)
**Target Platform**: Web application (macOS/Linux server, modern browsers)
**Project Type**: Web application (backend + frontend)
**Performance Goals**: Transaction history loads in under 2 seconds, interest accrual within 1 hour of scheduled time
**Constraints**: Integer arithmetic for all money calculations, single-family personal use
**Scale/Scope**: Single-family app, ~2-5 children, minimal concurrency

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development

- **PASS**: All store and handler changes will have tests written first (TDD red-green-refactor)
- Interest schedule store: contract tests for CRUD, ListDue
- Allowance one-per-child constraint: unit tests for enforcement
- Modified ApplyInterest: unit tests for proration (weekly/biweekly/monthly)
- Handler tests for new endpoints (allowance CRUD, interest schedule CRUD)
- No untested code paths for financial calculations

### II. Security-First Design

- **PASS**: All new endpoints require authentication via existing middleware
- Parent-only endpoints use `requireParent` middleware
- Child-accessible endpoints verify `familyID` match and `userType` restrictions
- Input validation on all request bodies (amount, frequency, day ranges)
- No new sensitive data types introduced

### III. Simplicity

- **PASS**: Reuses existing patterns (store, handler, scheduler)
- Interest schedule table mirrors allowance schedule structure — familiar, consistent
- One allowance per child simplifies the data model
- Consolidating into ManageChild reduces UI complexity
- No new dependencies added

### Post-Design Re-check

- **PASS**: All design decisions follow established patterns
- `InterestScheduleStore` follows `ScheduleStore` patterns exactly
- Schedule calculation reuses `schedule_calc.go` functions
- New API endpoints follow existing REST conventions
- Frontend components follow existing form/display patterns

## Project Structure

### Documentation (this feature)

```text
specs/006-account-management-enhancements/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── api.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── store/
│   │   ├── schedule.go          # Modified: add GetByChildID, UNIQUE constraint migration
│   │   ├── schedule_test.go     # Modified: add one-per-child tests
│   │   ├── interest.go          # Modified: ApplyInterest periodsPerYear param
│   │   ├── interest_test.go     # Modified: update tests for new param
│   │   ├── interest_schedule.go # New: InterestScheduleStore
│   │   ├── interest_schedule_test.go # New: store tests
│   │   └── sqlite.go            # Modified: migrations
│   ├── interest/
│   │   ├── handler.go           # Modified: add interest schedule endpoints
│   │   ├── handler_test.go      # Modified: add schedule endpoint tests
│   │   ├── scheduler.go         # Modified: use interest_schedules instead of last_interest_at
│   │   └── scheduler_test.go    # Modified: update scheduler tests
│   └── allowance/
│       ├── handler.go           # Modified: add child-scoped allowance endpoints
│       └── handler_test.go      # Modified: add child-scoped endpoint tests
├── main.go                      # Modified: wire new routes

frontend/
├── src/
│   ├── types.ts                 # Modified: add InterestSchedule type, update BalanceResponse
│   ├── api.ts                   # Modified: add new API functions
│   ├── components/
│   │   ├── ManageChild.tsx      # Modified: add transaction history, allowance form, interest schedule
│   │   ├── ChildAllowanceForm.tsx   # New: inline allowance management
│   │   ├── InterestScheduleForm.tsx # New: interest schedule configuration
│   │   └── InterestRateForm.tsx     # Modified: fix pre-population on async load
│   └── pages/
│       ├── ParentDashboard.tsx  # Modified: remove standalone allowance section
│       └── ChildDashboard.tsx   # Modified: show interest rate and next payment
```

**Structure Decision**: Web application structure (existing), extending backend/frontend split with new store, handler, and component files following established patterns.
