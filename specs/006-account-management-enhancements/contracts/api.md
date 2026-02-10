# API Contracts: Account Management Enhancements

## Existing Endpoints (unchanged)

These endpoints already exist and require no changes:

- `GET /api/children/{id}/transactions` — Returns transaction list (already supports parent access)
- `PUT /api/children/{id}/interest-rate` — Sets interest rate (already exists from 005)

## Modified Endpoints

### GET /api/children/{id}/balance

**Change**: Add `next_interest_at` field to response.

**Auth**: Any authenticated user (parent sees family children, child sees own)

**Response** (200):
```json
{
  "child_id": 1,
  "first_name": "Emma",
  "balance_cents": 10000,
  "interest_rate_bps": 500,
  "interest_rate_display": "5.00%",
  "next_interest_at": "2026-03-01T00:00:00Z"
}
```

`next_interest_at` is `null` if no interest schedule is configured.

---

## New Endpoints

### GET /api/children/{childId}/allowance

Get the single allowance schedule for a child.

**Auth**: Parent (same family) or Child (own data)

**Response** (200 — schedule exists):
```json
{
  "id": 1,
  "child_id": 1,
  "parent_id": 1,
  "amount_cents": 1000,
  "frequency": "weekly",
  "day_of_week": 5,
  "day_of_month": null,
  "note": "Weekly allowance",
  "status": "active",
  "next_run_at": "2026-02-14T00:00:00Z",
  "created_at": "2026-02-01T00:00:00Z",
  "updated_at": "2026-02-01T00:00:00Z"
}
```

**Response** (200 — no schedule):
```json
null
```

**Response** (403):
```json
{
  "error": "forbidden",
  "message": "You do not have permission to view this child's allowance."
}
```

---

### PUT /api/children/{childId}/allowance

Create or update the allowance schedule for a child. If no schedule exists, creates one. If one exists, updates it.

**Auth**: Parent (same family)

**Request**:
```json
{
  "amount_cents": 1000,
  "frequency": "weekly",
  "day_of_week": 5,
  "note": "Weekly allowance"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| amount_cents | integer | yes | 1 to 99999999 |
| frequency | string | yes | "weekly", "biweekly", "monthly" |
| day_of_week | integer | conditional | 0-6, required for weekly/biweekly |
| day_of_month | integer | conditional | 1-31, required for monthly |
| note | string | no | max 500 chars |

**Response** (200):
```json
{
  "id": 1,
  "child_id": 1,
  "parent_id": 1,
  "amount_cents": 1000,
  "frequency": "weekly",
  "day_of_week": 5,
  "day_of_month": null,
  "note": "Weekly allowance",
  "status": "active",
  "next_run_at": "2026-02-14T00:00:00Z",
  "created_at": "2026-02-01T00:00:00Z",
  "updated_at": "2026-02-09T00:00:00Z"
}
```

**Response** (400):
```json
{
  "error": "invalid_amount",
  "message": "Amount must be between 1 cent and $999,999.99."
}
```

---

### DELETE /api/children/{childId}/allowance

Remove the allowance schedule for a child.

**Auth**: Parent (same family)

**Response** (204): No content

**Response** (404):
```json
{
  "error": "not_found",
  "message": "No allowance schedule found for this child."
}
```

---

### POST /api/children/{childId}/allowance/pause

Pause the child's allowance schedule.

**Auth**: Parent (same family)

**Response** (200): Returns updated schedule (same shape as GET)

**Response** (404):
```json
{
  "error": "not_found",
  "message": "No allowance schedule found for this child."
}
```

---

### POST /api/children/{childId}/allowance/resume

Resume the child's allowance schedule.

**Auth**: Parent (same family)

**Response** (200): Returns updated schedule with recalculated `next_run_at`

---

### PUT /api/children/{childId}/interest-schedule

Create or update the interest accrual schedule for a child.

**Auth**: Parent (same family)

**Request**:
```json
{
  "frequency": "monthly",
  "day_of_month": 1
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| frequency | string | yes | "weekly", "biweekly", "monthly" |
| day_of_week | integer | conditional | 0-6, required for weekly/biweekly |
| day_of_month | integer | conditional | 1-31, required for monthly |

**Response** (200):
```json
{
  "id": 1,
  "child_id": 1,
  "frequency": "monthly",
  "day_of_month": 1,
  "status": "active",
  "next_run_at": "2026-03-01T00:00:00Z"
}
```

---

### GET /api/children/{childId}/interest-schedule

Get the interest accrual schedule for a child.

**Auth**: Any authenticated user (parent same family, child own data)

**Response** (200 — exists):
```json
{
  "id": 1,
  "child_id": 1,
  "frequency": "monthly",
  "day_of_month": 1,
  "day_of_week": null,
  "status": "active",
  "next_run_at": "2026-03-01T00:00:00Z"
}
```

**Response** (200 — none):
```json
null
```

---

### DELETE /api/children/{childId}/interest-schedule

Remove the interest accrual schedule. Interest will no longer be automatically credited.

**Auth**: Parent (same family)

**Response** (204): No content

---

## Endpoint Summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /api/children/{id}/balance | auth | Enhanced with next_interest_at |
| GET | /api/children/{childId}/allowance | auth | Get child's allowance |
| PUT | /api/children/{childId}/allowance | parent | Create/update allowance |
| DELETE | /api/children/{childId}/allowance | parent | Remove allowance |
| POST | /api/children/{childId}/allowance/pause | parent | Pause allowance |
| POST | /api/children/{childId}/allowance/resume | parent | Resume allowance |
| PUT | /api/children/{childId}/interest-schedule | parent | Create/update interest schedule |
| GET | /api/children/{childId}/interest-schedule | auth | Get interest schedule |
| DELETE | /api/children/{childId}/interest-schedule | parent | Remove interest schedule |
