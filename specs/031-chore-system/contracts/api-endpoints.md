# API Contracts: Chore & Task System

**Feature Branch**: `031-chore-system`
**Created**: 2026-03-22

## Chore Management (Parent Only)

### POST /api/chores

Create a new chore and assign it to children.

**Auth**: `requireParent`

**Request**:
```json
{
  "name": "Mow the lawn",
  "description": "Front and back yard",
  "reward_cents": 500,
  "recurrence": "weekly",
  "day_of_week": 6,
  "child_ids": [1, 3]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| name | string | yes | 1-100 chars |
| description | string | no | max 500 chars |
| reward_cents | integer | yes | 0 to 99,999,999 |
| recurrence | string | yes | one_time, daily, weekly, monthly |
| day_of_week | integer | conditional | 0-6, required if weekly |
| day_of_month | integer | conditional | 1-31, required if monthly |
| child_ids | integer[] | yes | at least 1, all must belong to parent's family |

**Response** `201 Created`:
```json
{
  "chore": {
    "id": 1,
    "family_id": 1,
    "name": "Mow the lawn",
    "description": "Front and back yard",
    "reward_cents": 500,
    "recurrence": "weekly",
    "day_of_week": 6,
    "day_of_month": null,
    "is_active": true,
    "created_at": "2026-03-22T10:00:00Z",
    "updated_at": "2026-03-22T10:00:00Z",
    "assignments": [
      { "child_id": 1, "child_name": "Alice" },
      { "child_id": 3, "child_name": "Bob" }
    ]
  }
}
```

**Errors**: `400 invalid_request`, `400 invalid_name`, `400 invalid_amount`, `400 invalid_recurrence`, `400 invalid_children`, `404 child_not_found`, `403 forbidden`

---

### GET /api/chores

List all chores for the parent's family with assignment and instance status summary.

**Auth**: `requireParent`

**Response** `200 OK`:
```json
{
  "chores": [
    {
      "id": 1,
      "name": "Mow the lawn",
      "description": "Front and back yard",
      "reward_cents": 500,
      "recurrence": "weekly",
      "day_of_week": 6,
      "is_active": true,
      "created_at": "2026-03-22T10:00:00Z",
      "assignments": [
        { "child_id": 1, "child_name": "Alice" },
        { "child_id": 3, "child_name": "Bob" }
      ],
      "pending_count": 1
    }
  ]
}
```

---

### PUT /api/chores/{id}

Update a chore's name, description, reward, or recurrence. Changes apply to future instances only.

**Auth**: `requireParent`

**Request**: Same fields as POST except `child_ids` (all optional, at least one must change). Assignment changes are not supported via update — delete and recreate the chore to change assigned children.

**Response** `200 OK`: Updated chore object (same shape as POST response).

**Errors**: `400 invalid_request`, `404 not_found`, `403 forbidden`

---

### DELETE /api/chores/{id}

Delete a chore. Cancels pending instances. Preserves completed transaction history.

**Auth**: `requireParent`

**Response** `204 No Content`

**Errors**: `404 not_found`, `403 forbidden`

---

### PATCH /api/chores/{id}/deactivate

Deactivate a recurring chore (stop generating new instances).

**Auth**: `requireParent`

**Response** `200 OK`:
```json
{
  "chore": { "...": "...", "is_active": false }
}
```

---

### PATCH /api/chores/{id}/activate

Reactivate a deactivated recurring chore.

**Auth**: `requireParent`

**Response** `200 OK`:
```json
{
  "chore": { "...": "...", "is_active": true }
}
```

---

## Chore Instances

### GET /api/chores/pending

List all pending-approval instances across all children for the parent's family.

**Auth**: `requireParent`

**Response** `200 OK`:
```json
{
  "instances": [
    {
      "id": 5,
      "chore_id": 1,
      "chore_name": "Mow the lawn",
      "child_id": 1,
      "child_name": "Alice",
      "reward_cents": 500,
      "status": "pending_approval",
      "completed_at": "2026-03-22T14:30:00Z",
      "created_at": "2026-03-22T10:00:00Z"
    }
  ]
}
```

---

### POST /api/chore-instances/{id}/approve

Approve a pending chore instance. Deposits reward into child's account.

**Auth**: `requireParent`

**Response** `200 OK`:
```json
{
  "instance": {
    "id": 5,
    "status": "approved",
    "reviewed_at": "2026-03-22T15:00:00Z",
    "transaction_id": 42
  },
  "new_balance": 1500
}
```

**Errors**: `400 invalid_status` (not pending), `403 child_disabled`, `404 not_found`, `403 forbidden`

---

### POST /api/chore-instances/{id}/reject

Reject a pending chore instance. Returns it to available.

**Auth**: `requireParent`

**Request**:
```json
{
  "reason": "Yard wasn't fully mowed"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| reason | string | no | max 500 chars |

**Response** `200 OK`:
```json
{
  "instance": {
    "id": 5,
    "status": "available",
    "rejection_reason": "Yard wasn't fully mowed",
    "reviewed_at": "2026-03-22T15:00:00Z"
  }
}
```

---

## Child Endpoints

### GET /api/child/chores

List the authenticated child's chore instances grouped by status.

**Auth**: `requireAuth` (child)

**Response** `200 OK`:
```json
{
  "available": [
    {
      "id": 10,
      "chore_id": 1,
      "chore_name": "Mow the lawn",
      "chore_description": "Front and back yard",
      "reward_cents": 500,
      "rejection_reason": null,
      "period_start": "2026-03-17",
      "period_end": "2026-03-23",
      "created_at": "2026-03-17T05:00:00Z"
    }
  ],
  "pending": [
    {
      "id": 8,
      "chore_id": 2,
      "chore_name": "Clean your room",
      "reward_cents": 200,
      "completed_at": "2026-03-22T09:00:00Z"
    }
  ],
  "completed": [
    {
      "id": 5,
      "chore_id": 1,
      "chore_name": "Mow the lawn",
      "reward_cents": 500,
      "status": "approved",
      "reviewed_at": "2026-03-15T15:00:00Z"
    }
  ]
}
```

---

### POST /api/child/chores/{id}/complete

Mark a chore instance as complete (moves to pending_approval).

**Auth**: `requireAuth` (child)

**Response** `200 OK`:
```json
{
  "instance": {
    "id": 10,
    "status": "pending_approval",
    "completed_at": "2026-03-22T14:30:00Z"
  }
}
```

**Errors**: `400 invalid_status` (not available), `404 not_found`, `403 forbidden`

---

### GET /api/child/chores/earnings

Summary of chore earnings for the authenticated child.

**Auth**: `requireAuth` (child)

**Response** `200 OK`:
```json
{
  "total_earned_cents": 2500,
  "chores_completed": 7,
  "recent": [
    {
      "chore_name": "Mow the lawn",
      "reward_cents": 500,
      "approved_at": "2026-03-22T15:00:00Z"
    }
  ]
}
```
