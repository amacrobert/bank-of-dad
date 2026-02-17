# Implementation Plan: Combine Transaction Cards

**Branch**: `014-combine-transactions` | **Date**: 2026-02-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/014-combine-transactions/spec.md`

## Summary

Replace the separate "Upcoming allowance/interest" and "Recent transactions" cards with a single "Transactions" card containing "Upcoming" and "Recent" sub-sections. This is a frontend-only refactor affecting two pages (`ChildDashboard.tsx` and `ManageChild.tsx`). A new `TransactionsCard` component will be created that combines the data-fetching and rendering logic of `UpcomingPayments.tsx` and `TransactionHistory.tsx`, after which both old components are deleted.

## Technical Context

**Language/Version**: TypeScript 5.3.3 + React 18.2.0
**Primary Dependencies**: react-router-dom, Vite, lucide-react (icons)
**Storage**: N/A (no backend changes)
**Testing**: Manual testing (existing pattern — no frontend test framework in place)
**Target Platform**: Web browser (responsive)
**Project Type**: Web application (frontend-only change)
**Performance Goals**: Match existing load times — no new API calls introduced
**Constraints**: Preserve all existing transaction data display, formatting, and behavior
**Scale/Scope**: 2 pages affected, 1 new component, 2 components removed, 2 pages updated

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | No backend logic changes. Frontend currently has no test framework; this is a presentational refactor that preserves existing behavior. Manual testing against acceptance scenarios is sufficient. |
| II. Security-First Design | PASS | No authentication, authorization, or data handling changes. Same API endpoints, same data displayed. |
| III. Simplicity | PASS | This change *reduces* complexity by consolidating two separate components into one. Removes duplication between parent and child dashboard transaction rendering. YAGNI respected — no new features added. |

**Post-design re-check**: All gates still pass. No new abstractions, dependencies, or complexity introduced.

## Project Structure

### Documentation (this feature)

```text
specs/014-combine-transactions/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (empty — no API changes)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
frontend/src/
├── components/
│   ├── TransactionsCard.tsx        # NEW: combined transactions card
│   ├── UpcomingPayments.tsx        # DELETE: replaced by TransactionsCard
│   ├── TransactionHistory.tsx      # DELETE: replaced by TransactionsCard
│   ├── ManageChild.tsx             # MODIFY: use TransactionsCard
│   └── ui/
│       ├── Card.tsx                # UNCHANGED: used by TransactionsCard
│       └── LoadingSpinner.tsx      # UNCHANGED: used by TransactionsCard
├── pages/
│   ├── ChildDashboard.tsx          # MODIFY: use TransactionsCard
│   └── ParentDashboard.tsx         # UNCHANGED
├── api.ts                          # UNCHANGED: existing API calls reused
└── types.ts                        # UNCHANGED: existing types reused
```

**Structure Decision**: Web application structure. Only the `frontend/src/components/` and `frontend/src/pages/` directories are affected. No backend changes.

## Complexity Tracking

> No constitution violations. This feature reduces complexity.
