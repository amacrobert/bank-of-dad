# Research: Timezone-Aware Scheduling

**Feature**: 015-timezone-aware-scheduling
**Date**: 2026-02-17

## Root Cause Analysis

### Problem 1: Payments fire at the wrong time
- **schedule_calc.go** computes `next_run_at` as midnight UTC: `time.Date(year, month, day, 0, 0, 0, 0, time.UTC)`
- When the scheduler evaluates `next_run_at <= time.Now()`, it fires at midnight UTC
- For America/New_York (UTC-5), midnight UTC = 7pm previous day — payment fires on the wrong calendar date

### Problem 2: "Upcoming" date displays incorrectly
- `next_run_at` is stored as midnight UTC (e.g., `2026-02-18T00:00:00Z` for Wednesday)
- Frontend parses this with `new Date("2026-02-18T00:00:00Z")` and calls `.toLocaleDateString(undefined, ...)` using the browser's timezone
- In EST browser, midnight UTC = 7pm Feb 17 EST → displays as "Feb 17" instead of "Feb 18"

### Problem 3: Transaction dates inconsistent
- Transactions have `created_at` stored as actual UTC timestamp of when they occurred
- Browser displays them in local timezone, which is usually correct for the browser user
- But if the family timezone differs from browser timezone, or if a scheduler-created transaction fires at midnight UTC, the displayed date can be off

## Research Decisions

### Decision 1: Where to apply timezone — schedule calculation, not storage

- **Decision**: Modify `schedule_calc.go` to accept a `*time.Location` parameter. Compute `next_run_at` as midnight in the family's timezone, stored as the equivalent UTC timestamp in the database.
- **Rationale**: The `next_run_at` column is already `TIMESTAMPTZ`. By storing "midnight EST" as its UTC equivalent (5:00 AM UTC), the scheduler's existing `next_run_at <= time.Now()` check naturally fires at the correct time. No changes needed to the scheduler's ListDue query logic.
- **Alternatives considered**:
  - Store timezone offset alongside `next_run_at` — rejected: adds column for no benefit since we can derive it from the family's timezone
  - Convert `time.Now()` to family timezone before comparing — rejected: would require per-family timezone lookups in the hot path of ListDue; storing the correct UTC target time is cleaner

### Decision 2: How schedulers access family timezone — SQL JOIN

- **Decision**: Modify `ListDue()` queries for both allowance and interest schedules to JOIN with `children` and `families` tables, returning `families.timezone` alongside each schedule.
- **Rationale**: Single query, no N+1 lookups. The scheduler processes one schedule at a time and needs the timezone to compute the next `next_run_at` after execution.
- **Alternatives considered**:
  - Add `familyStore` dependency and lookup per schedule — rejected: N+1 queries
  - Cache family timezones — rejected: premature optimization (YAGNI), and timezone can change

### Decision 3: How handlers access family timezone — existing familyStore

- **Decision**: Add `familyStore` dependency to allowance and interest handlers. Use `familyStore.GetTimezone(familyID)` where `familyID` comes from JWT context (already available).
- **Rationale**: Handlers already have familyID from auth context. One extra query per schedule create/update/resume is negligible.
- **Alternatives considered**:
  - Pass timezone in request body — rejected: timezone is a family-level setting, not a per-request parameter
  - Include timezone in JWT claims — rejected: would require re-login when timezone changes

### Decision 4: Frontend date formatting — React Context for family timezone

- **Decision**: Create a `TimezoneProvider` React context that fetches the family timezone from `GET /api/settings` once on app load. All date formatting functions use this timezone via `toLocaleDateString(undefined, { timeZone: familyTimezone, ... })`.
- **Rationale**: The frontend currently has no React Context. A single context avoids threading timezone as a prop through many components. The `getSettings()` API already exists.
- **Alternatives considered**:
  - Include timezone in each API response — rejected: duplicates data, requires changing many endpoints
  - Use browser timezone — rejected: doesn't match family setting; user could be traveling
  - Prop-drill timezone through components — rejected: many components need it; context is simpler

### Decision 5: Handling existing data on deployment — startup recalculation

- **Decision**: Add a `RecalculateAllNextRuns()` function that runs once on application startup (in the scheduler's Start method). It iterates all active schedules, looks up the family timezone, and recalculates `next_run_at` using the new timezone-aware logic.
- **Rationale**: Without this, the first payment after deployment fires at the old UTC midnight time (wrong time for the family's timezone). After that first execution, subsequent runs self-correct. A startup recalculation ensures immediate correctness.
- **Alternatives considered**:
  - Let schedules self-correct on next execution — rejected: one wrong-time payment per schedule is avoidable
  - Database migration — rejected: timezone calculation logic is complex; Go code is the right place
  - Manual admin endpoint — rejected: error-prone, easy to forget

### Decision 6: Fallback when no timezone configured

- **Decision**: Fall back to UTC. The migration from 013-parent-settings defaults `families.timezone` to `'America/New_York'`, so in practice all families have a timezone. But if somehow empty, use UTC.
- **Rationale**: Safe default. UTC is well-understood and unambiguous.
- **Alternatives considered**: None — UTC is the standard fallback.

## Key File Inventory

| File | Current Role | Changes Needed |
|------|-------------|----------------|
| `backend/internal/allowance/schedule_calc.go` | Computes next_run_at in UTC | Accept `*time.Location`, use it instead of `time.UTC` |
| `backend/internal/allowance/scheduler.go` | Processes due allowances | Pass timezone to `CalculateNextRunAfterExecution`; add `familyStore`; add startup recalc |
| `backend/internal/interest/scheduler.go` | Processes due interest | Pass timezone to calculation; add `familyStore`; add startup recalc |
| `backend/internal/allowance/handler.go` | Schedule CRUD handlers | Add `familyStore`; pass timezone to `CalculateNextRun` |
| `backend/internal/interest/handler.go` | Interest schedule handlers | Add `familyStore`; pass timezone to calculation |
| `backend/internal/store/schedule.go` | AllowanceSchedule store + ListDue | Modify ListDue to JOIN families, return timezone |
| `backend/internal/store/interest_schedule.go` | InterestSchedule store + ListDue | Modify ListDue to JOIN families, return timezone |
| `backend/internal/store/family.go` | Family CRUD + timezone | No changes (GetTimezone already exists) |
| `backend/main.go` | Wiring & startup | Pass familyStore to schedulers and handlers |
| `frontend/src/components/TransactionsCard.tsx` | Date display for transactions & upcoming | Use family timezone in date formatting |
| `frontend/src/components/ChildAllowanceForm.tsx` | Shows next run date | Use family timezone in `formatNextRun` |
| `frontend/src/components/InterestForm.tsx` | Shows next payout date | Use family timezone in `formatNextRun` |
| `frontend/src/pages/SettingsPage.tsx` | Timezone settings | No changes needed |
| `frontend/src/api.ts` | API client | No changes needed (getSettings already exists) |
| `frontend/src/types.ts` | Type definitions | No changes needed |
