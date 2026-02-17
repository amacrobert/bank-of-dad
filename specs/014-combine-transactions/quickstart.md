# Quickstart: Combine Transaction Cards

**Feature**: 014-combine-transactions
**Date**: 2026-02-16

## Prerequisites

- Node.js and npm installed
- Docker and Docker Compose installed (for full-stack local dev)
- Repository cloned and on branch `014-combine-transactions`

## Local Development Setup

```bash
# Start the full stack (backend + frontend + database)
docker compose up

# Or for frontend-only development (if backend is already running):
cd frontend
npm install
npm run dev
```

## What to Build

### New Component
- **`frontend/src/components/TransactionsCard.tsx`**: Combined transactions card with "Upcoming" and "Recent" sub-sections

### Files to Modify
- **`frontend/src/pages/ChildDashboard.tsx`**: Replace `UpcomingPayments` + `TransactionHistory` usage with `TransactionsCard`
- **`frontend/src/components/ManageChild.tsx`**: Replace `UpcomingPayments` + `TransactionHistory` usage with `TransactionsCard`

### Files to Delete
- **`frontend/src/components/UpcomingPayments.tsx`**: Replaced by TransactionsCard
- **`frontend/src/components/TransactionHistory.tsx`**: Replaced by TransactionsCard

## Testing

Manual testing on both dashboards:

1. **Parent dashboard**: Log in as parent → select a child → verify single "Transactions" card with "Upcoming" and "Recent" sections
2. **Child dashboard**: Log in as child → verify single "Transactions" card with "Upcoming" and "Recent" sections
3. **Empty states**: Test with a child that has no allowance schedules and no transactions
4. **Data integrity**: Compare transaction data displayed before and after the change to confirm no information loss

## Key APIs Used (Unchanged)

| Endpoint | Purpose |
|----------|---------|
| `GET /api/children/{id}/transactions` | Fetch recent/past transactions |
| `GET /api/children/{childId}/upcoming-allowances` | Fetch upcoming allowance payments |
| `GET /api/children/{childId}/interest-schedule` | Fetch interest schedule for upcoming interest |
| `GET /api/children/{id}/balance` | Fetch balance and interest rate |
