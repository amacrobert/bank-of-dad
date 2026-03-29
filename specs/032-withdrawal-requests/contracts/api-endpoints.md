# API Contracts: Withdrawal Requests

**Feature**: 032-withdrawal-requests | **Date**: 2026-03-28

## Child Endpoints

### POST /api/child/withdrawal-requests

Submit a new withdrawal request. Requires child authentication.

**Request**:
```json
{
    "amount_cents": 2000,
    "reason": "Birthday present for mom"
}
```

**Validation**:
- `amount_cents`: required, integer, > 0, <= 99999999
- `reason`: required, string, 1–500 characters
- Child must not have an existing pending request
- Child's available balance must be >= amount_cents
- Child's account must not be disabled

**Success Response** (201 Created):
```json
{
    "withdrawal_request": {
        "id": 1,
        "child_id": 5,
        "family_id": 2,
        "amount_cents": 2000,
        "reason": "Birthday present for mom",
        "status": "pending",
        "created_at": "2026-03-28T12:00:00Z",
        "updated_at": "2026-03-28T12:00:00Z"
    }
}
```

**Error Responses**:
- `400` — Invalid input (amount, reason validation)
- `409` — Child already has a pending request
- `422` — Insufficient available balance
- `403` — Account is disabled

---

### GET /api/child/withdrawal-requests

List the authenticated child's withdrawal requests (all statuses). Requires child authentication.

**Query Parameters**:
- `status` (optional): Filter by status (`pending`, `approved`, `denied`, `cancelled`)

**Success Response** (200 OK):
```json
{
    "withdrawal_requests": [
        {
            "id": 1,
            "child_id": 5,
            "family_id": 2,
            "amount_cents": 2000,
            "reason": "Birthday present for mom",
            "status": "pending",
            "created_at": "2026-03-28T12:00:00Z",
            "updated_at": "2026-03-28T12:00:00Z"
        }
    ]
}
```

---

### POST /api/child/withdrawal-requests/{id}/cancel

Cancel a pending withdrawal request. Requires child authentication. Child must own the request.

**Request**: Empty body.

**Success Response** (200 OK):
```json
{
    "withdrawal_request": {
        "id": 1,
        "status": "cancelled",
        "updated_at": "2026-03-28T12:05:00Z"
    }
}
```

**Error Responses**:
- `404` — Request not found or not owned by child
- `409` — Request is not in pending status

---

## Parent Endpoints

### GET /api/withdrawal-requests

List withdrawal requests for the parent's family. Requires parent authentication.

**Query Parameters**:
- `status` (optional): Filter by status (`pending`, `approved`, `denied`, `cancelled`)
- `child_id` (optional): Filter by specific child

**Success Response** (200 OK):
```json
{
    "withdrawal_requests": [
        {
            "id": 1,
            "child_id": 5,
            "child_name": "Emma",
            "family_id": 2,
            "amount_cents": 2000,
            "reason": "Birthday present for mom",
            "status": "pending",
            "created_at": "2026-03-28T12:00:00Z",
            "updated_at": "2026-03-28T12:00:00Z"
        }
    ]
}
```

Note: `child_name` is included to avoid extra lookups on the frontend.

---

### POST /api/withdrawal-requests/{id}/approve

Approve a pending withdrawal request. Requires parent authentication. Parent must be in the same family.

**Request**:
```json
{
    "confirm_goal_impact": false
}
```

**Validation**:
- Request must be in `pending` status
- Child's available balance must be >= amount_cents at approval time
- Child's account must not be disabled
- If withdrawal impacts savings goals, returns 409 unless `confirm_goal_impact: true`

**Success Response** (200 OK):
```json
{
    "withdrawal_request": {
        "id": 1,
        "status": "approved",
        "reviewed_by_parent_id": 3,
        "reviewed_at": "2026-03-28T14:00:00Z",
        "transaction_id": 42,
        "updated_at": "2026-03-28T14:00:00Z"
    },
    "new_balance_cents": 3000
}
```

**Error Responses**:
- `404` — Request not found or not in parent's family
- `409` — Request not pending, or goal impact requires confirmation
- `422` — Insufficient available balance

---

### POST /api/withdrawal-requests/{id}/deny

Deny a pending withdrawal request. Requires parent authentication. Parent must be in the same family.

**Request**:
```json
{
    "reason": "Save up a bit more first"
}
```

**Validation**:
- Request must be in `pending` status
- `reason`: optional, string, max 500 characters

**Success Response** (200 OK):
```json
{
    "withdrawal_request": {
        "id": 1,
        "status": "denied",
        "denial_reason": "Save up a bit more first",
        "reviewed_by_parent_id": 3,
        "reviewed_at": "2026-03-28T14:00:00Z",
        "updated_at": "2026-03-28T14:00:00Z"
    }
}
```

**Error Responses**:
- `404` — Request not found or not in parent's family
- `409` — Request not in pending status

---

### GET /api/withdrawal-requests/pending/count

Get count of pending withdrawal requests for the parent's family. Used for badge/indicator display.

**Success Response** (200 OK):
```json
{
    "count": 2
}
```
