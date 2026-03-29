# Implementation Plan: Withdrawal Requests

**Branch**: `032-withdrawal-requests` | **Date**: 2026-03-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/032-withdrawal-requests/spec.md`

## Summary

Enable children to request withdrawals from their accounts, creating a parent-approval workflow similar to the existing chore approval system. Children submit requests with an amount and reason; parents review, approve, or deny them. Approved requests create a distinct "withdrawal_request" transaction type. One pending request per child at a time.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: GORM (gorm.io/gorm), pgx/v5 (driver), Vite, Tailwind CSS 4, lucide-react
**Storage**: PostgreSQL 17 — 1 new table (`withdrawal_requests`), 1 new transaction type
**Testing**: `go test -p 1 ./...` (backend), `npx tsc --noEmit && npm run build` (frontend)
**Target Platform**: Web application (browser)
**Project Type**: Web service (Go backend + React SPA frontend)
**Performance Goals**: Standard web app — requests visible immediately after submission
**Constraints**: Single pending request per child; validate against available balance (minus goal allocations)
**Scale/Scope**: Family-scale app (small number of users per family)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Contract tests for all new endpoints, integration tests for approval workflow, unit tests for balance validation logic |
| II. Security-First Design | PASS | All endpoints require auth; child endpoints use `requireAuth`, parent endpoints use `requireParent`; children can only access own requests; parents can only access family requests; input validation on amount and reason |
| III. Simplicity | PASS | Follows existing chore approval pattern exactly; one new table; no new dependencies; single pending request limit keeps UX simple |

No constitution violations. No complexity tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/032-withdrawal-requests/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api-endpoints.md
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── models/
│   └── withdrawal_request.go    # New model + status constants
├── repositories/
│   └── withdrawal_request_repo.go  # New repo (GORM)
├── internal/
│   └── withdrawal/
│       └── handler.go           # New handler (child submit, parent approve/deny, list, cancel)
├── migrations/
│   ├── 013_withdrawal_requests.up.sql
│   └── 013_withdrawal_requests.down.sql
└── main.go                      # Route registration

frontend/
├── src/
│   ├── types.ts                 # Add WithdrawalRequest types
│   ├── api.ts                   # Add withdrawal request API functions
│   ├── pages/
│   │   └── ChildDashboard.tsx   # Add pending request display + request form trigger
│   └── components/
│       ├── WithdrawalRequestForm.tsx      # Child: submit request (amount + reason)
│       ├── WithdrawalRequestCard.tsx      # Shared: display a request with status
│       ├── PendingWithdrawalRequests.tsx  # Parent: list + approve/deny pending requests
│       └── ManageChild.tsx               # Add pending request indicator + review UI
```

**Structure Decision**: Follows existing patterns — new `withdrawal/` handler package parallels `chore/` package; new GORM model in `models/`; new repo in `repositories/`. Frontend adds components to existing pages rather than creating new routes, keeping the child dashboard as the central hub.
