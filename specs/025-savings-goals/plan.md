# Implementation Plan: Savings Goals

**Branch**: `025-savings-goals` | **Date**: 2026-03-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/025-savings-goals/spec.md`

## Summary

Children can create savings goals with names, target amounts, optional emojis, and optional target dates. Money is allocated from the child's available balance toward goals (reserved fund model), with progress tracked via visual indicators themed to each child's preferences. Goals reaching their target trigger a celebration animation and move to a completed section. Parents can view goals read-only. The existing withdrawal flow is extended to warn parents when a withdrawal would impact goal allocations.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: pgx/v5 (existing), Vite + Tailwind CSS 4 (existing), lucide-react (existing), Recharts (existing)
**Storage**: PostgreSQL 17 — 2 new tables (`savings_goals`, `goal_allocations`), no changes to existing tables
**Testing**: `go test -p 1 ./...` (backend), `npx tsc --noEmit && npm run build` (frontend)
**Target Platform**: Web application (responsive, mobile-first)
**Project Type**: Web application (Go API + React SPA)
**Performance Goals**: Goal operations < 1s response, dashboard load with goals < 2s
**Constraints**: Max 5 active goals per child, money in int64 cents, interest on total balance
**Scale/Scope**: Small user base (family app), ~6 new API endpoints, ~5 new frontend components, 2 new DB tables

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development — PASS

- All new store methods will have unit tests written before implementation
- All new API endpoints will have handler tests (success + error cases)
- Goal allocation and de-allocation tested for atomicity and edge cases
- Financial calculations (proportional reduction, available balance) have dedicated tests
- No untested code paths affecting balance or goal data

### II. Security-First Design — PASS

- All new endpoints require authentication (`requireAuth` or child-specific auth)
- Authorization enforces: children can only access own goals, parents can view family children's goals (read-only)
- Family ID check prevents cross-family access
- Input validation on all fields: name length, amount ranges, emoji, date format
- Goal allocation amounts validated against available balance to prevent over-allocation
- No new sensitive data introduced (goals are not personally identifiable beyond the child relationship)

### III. Simplicity — PASS

- Follows existing store/handler/route patterns exactly — no new architectural concepts
- 2 new tables with straightforward relationships
- Available balance is computed (not stored) to avoid dual-source-of-truth
- Max 5 active goals keeps queries trivial (no pagination needed for active goals)
- Celebration animation is frontend-only CSS — no backend involvement
- No new dependencies required

### Post-Phase 1 Re-check — PASS

- Data model uses 2 tables with clear relationships — minimal complexity
- API contracts follow existing REST patterns with consistent error responses
- No premature abstractions — direct store methods, no repository pattern
- Goal limit enforced at application layer (simple count query)

## Project Structure

### Documentation (this feature)

```text
specs/025-savings-goals/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: research decisions
├── data-model.md        # Phase 1: database schema
├── quickstart.md        # Phase 1: setup guide
├── contracts/
│   └── api.md           # Phase 1: API endpoint contracts
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── main.go                          # + register goal routes and store
├── migrations/
│   ├── 007_savings_goals.up.sql     # NEW: create tables
│   └── 007_savings_goals.down.sql   # NEW: drop tables
├── internal/
│   ├── store/
│   │   └── savings_goal.go          # NEW: SavingsGoalStore
│   ├── goals/
│   │   └── handler.go               # NEW: goal HTTP handlers
│   ├── balance/
│   │   └── handler.go               # MODIFIED: withdrawal goal impact warning
│   └── testutil/
│       └── db.go                    # MODIFIED: add tables to TRUNCATE

frontend/
├── src/
│   ├── types.ts                     # MODIFIED: add goal types
│   ├── api.ts                       # MODIFIED: add goal API functions
│   ├── App.tsx                      # MODIFIED: add goal routes
│   ├── pages/
│   │   ├── ChildDashboard.tsx       # MODIFIED: add goals summary + balance breakdown
│   │   └── SavingsGoalsPage.tsx     # NEW: child goals management page
│   └── components/
│       ├── BalanceDisplay.tsx        # MODIFIED: add available/saved breakdown
│       ├── GoalCard.tsx             # NEW: goal card with progress
│       ├── GoalForm.tsx             # NEW: create/edit goal form
│       ├── GoalProgressRing.tsx     # NEW: circular progress indicator
│       └── ConfettiCelebration.tsx  # NEW: celebration animation
```

**Structure Decision**: Follows existing web application layout. Backend adds a new `goals/` handler package (paralleling `balance/`, `allowance/`, `interest/`) and a new store file. Frontend adds a new page and goal-specific components.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
