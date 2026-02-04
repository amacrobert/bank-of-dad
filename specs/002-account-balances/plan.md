# Implementation Plan: Account Balances

**Branch**: `002-account-balances` | **Date**: 2026-02-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-account-balances/spec.md`

## Summary

Add account balance tracking for children with parent-managed deposits/withdrawals. Parents view all children's balances on their dashboard and can add/remove money. Children view their own balance and full transaction history (read-only). Follows existing Go backend patterns with SQLite storage and React frontend.

## Technical Context

**Language/Version**: Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `modernc.org/sqlite`, `testify`, react-router-dom, Vite
**Storage**: SQLite with WAL mode (existing pattern - separate read/write connections)
**Testing**: Go testing with testify assertions (backend), Vite test runner (frontend)
**Target Platform**: Linux server (Docker), Web browsers
**Project Type**: Web application (backend + frontend)
**Performance Goals**: Dashboard loads in <2 seconds (per SC-001, SC-004)
**Constraints**: Balance amounts stored with 2 decimal precision, no negative balances
**Scale/Scope**: Small-scale family finance app, low concurrency expected

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development

| Requirement | Status | Notes |
|-------------|--------|-------|
| Write tests before implementation | WILL COMPLY | Tasks will specify test-first workflow |
| Red-Green-Refactor cycle | WILL COMPLY | Each task starts with failing test |
| Contract tests for API endpoints | WILL COMPLY | API tests for balance/transaction endpoints |
| Integration tests for user journeys | WILL COMPLY | Parent deposit, child view flows |
| Unit tests for business logic | WILL COMPLY | Balance calculations, validation |
| No untested financial code | WILL COMPLY | All balance operations tested |

### II. Security-First Design

| Requirement | Status | Notes |
|-------------|--------|-------|
| Data protection | COMPLIANT | SQLite with existing patterns, no new sensitive data types |
| Authentication required | WILL COMPLY | All balance endpoints require auth |
| Authorization (parent/child roles) | WILL COMPLY | Parents modify own children only; children view own only |
| Input validation | WILL COMPLY | Amount validation, note sanitization |
| Logging | WILL COMPLY | Transactions table serves as audit log |

### III. Simplicity

| Requirement | Status | Notes |
|-------------|--------|-------|
| YAGNI | COMPLIANT | No multi-currency, scheduled deposits, or interest |
| Minimal dependencies | COMPLIANT | No new dependencies required |
| Clear over clever | WILL COMPLY | Follow existing store/handler patterns |
| Kid-friendly UX | WILL COMPLY | Simple balance display, clear transaction history |
| Single responsibility | WILL COMPLY | Separate BalanceStore, TransactionStore, handlers |
| No premature optimization | COMPLIANT | Simple queries, pagination if needed later |

**Gate Result**: PASS - No constitution violations. Proceed to Phase 0.

---

### Post-Design Re-Check (Phase 1 Complete)

| Principle | Design Artifact | Compliance |
|-----------|----------------|------------|
| Test-First | quickstart.md includes test patterns and TDD workflow | COMPLIANT |
| Security-First | API contract includes auth on all endpoints; authorization rules documented | COMPLIANT |
| Simplicity | No new dependencies; follows existing patterns; no over-engineering | COMPLIANT |

**Post-Design Gate Result**: PASS - Ready for task generation.

## Project Structure

### Documentation (this feature)

```text
specs/002-account-balances/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api.yaml         # OpenAPI spec for balance endpoints
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── balance/         # NEW: Balance and transaction handlers
│   │   ├── handler.go
│   │   └── handler_test.go
│   └── store/
│       ├── balance.go       # NEW: Balance store
│       ├── balance_test.go  # NEW: Balance store tests
│       ├── transaction.go   # NEW: Transaction store
│       └── transaction_test.go
├── main.go              # Add balance routes
└── go.mod

frontend/
├── src/
│   ├── api.ts           # Add balance API functions
│   ├── types.ts         # Add balance/transaction types
│   ├── pages/
│   │   ├── ParentDashboard.tsx  # Modify: show balances
│   │   └── ChildDashboard.tsx   # Modify: show balance + history
│   └── components/
│       ├── BalanceDisplay.tsx      # NEW: Balance component
│       ├── TransactionHistory.tsx  # NEW: Transaction list
│       ├── DepositForm.tsx         # NEW: Parent deposit UI
│       └── WithdrawForm.tsx        # NEW: Parent withdraw UI
└── tests/
```

**Structure Decision**: Follows existing web application pattern. Balance logic added to `internal/balance/` for handlers, with store layer additions in `internal/store/`. Frontend follows existing component patterns.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none) | — | — |
