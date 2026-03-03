# Quickstart: Savings Goals

**Branch**: `025-savings-goals` | **Date**: 2026-03-02

## Prerequisites

- Go 1.24+
- Node.js 18+ / npm
- PostgreSQL 17 running on localhost:5432
- `bankofdad` and `bankofdad_test` databases

## Setup

```bash
# Checkout the feature branch
git checkout 025-savings-goals

# Backend: run migrations
cd backend
go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
  -path migrations -database "postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable" up

# Backend: start server
go run main.go

# Frontend: install deps and start dev server
cd ../frontend
npm install
npm run dev
```

## Running Tests

```bash
# Backend tests (serial execution required for shared test DB)
cd backend && go test -p 1 ./...

# Frontend type check + build
cd frontend && npx tsc --noEmit && npm run build

# Frontend lint
cd frontend && npm run lint
```

## New Files (this feature)

### Backend

| File | Purpose |
| ---- | ------- |
| `backend/migrations/007_savings_goals.up.sql` | Create savings_goals and goal_allocations tables |
| `backend/migrations/007_savings_goals.down.sql` | Drop savings_goals and goal_allocations tables |
| `backend/internal/store/savings_goal.go` | SavingsGoalStore — CRUD + allocation operations |
| `backend/internal/goals/handler.go` | HTTP handlers for savings goal endpoints |

### Frontend

| File | Purpose |
| ---- | ------- |
| `frontend/src/pages/SavingsGoalsPage.tsx` | Child-facing goals page (create, view, manage) |
| `frontend/src/components/GoalCard.tsx` | Individual goal card with progress indicator |
| `frontend/src/components/GoalForm.tsx` | Create/edit goal form |
| `frontend/src/components/GoalProgressRing.tsx` | Circular progress indicator for goals |
| `frontend/src/components/ConfettiCelebration.tsx` | Celebration animation for goal completion |

### Modified Files

| File | Change |
| ---- | ------ |
| `backend/main.go` | Register new savings goal routes and store |
| `backend/internal/balance/handler.go` | Add goal impact warning to withdrawal flow |
| `backend/internal/store/transaction.go` | Add proportional goal reduction on impacted withdrawals |
| `backend/internal/testutil/db.go` | Add savings_goals, goal_allocations to TRUNCATE list |
| `frontend/src/types.ts` | Add SavingsGoal, GoalAllocation types |
| `frontend/src/api.ts` | Add savings goal API functions |
| `frontend/src/App.tsx` | Add /child/goals route |
| `frontend/src/pages/ChildDashboard.tsx` | Add goals summary section, update balance display with available/saved breakdown |
| `frontend/src/components/BalanceDisplay.tsx` | Add optional available/saved breakdown display |

## Key Patterns to Follow

- **Store**: `NewSavingsGoalStore(db *sql.DB)`, `$1/$2` placeholders, `RETURNING id`, atomic transactions
- **Handler**: struct with store deps, `writeJSON` helper, `ErrorResponse` type, auth context extraction
- **Routes**: `mux.Handle("METHOD /path", middleware(http.HandlerFunc(handler.Method)))`
- **Money**: int64 cents throughout
- **Frontend**: React hooks, Card/Button/Input components, Tailwind theme colors, `api.ts` fetch wrapper
