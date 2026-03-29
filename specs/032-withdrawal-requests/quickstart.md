# Quickstart: Withdrawal Requests

**Feature**: 032-withdrawal-requests | **Date**: 2026-03-28

## Prerequisites

- Go 1.24, Node.js 24, PostgreSQL 17 running locally
- `bankofdad_test` database for tests
- Feature branch `032-withdrawal-requests` checked out

## Implementation Order

### 1. Database Migration

```bash
# Create migration files
# backend/migrations/013_withdrawal_requests.up.sql
# backend/migrations/013_withdrawal_requests.down.sql
```

Creates `withdrawal_requests` table and updates the transaction type check constraint to include `withdrawal_request`.

### 2. Backend Model

```bash
# backend/models/withdrawal_request.go
```

Define `WithdrawalRequest` struct with GORM tags and `WithdrawalRequestStatus` constants.

### 3. Backend Repository

```bash
# backend/repositories/withdrawal_request_repo.go
```

GORM-based repo following `ChoreRepo` / `ChoreInstanceRepo` patterns:
- `Create(req)` — insert with pending status check
- `GetByID(id)` — fetch single request
- `ListByChild(childID, statusFilter)` — child's requests
- `ListByFamily(familyID, statusFilter)` — parent's family requests
- `Approve(id, parentID, transactionID)` — pending → approved
- `Deny(id, parentID, reason)` — pending → denied
- `Cancel(id, childID)` — pending → cancelled
- `PendingCountByFamily(familyID)` — badge count

### 4. Backend Handler

```bash
# backend/internal/withdrawal/handler.go
```

Handler struct with repos (withdrawal request, transaction, child, balance). Methods:
- `HandleSubmitRequest` — child POST
- `HandleListRequests` — child GET / parent GET
- `HandleCancelRequest` — child POST cancel
- `HandleApprove` — parent POST approve (creates transaction)
- `HandleDeny` — parent POST deny
- `HandlePendingCount` — parent GET count

### 5. Route Registration

Add routes in `backend/main.go`:
```go
// Child endpoints
mux.Handle("POST /api/child/withdrawal-requests", requireAuth(...))
mux.Handle("GET /api/child/withdrawal-requests", requireAuth(...))
mux.Handle("POST /api/child/withdrawal-requests/{id}/cancel", requireAuth(...))

// Parent endpoints
mux.Handle("GET /api/withdrawal-requests", requireParent(...))
mux.Handle("POST /api/withdrawal-requests/{id}/approve", requireParent(...))
mux.Handle("POST /api/withdrawal-requests/{id}/deny", requireParent(...))
mux.Handle("GET /api/withdrawal-requests/pending/count", requireParent(...))
```

### 6. Frontend Types & API

Add types to `frontend/src/types.ts` and API functions to `frontend/src/api.ts`.

### 7. Frontend Components

- `WithdrawalRequestForm.tsx` — child submits request (amount + reason)
- `WithdrawalRequestCard.tsx` — displays a request with status badge
- `PendingWithdrawalRequests.tsx` — parent reviews pending requests

### 8. Frontend Integration

- **ChildDashboard**: Add pending request display + "Request Withdrawal" button
- **ManageChild**: Add pending request indicator + approve/deny UI
- **ParentDashboard**: Add badge for pending request count

## Running Tests

```bash
cd backend && go test -p 1 ./...
```

## Key Patterns to Follow

- **Transaction type**: Use `withdrawal_request` (not `withdrawal`) for approved requests
- **Balance check**: Use available balance (total minus goal allocations)
- **Status transitions**: Use `RowsAffected == 0` pattern from chore approval to detect invalid transitions
- **Goal impact**: Reuse existing goal-impact warning flow from `balance/handler.go`
- **Auth**: `requireAuth` for child endpoints, `requireParent` for parent endpoints
