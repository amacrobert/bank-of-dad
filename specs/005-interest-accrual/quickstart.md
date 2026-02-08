# Quickstart: Interest Accrual

## Overview

This feature adds interest accrual to child savings accounts. Parents set per-child annual interest rates, and the system automatically calculates and credits monthly interest as transactions.

## Key Flows

### 1. Parent Sets Interest Rate
1. Parent navigates to child management
2. Sets annual interest rate (e.g., 5%)
3. Rate is saved and displayed on the child's account

### 2. Monthly Interest Accrual
1. Background scheduler runs periodically (e.g., hourly)
2. Finds children with `interest_rate_bps > 0`, `balance_cents > 0`, and `last_interest_at` not in current month
3. Calculates: `balance_cents * rate_bps / 12 / 10000`
4. Creates an `interest` transaction and updates `balance_cents` and `last_interest_at` atomically
5. Skips if calculated interest rounds to $0.00

### 3. Viewing Interest
- Interest transactions appear in existing transaction history with type `interest`
- Both parents and children see interest transactions labeled "Interest earned"

## Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `backend/internal/store/sqlite.go` | Modify | Add migrations (columns + CHECK constraint) |
| `backend/internal/store/interest.go` | Create | Interest store (set rate, apply interest, query) |
| `backend/internal/store/interest_test.go` | Create | Store tests |
| `backend/internal/interest/scheduler.go` | Create | Background interest scheduler |
| `backend/internal/interest/scheduler_test.go` | Create | Scheduler tests |
| `backend/internal/interest/handler.go` | Create | HTTP handler for interest rate endpoint |
| `backend/internal/interest/handler_test.go` | Create | Handler tests |
| `backend/internal/balance/handler.go` | Modify | Include interest rate in balance response |
| `backend/main.go` | Modify | Wire up interest scheduler and routes |
| `frontend/src/types.ts` | Modify | Add `interest` transaction type |
| `frontend/src/components/TransactionHistory.tsx` | Modify | Display interest transactions |
| `frontend/src/components/ChildCard.tsx` or similar | Modify | Show interest rate on parent dashboard |
| `frontend/src/components/InterestRateForm.tsx` | Create | Interest rate configuration UI |

## Verification

```bash
# Backend tests
cd backend && go test ./internal/store/... ./internal/interest/...

# Manual verification
# 1. Set interest rate via API
curl -X PUT http://localhost:8080/api/children/1/interest-rate \
  -H 'Content-Type: application/json' \
  -d '{"interest_rate_bps": 500}'

# 2. Trigger interest accrual (scheduler runs automatically)
# 3. Check transactions for interest entry
curl http://localhost:8080/api/children/1/transactions
```
