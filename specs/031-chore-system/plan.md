# Implementation Plan: Chore & Task System

**Branch**: `031-chore-system` | **Date**: 2026-03-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/031-chore-system/spec.md`

## Summary

Add a chore system where parents create tasks with reward amounts, children mark them complete, and parents approve to trigger automatic deposits. Supports one-time and recurring chores (daily/weekly/monthly) with a background scheduler to generate new instances and expire missed ones.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: GORM (gorm.io/gorm), pgx/v5 (driver), Vite, Tailwind CSS 4, lucide-react
**Storage**: PostgreSQL 17 — 3 new tables (`chores`, `chore_assignments`, `chore_instances`), 1 new transaction type
**Testing**: `go test -p 1 ./...` (backend), `npx tsc --noEmit && npm run build` (frontend)
**Target Platform**: Web application (responsive browser)
**Project Type**: Web service (Go HTTP backend + React SPA frontend)
**Performance Goals**: Chore lifecycle under 2 minutes; scheduler processes within 5-minute tick
**Constraints**: Family timezone-aware scheduling; atomic deposit on approval; disabled child accounts blocked
**Scale/Scope**: Small family app — dozens of chores per family, not thousands

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Test-First Development** | PASS | Contract tests for all chore endpoints; integration tests for approve→deposit flow; unit tests for scheduler logic |
| **II. Security-First Design** | PASS | All endpoints require auth; parent-only for create/edit/approve/reject; family-scoped data access; child role limited to view+complete |
| **III. Simplicity** | PASS | Follows existing patterns (repo, handler, scheduler); no new dependencies; reuses transaction system for deposits |

No constitution violations. Complexity Tracking section not needed.

## Project Structure

### Documentation (this feature)

```text
specs/031-chore-system/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api-endpoints.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── models/
│   └── chore.go                  # Chore, ChoreAssignment, ChoreInstance models
├── repositories/
│   ├── chore_repo.go             # Chore CRUD + assignment operations
│   ├── chore_repo_test.go
│   ├── chore_instance_repo.go    # Instance lifecycle (create, complete, approve, reject, expire)
│   ├── chore_instance_repo_test.go
│   └── transaction_repo.go       # Add DepositChore method (existing file)
├── internal/
│   └── chore/
│       ├── handler.go            # HTTP handlers for chore endpoints
│       ├── handler_test.go
│       ├── scheduler.go          # Background processor for recurring instances + expiry
│       └── scheduler_test.go
├── migrations/
│   ├── 011_chores.up.sql
│   └── 011_chores.down.sql
└── main.go                       # Wire chore handler + scheduler + routes

frontend/
├── src/
│   ├── types.ts                  # Add Chore, ChoreInstance, ChoreAssignment types
│   ├── pages/
│   │   ├── ParentChores.tsx      # Parent chore management + approval queue
│   │   └── ChildChores.tsx       # Child chore list + completion
│   └── components/
│       ├── ChoreForm.tsx         # Create/edit chore form
│       ├── ChoreCard.tsx         # Individual chore display (parent + child variants)
│       └── ChoreApprovalQueue.tsx # Pending approvals list
└── App.tsx                       # Add chore routes
```

**Structure Decision**: Web application structure matching existing codebase conventions. New `chore` package under `internal/` for handler and scheduler (parallels `allowance/` and `interest/`). Models and repositories at top level following the GORM refactor pattern. Frontend adds two new pages and three components.
