# Quickstart: Chore & Task System

**Feature Branch**: `031-chore-system`
**Created**: 2026-03-22

## Prerequisites

- Go 1.24 installed
- PostgreSQL 17 running locally (port 5432)
- Node.js + npm installed
- `bankofdad_test` database available for tests

## Getting Started

### 1. Run the migration

```bash
cd backend
# Apply the new migration (creates chores, chore_assignments, chore_instances tables)
migrate -path migrations -database "postgres://localhost:5432/bankofdad?sslmode=disable" up
```

### 2. Run backend tests

```bash
cd backend
go test -p 1 ./...
```

### 3. Run frontend checks

```bash
cd frontend
npx tsc --noEmit && npm run build
npm run lint
```

### 4. Start the app

```bash
# Terminal 1: Backend
cd backend
go run main.go

# Terminal 2: Frontend
cd frontend
npm run dev
```

## Key Files to Know

| File | Purpose |
|------|---------|
| `backend/models/chore.go` | Chore, ChoreAssignment, ChoreInstance GORM models |
| `backend/repositories/chore_repo.go` | Chore CRUD + assignment management |
| `backend/repositories/chore_instance_repo.go` | Instance lifecycle (create, complete, approve, reject, expire) |
| `backend/internal/chore/handler.go` | HTTP handlers for all chore endpoints |
| `backend/internal/chore/scheduler.go` | Background processor for recurring instances + expiry |
| `backend/migrations/011_chores.up.sql` | Database migration |
| `frontend/src/pages/ParentChores.tsx` | Parent chore management UI |
| `frontend/src/pages/ChildChores.tsx` | Child chore list + completion UI |

## Testing the Feature

### Manual Test Flow

1. **Login as parent** → navigate to chore management
2. **Create a one-time chore**: "Mow the lawn", $5.00, assign to a child
3. **Login as child** → see the chore in "Available" section
4. **Mark as complete** → chore moves to "Pending Approval"
5. **Login as parent** → see pending approval in queue
6. **Approve** → verify $5.00 deposit appears in child's transaction history
7. **Check child balance** → should increase by $5.00

### Recurring Chore Test Flow

1. Create a weekly recurring chore
2. Wait for scheduler tick (5 minutes) or trigger manually in tests
3. Verify new instance appears for current period
4. Complete and approve it
5. Verify next period's instance is generated on next tick

## Architecture Notes

- **Scheduler**: Runs every 5 minutes (same as allowance scheduler). Two jobs per tick:
  1. Generate new instances for active recurring chores where current period has no instance
  2. Expire available instances whose period has ended
- **Deposits**: Uses `TransactionRepo.DepositChore()` — same atomic pattern as allowance deposits
- **Family scoping**: All queries filter by family_id from JWT context
- **Disabled children**: Approval blocked; instance generation skipped
