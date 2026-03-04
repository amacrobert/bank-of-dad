# API Contracts: Savings Goals

**Branch**: `025-savings-goals` | **Date**: 2026-03-02

## Endpoints

### GET /api/children/{id}/savings-goals

List all savings goals for a child. Returns active goals first (sorted by creation date), then completed goals (sorted by completion date, most recent first).

**Auth**: `requireAuth` — parent (any child in family) or child (own goals only)

**Response 200**:
```json
{
  "goals": [
    {
      "id": 1,
      "child_id": 1,
      "name": "New Skateboard",
      "target_cents": 4500,
      "saved_cents": 2000,
      "emoji": "🛹",
      "target_date": "2026-06-01",
      "status": "active",
      "completed_at": null,
      "created_at": "2026-03-02T10:00:00Z",
      "updated_at": "2026-03-02T10:00:00Z"
    }
  ],
  "available_balance_cents": 3000,
  "total_saved_cents": 2000
}
```

**Error Responses**:
- `403 Forbidden` — user does not have access to this child
- `404 Not Found` — child not found

---

### POST /api/children/{id}/savings-goals

Create a new savings goal for a child.

**Auth**: `requireAuth` — child only (own goals)

**Request Body**:
```json
{
  "name": "New Skateboard",
  "target_cents": 4500,
  "emoji": "🛹",
  "target_date": "2026-06-01"
}
```

| Field        | Type    | Required | Constraints                |
| ------------ | ------- | -------- | -------------------------- |
| name         | string  | yes      | 1–50 characters            |
| target_cents | integer | yes      | > 0, ≤ 99999999            |
| emoji        | string  | no       | single emoji character     |
| target_date  | string  | no       | ISO 8601 date (YYYY-MM-DD), must be in the future |

**Response 201**:
```json
{
  "id": 1,
  "child_id": 1,
  "name": "New Skateboard",
  "target_cents": 4500,
  "saved_cents": 0,
  "emoji": "🛹",
  "target_date": "2026-06-01",
  "status": "active",
  "completed_at": null,
  "created_at": "2026-03-02T10:00:00Z",
  "updated_at": "2026-03-02T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request` — validation errors (missing name, invalid amount, etc.)
- `403 Forbidden` — only children can create their own goals
- `409 Conflict` — child already has 5 active goals

---

### PUT /api/children/{id}/savings-goals/{goalId}

Update an active savings goal's details.

**Auth**: `requireAuth` — child only (own goals)

**Request Body**:
```json
{
  "name": "Pro Skateboard",
  "target_cents": 6000,
  "emoji": "🛹",
  "target_date": "2026-07-01"
}
```

All fields are optional — only provided fields are updated.

**Response 200**: Updated goal object (same shape as POST response)

**Error Responses**:
- `400 Bad Request` — validation errors
- `403 Forbidden` — not the goal owner
- `404 Not Found` — goal not found or not active

**Special behavior**: If `target_cents` is reduced to ≤ `saved_cents`, the goal is automatically marked as completed and `completed_at` is set.

---

### DELETE /api/children/{id}/savings-goals/{goalId}

Delete an active savings goal. Allocated funds return to available balance.

**Auth**: `requireAuth` — child only (own goals)

**Response 200**:
```json
{
  "released_cents": 2000,
  "available_balance_cents": 5000
}
```

**Error Responses**:
- `403 Forbidden` — not the goal owner
- `404 Not Found` — goal not found or not active

---

### POST /api/children/{id}/savings-goals/{goalId}/allocate

Allocate or de-allocate funds to/from a savings goal.

**Auth**: `requireAuth` — child only (own goals)

**Request Body**:
```json
{
  "amount_cents": 1000
}
```

| Field        | Type    | Required | Constraints                                              |
| ------------ | ------- | -------- | -------------------------------------------------------- |
| amount_cents | integer | yes      | Non-zero. Positive = allocate, negative = de-allocate    |

**Response 200**:
```json
{
  "goal": {
    "id": 1,
    "child_id": 1,
    "name": "New Skateboard",
    "target_cents": 4500,
    "saved_cents": 3000,
    "emoji": "🛹",
    "target_date": "2026-06-01",
    "status": "active",
    "completed_at": null,
    "created_at": "2026-03-02T10:00:00Z",
    "updated_at": "2026-03-02T12:00:00Z"
  },
  "available_balance_cents": 2000,
  "completed": false
}
```

When the allocation causes `saved_cents >= target_cents`:
```json
{
  "goal": {
    "status": "completed",
    "completed_at": "2026-03-02T12:00:00Z"
  },
  "available_balance_cents": 500,
  "completed": true
}
```

**Error Responses**:
- `400 Bad Request` — amount is 0, allocation exceeds available balance, or de-allocation exceeds saved amount
- `403 Forbidden` — not the goal owner
- `404 Not Found` — goal not found or not active

---

### GET /api/children/{id}/savings-goals/{goalId}/allocations

List allocation history for a specific goal.

**Auth**: `requireAuth` — parent (family member) or child (own goals)

**Response 200**:
```json
{
  "allocations": [
    {
      "id": 1,
      "goal_id": 1,
      "amount_cents": 1000,
      "created_at": "2026-03-02T10:00:00Z"
    },
    {
      "id": 2,
      "goal_id": 1,
      "amount_cents": -500,
      "created_at": "2026-03-03T14:00:00Z"
    }
  ]
}
```

---

## Modified Endpoints

### POST /api/children/{id}/withdraw (existing)

**Change**: Before processing the withdrawal, check if the resulting balance would fall below total goal allocations. If so, include a warning in the response requiring confirmation.

**New request field** (optional):
```json
{
  "amount_cents": 5000,
  "note": "Birthday gift",
  "confirm_goal_impact": true
}
```

**New error response** (when goals would be impacted and `confirm_goal_impact` is not `true`):
```json
{
  "error": "goal_impact_warning",
  "message": "This withdrawal will reduce savings goals allocations.",
  "affected_goals": [
    { "id": 1, "name": "New Skateboard", "current_saved_cents": 2000, "new_saved_cents": 1200 },
    { "id": 2, "name": "Video Game", "current_saved_cents": 1500, "new_saved_cents": 900 }
  ],
  "total_released_cents": 1400
}
```

The parent must re-submit with `confirm_goal_impact: true` to proceed. When confirmed, goal allocations are reduced proportionally.

### GET /api/children/{id}/balance (existing)

**New response fields**:
```json
{
  "child_id": 1,
  "first_name": "Emma",
  "balance_cents": 5000,
  "available_balance_cents": 3000,
  "total_saved_cents": 2000,
  "active_goals_count": 2,
  "interest_rate_bps": 500,
  "interest_rate_display": "5.00%",
  "next_interest_at": "2026-04-01T00:00:00Z"
}
```

Added: `available_balance_cents`, `total_saved_cents`, `active_goals_count`
