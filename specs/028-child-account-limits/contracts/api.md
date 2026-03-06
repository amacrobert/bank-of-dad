# API Contracts: Free Tier Child Account Limits

## Modified Endpoints

### POST /children — Create Child

**Change**: Remove the hard block at 2 children. Instead, check account type and set `is_disabled` on children beyond the free tier limit.

**Request**: Unchanged.

**Response**: Add `is_disabled` field to the child object in the response.

```json
{
  "id": 4,
  "first_name": "Charlie",
  "is_locked": false,
  "is_disabled": true,
  "balance_cents": 0,
  "avatar": "🦊",
  "created_at": "2026-03-06T12:00:00Z",
  "login_url": "/family-login/smith-family"
}
```

### GET /children — List Children

**Change**: Add `is_disabled` field to each child object. All children are returned (including disabled ones).

```json
{
  "children": [
    { "id": 1, "first_name": "Alice", "is_disabled": false, "is_locked": false, "balance_cents": 5000, "avatar": "🐱" },
    { "id": 2, "first_name": "Bob", "is_disabled": false, "is_locked": false, "balance_cents": 3000, "avatar": "🐶" },
    { "id": 3, "first_name": "Charlie", "is_disabled": true, "is_locked": false, "balance_cents": 0, "avatar": "🦊" }
  ]
}
```

### POST /children/:id/deposit — Deposit

**Change**: Returns 403 if child `is_disabled`.

```json
{
  "error": "Account disabled",
  "message": "This account is disabled. Upgrade to Plus to enable all children."
}
```

### POST /children/:id/withdraw — Withdraw

**Change**: Returns 403 if child `is_disabled`.

Same error response as deposit.

### POST /auth/child/login — Child Login

**Change**: Returns 403 if child `is_disabled`.

```json
{
  "error": "Account disabled",
  "message": "This account is disabled. Ask your parent to upgrade to Plus."
}
```

### GET /families/:slug/children — Family Children (Public Login Page)

**Change**: Add `is_disabled` field. Disabled children should not appear in the login child selector.

## No Changes

These endpoints remain unchanged:
- `PUT /children/:id/password` — Still works for disabled children (settings page)
- `PUT /children/:id/name` — Still works for disabled children (settings page)
- `DELETE /children/:id` — Still works for disabled children (settings page)
- `GET /children/:id/balance` — Not called for disabled children in practice, but no backend guard needed
- `GET /children/:id/transactions` — Same as above
