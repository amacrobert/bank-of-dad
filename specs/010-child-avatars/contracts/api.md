# API Contracts: Child Avatars

**Feature**: 010-child-avatars
**Date**: 2026-02-11

## Modified Endpoints

### POST /api/children (Create Child)

**Auth**: Parent (requireParent middleware)

**Request body** (changed — new optional field):

```json
{
  "first_name": "Alice",
  "password": "secret123",
  "avatar": "🌻"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| first_name | string | Yes | Existing |
| password | string | Yes | Existing |
| avatar | string | No | New — emoji string, omit or null for no avatar |

**Response** `201 Created` (changed — new field):

```json
{
  "id": 1,
  "first_name": "Alice",
  "family_slug": "smith-family",
  "login_url": "/smith-family",
  "avatar": "🌻"
}
```

| Field | Type | Notes |
|-------|------|-------|
| avatar | string \| null | New — null if no avatar set |

---

### GET /api/children (List Children)

**Auth**: Parent (requireParent middleware)

**Response** `200 OK` (changed — new field per child):

```json
{
  "children": [
    {
      "id": 1,
      "first_name": "Alice",
      "is_locked": false,
      "balance_cents": 5000,
      "created_at": "2026-02-11T10:00:00Z",
      "avatar": "🌻"
    },
    {
      "id": 2,
      "first_name": "Bob",
      "is_locked": false,
      "balance_cents": 3000,
      "created_at": "2026-02-11T10:00:00Z",
      "avatar": null
    }
  ]
}
```

| Field | Type | Notes |
|-------|------|-------|
| avatar | string \| null | New — null if no avatar set |

---

### PUT /api/children/{id}/name (Update Name and Avatar)

**Auth**: Parent (requireParent middleware)

**Request body** (changed — new optional field):

```json
{
  "first_name": "Alice",
  "avatar": "🦋"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| first_name | string | Yes | Existing |
| avatar | string \| null | No | New — set to emoji string, null to clear, omit to leave unchanged |

**Response** `200 OK` (changed — new field):

```json
{
  "message": "Name updated",
  "first_name": "Alice",
  "avatar": "🦋"
}
```

| Field | Type | Notes |
|-------|------|-------|
| avatar | string \| null | New — reflects current avatar after update |

---

## Unchanged Endpoints

The following endpoints are NOT modified:

- `GET /api/children/{id}/balance` — returns balance data, not child profile
- `GET /api/children/{id}/transactions` — transaction history
- `PUT /api/children/{id}/password` — password reset
- `DELETE /api/children/{id}` — delete child (cascade handles avatar implicitly)
