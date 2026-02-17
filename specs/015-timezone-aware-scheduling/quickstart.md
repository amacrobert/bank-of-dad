# Quickstart: Timezone-Aware Scheduling

**Feature**: 015-timezone-aware-scheduling
**Date**: 2026-02-17

## Overview

This feature makes scheduled payments (allowance and interest) fire at midnight in the family's configured timezone instead of midnight UTC, and ensures all date displays in the frontend use the family's timezone.

## Key Changes

### Backend: `schedule_calc.go`
- All date calculation functions (`CalculateNextRun`, `CalculateNextRunAfterExecution`, `nextWeeklyDate`, `nextBiweeklyDate`, `nextMonthlyDate`) accept a `*time.Location` parameter
- Dates are computed as midnight in the given location instead of midnight UTC
- Example: Wednesday in `America/New_York` → `time.Date(2026, 2, 18, 0, 0, 0, 0, loc)` → stored as `2026-02-18T05:00:00Z`

### Backend: Schedulers
- `ListDue()` queries now JOIN with `children` and `families` to return `families.timezone`
- After executing a schedule, the next run is calculated using the family's timezone
- On startup, all active schedules are recalculated with timezone-aware logic (one-time correction)

### Backend: Handlers
- Schedule create/update/resume handlers look up family timezone via `familyStore.GetTimezone(familyID)`
- Pass `*time.Location` to `CalculateNextRun` instead of using `time.UTC`

### Frontend: Timezone Context
- New `TimezoneProvider` wraps the app, fetches family timezone from `GET /api/settings`
- New `useTimezone()` hook provides timezone string to any component
- All date formatting calls use `{ timeZone: familyTimezone }` option

## Development Setup

No new dependencies. No database migrations. No environment variables.

```bash
# Backend
cd backend && go test -p 1 ./...

# Frontend
cd frontend && npm test && npm run lint
```

## Testing Strategy

### Backend Unit Tests
- `schedule_calc_test.go`: Test all frequency types with various timezones (UTC, EST, PST, IST)
- Verify DST transitions (spring forward / fall back) produce correct results
- Verify non-hour-offset timezones (e.g., `Asia/Kolkata` UTC+5:30) work correctly

### Backend Integration Tests
- Scheduler tests: verify due schedules fire at the correct timezone-adjusted time
- Handler tests: verify create/update/resume produce correct `next_run_at` values
- Startup recalculation: verify existing UTC-midnight values are corrected

### Frontend Tests
- Verify date formatting uses family timezone, not browser timezone
- Verify context provides fallback to UTC when settings not loaded

## File Map

| File | Change Type |
|------|-------------|
| `backend/internal/allowance/schedule_calc.go` | Modified — add `*time.Location` param |
| `backend/internal/allowance/schedule_calc_test.go` | Modified — timezone-aware test cases |
| `backend/internal/allowance/scheduler.go` | Modified — familyStore dep, timezone in calc |
| `backend/internal/interest/scheduler.go` | Modified — familyStore dep, timezone in calc |
| `backend/internal/allowance/handler.go` | Modified — familyStore dep, timezone lookup |
| `backend/internal/interest/handler.go` | Modified — familyStore dep, timezone lookup |
| `backend/internal/store/schedule.go` | Modified — ListDue JOINs families |
| `backend/internal/store/interest_schedule.go` | Modified — ListDue JOINs families |
| `backend/main.go` | Modified — pass familyStore to handlers/schedulers |
| `frontend/src/context/TimezoneContext.tsx` | New — timezone React context |
| `frontend/src/components/TransactionsCard.tsx` | Modified — use timezone context |
| `frontend/src/components/ChildAllowanceForm.tsx` | Modified — use timezone context |
| `frontend/src/components/InterestForm.tsx` | Modified — use timezone context |
| `frontend/src/App.tsx` (or equivalent root) | Modified — wrap with TimezoneProvider |
