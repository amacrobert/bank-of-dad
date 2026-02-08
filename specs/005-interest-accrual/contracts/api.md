# API Contracts: Interest Accrual

## Endpoints

### PUT /api/children/{id}/interest-rate

**Purpose**: Parent sets or updates the interest rate for a child's account.

**Auth**: Parent only (must own the child's family)

**Request**:
```json
{
  "interest_rate_bps": 500
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `interest_rate_bps` | integer | yes | 0–10000 (0%–100%) |

**Response 200**:
```json
{
  "child_id": 1,
  "interest_rate_bps": 500,
  "interest_rate_display": "5.00%"
}
```

**Response 400** (validation error):
```json
{
  "error": "Invalid interest rate",
  "message": "Interest rate must be between 0% and 100%"
}
```

**Response 403** (wrong family):
```json
{
  "error": "Forbidden"
}
```

---

### GET /api/children/{id}/balance

**Purpose**: Get child's balance (existing endpoint — enhanced response).

**Auth**: Parent or child (must be in the same family / be the child)

**Response 200** (enhanced — adds interest fields):
```json
{
  "child_id": 1,
  "first_name": "Alice",
  "balance_cents": 10000,
  "interest_rate_bps": 500,
  "interest_rate_display": "5.00%"
}
```

---

### GET /api/children/{id}/transactions

**Purpose**: Get transaction history (existing endpoint — now includes interest transactions).

**Auth**: Parent or child (must be in the same family / be the child)

**Response 200** (existing format — interest transactions included):
```json
{
  "transactions": [
    {
      "id": 42,
      "child_id": 1,
      "parent_id": 1,
      "amount_cents": 83,
      "type": "interest",
      "note": "5.00% annual rate",
      "schedule_id": null,
      "created_at": "2026-02-01T00:00:00Z"
    },
    {
      "id": 41,
      "child_id": 1,
      "parent_id": 1,
      "amount_cents": 5000,
      "type": "deposit",
      "note": "Birthday money",
      "schedule_id": null,
      "created_at": "2026-01-15T12:00:00Z"
    }
  ]
}
```

## Notes

- No new GET endpoints for interest-specific history — interest transactions appear in the existing `/transactions` endpoint with `type: "interest"`. Clients can filter by type if needed.
- The balance endpoint is enhanced to include interest rate info so the frontend can display it without an additional API call.
- Interest rate is expressed in basis points (integer) in the API to avoid floating-point issues. A human-readable `interest_rate_display` string is included for convenience.
