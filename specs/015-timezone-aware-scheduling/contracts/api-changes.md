# API Contract Changes: Timezone-Aware Scheduling

**Feature**: 015-timezone-aware-scheduling
**Date**: 2026-02-17

## No New Endpoints

This feature does not add or remove any API endpoints. All changes are behavioral — the same endpoints return the same response shapes, but with timezone-corrected values.

## Changed Behavior

### All schedule-related endpoints

The `next_run_at` field in responses now represents midnight in the family's configured timezone (expressed as UTC), instead of midnight UTC.

**Affected endpoints:**
- `POST /api/schedules` — response `next_run_at`
- `PUT /api/schedules/{id}` — response `next_run_at`
- `POST /api/schedules/{id}/resume` — response `next_run_at`
- `GET /api/schedules` — response items `next_run_at`
- `GET /api/schedules/{id}` — response `next_run_at`
- `GET /api/children/{childId}/allowance` — response `next_run_at`
- `PUT /api/children/{childId}/allowance` — response `next_run_at`
- `POST /api/children/{childId}/allowance/resume` — response `next_run_at`
- `GET /api/children/{childId}/upcoming-allowances` — response items `next_date`
- `PUT /api/children/{id}/interest` — response `schedule.next_run_at`
- `GET /api/children/{childId}/interest-schedule` — response `next_run_at`
- `GET /api/children/{id}/balance` — response `next_interest_at`

**Example change:**

Family timezone: `America/New_York`, schedule: weekly on Wednesday.

Before (midnight UTC):
```json
{ "next_run_at": "2026-02-18T00:00:00Z" }
```

After (midnight EST = 5am UTC):
```json
{ "next_run_at": "2026-02-18T05:00:00Z" }
```

The response format (RFC3339 string) is unchanged. Only the timestamp value changes.

### `GET /api/settings`

**No change.** Already returns `{ "timezone": "America/New_York" }`. The frontend will now use this value for date formatting.

### Transaction endpoints

**No change to response values.** `created_at` continues to be the actual UTC timestamp of when the transaction was created. The frontend changes how it displays this timestamp (using family timezone instead of browser timezone).
