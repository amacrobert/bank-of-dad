# Implementation Plan: Allowance Scheduling

**Branch**: `003-allowance-scheduling` | **Date**: 2026-02-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-allowance-scheduling/spec.md`

## Summary

Add automatic recurring allowance deposits for children. Parents configure schedules (weekly, biweekly, or monthly) that automatically deposit money to their children's accounts. The system uses a background goroutine to process schedules, following the existing session cleanup pattern. Children can see when their next allowance arrives.

## Technical Context

**Language/Version**: Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `modernc.org/sqlite`, `testify`, react-router-dom, Vite (existing - no new deps)
**Storage**: SQLite with WAL mode (existing pattern - separate read/write connections)
**Testing**: Go testing with testify assertions (backend), TypeScript build verification (frontend)
**Target Platform**: Linux server (Docker), Web browsers
**Project Type**: Web application (backend + frontend)
**Performance Goals**: Schedule creation in <1 minute (SC-001), deposits within 1 hour of scheduled time (SC-002)
**Constraints**: Minimum frequency is weekly; amount limits same as manual deposits ($0.01 - $999,999.99)
**Scale/Scope**: Small-scale family finance app, low concurrency expected; schedules processed in background

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development

| Requirement | Status | Notes |
|-------------|--------|-------|
| Write tests before implementation | WILL COMPLY | Tasks will specify test-first workflow |
| Red-Green-Refactor cycle | WILL COMPLY | Each task starts with failing test |
| Contract tests for API endpoints | WILL COMPLY | API tests for schedule CRUD endpoints |
| Integration tests for user journeys | WILL COMPLY | Parent creates schedule, child views upcoming |
| Unit tests for business logic | WILL COMPLY | Next run date calculations, frequency logic |
| No untested financial code | WILL COMPLY | All schedule execution and deposit code tested |

### II. Security-First Design

| Requirement | Status | Notes |
|-------------|--------|-------|
| Data protection | COMPLIANT | SQLite with existing patterns, no new sensitive data |
| Authentication required | WILL COMPLY | All schedule endpoints require auth |
| Authorization (parent/child roles) | WILL COMPLY | Parents create/manage schedules; children view only |
| Input validation | WILL COMPLY | Amount validation, frequency enum, day range |
| Logging | WILL COMPLY | Schedule execution logged; errors tracked |

### III. Simplicity

| Requirement | Status | Notes |
|-------------|--------|-------|
| YAGNI | COMPLIANT | Only weekly/biweekly/monthly; no end dates; no notifications |
| Minimal dependencies | COMPLIANT | Uses Go stdlib time package; no external scheduler |
| Clear over clever | WILL COMPLY | Follow existing store/handler patterns |
| Kid-friendly UX | WILL COMPLY | Simple "Next allowance: $X on Day" display |
| Single responsibility | WILL COMPLY | Separate ScheduleStore, handlers, background processor |
| No premature optimization | COMPLIANT | Simple loop every minute to check schedules |

**Gate Result**: PASS - No constitution violations. Proceed to Phase 0.

---

### Post-Design Re-Check (Phase 1 Complete)

| Principle | Design Artifact | Compliance |
|-----------|----------------|------------|
| Test-First | quickstart.md includes test patterns and TDD workflow | COMPLIANT |
| Security-First | API contract includes auth on all endpoints; authorization rules documented | COMPLIANT |
| Simplicity | No new dependencies; follows existing goroutine pattern; no over-engineering | COMPLIANT |

**Post-Design Gate Result**: PASS - Ready for task generation.

## Project Structure

### Documentation (this feature)

```text
specs/003-allowance-scheduling/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api.yaml         # OpenAPI spec for schedule endpoints
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── allowance/           # NEW: Allowance schedule handlers
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── scheduler.go     # Background processor
│   │   └── scheduler_test.go
│   └── store/
│       ├── schedule.go      # NEW: AllowanceSchedule store
│       └── schedule_test.go
├── main.go                  # Add schedule routes + start scheduler goroutine
└── go.mod

frontend/
├── src/
│   ├── api.ts               # Add schedule API functions
│   ├── types.ts             # Add AllowanceSchedule types
│   ├── pages/
│   │   ├── ParentDashboard.tsx  # Add link to schedule management
│   │   └── ChildDashboard.tsx   # Modify: show upcoming allowance
│   └── components/
│       ├── ScheduleList.tsx         # NEW: Parent's schedule list
│       ├── ScheduleForm.tsx         # NEW: Create/edit schedule form
│       └── UpcomingAllowance.tsx    # NEW: Child's upcoming allowance display
└── tests/
```

**Structure Decision**: Follows existing web application pattern. Schedule logic added to `internal/allowance/` for handlers and scheduler, with store layer in `internal/store/`. Frontend follows existing component patterns. Background scheduler follows session cleanup goroutine pattern.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none) | — | — |
