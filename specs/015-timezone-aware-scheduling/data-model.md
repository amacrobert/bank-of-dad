# Data Model: Timezone-Aware Scheduling

**Feature**: 015-timezone-aware-scheduling
**Date**: 2026-02-17

## Schema Changes

**No new tables or columns required.** The `families.timezone` column (TEXT, NOT NULL, DEFAULT 'America/New_York') already exists from feature 013-parent-settings.

## Behavioral Changes to Existing Data

### `allowance_schedules.next_run_at` (TIMESTAMPTZ)

**Before**: Stores midnight UTC on the target day (e.g., Wednesday 00:00:00 UTC).

**After**: Stores midnight in the family's configured timezone, expressed as UTC (e.g., Wednesday 00:00:00 EST = Wednesday 05:00:00 UTC for America/New_York).

The column type and semantics don't change — it's still "when should this schedule next fire" as a UTC timestamp. What changes is the value stored: it now represents midnight in the family's local timezone rather than midnight UTC.

### `interest_schedules.next_run_at` (TIMESTAMPTZ)

Same change as `allowance_schedules.next_run_at`.

### `transactions.created_at` (TIMESTAMPTZ)

**No change.** Transactions continue to record the actual UTC timestamp when they were created. The change is purely in how the frontend displays this date (using the family's timezone rather than the browser's timezone).

## Backend Struct Changes

### New: `DueAllowanceSchedule` (store package)

Returned by the modified `ListDue()` query that JOINs with families:

```
DueAllowanceSchedule {
    AllowanceSchedule    (embedded — all existing fields)
    FamilyTimezone string   // IANA timezone from families.timezone
}
```

### New: `DueInterestSchedule` (store package)

Same pattern for interest schedules:

```
DueInterestSchedule {
    InterestSchedule     (embedded — all existing fields)
    FamilyTimezone string   // IANA timezone from families.timezone
}
```

### Modified: `CalculateNextRun` signature

```
Before: CalculateNextRun(sched *AllowanceSchedule, after time.Time) time.Time
After:  CalculateNextRun(sched *AllowanceSchedule, after time.Time, loc *time.Location) time.Time
```

Same change for `CalculateNextRunAfterExecution`.

### Modified: Scheduler struct dependencies

Both `allowance.Scheduler` and `interest.Scheduler` gain a `familyStore *store.FamilyStore` field (used for startup recalculation and as a fallback).

### Modified: Handler struct dependencies

Both `allowance.Handler` and `interest.Handler` gain a `familyStore *store.FamilyStore` field (used to look up timezone when creating/updating/resuming schedules).

## Frontend Model Changes

### New: `TimezoneContext`

A React Context providing the family's timezone string to all components:

```
TimezoneContext {
    timezone: string     // IANA timezone identifier, e.g., "America/New_York"
    loading: boolean     // True while fetching settings
}
```

Populated by a single call to `GET /api/settings` on app mount.

## Entity Relationships (unchanged)

```
Family (1) ──── (*) Parent
   |
   └── timezone: string (IANA)
   |
Family (1) ──── (*) Child
                    |
                    ├── (*) AllowanceSchedule  →  next_run_at uses family.timezone
                    ├── (*) InterestSchedule   →  next_run_at uses family.timezone
                    └── (*) Transaction        →  created_at displayed in family.timezone
```

## Data Migration

No SQL migration needed. On application startup, the scheduler recalculates all active schedules' `next_run_at` values using the timezone-aware logic. This is idempotent and can safely run on every startup.
