# API Contract: Settings

**Feature Branch**: `013-parent-settings`
**Date**: 2026-02-16

## Endpoints

### GET /api/settings

Retrieve the current family settings for the authenticated parent.

**Authentication**: Required (parent only — `RequireParent` middleware)

**Request**: No body.

**Response 200**:
```json
{
  "timezone": "America/New_York"
}
```

**Response 401**: Unauthenticated
```json
{
  "error": "Unauthorized"
}
```

**Response 403**: Not a parent
```json
{
  "error": "Forbidden"
}
```

**Response 500**: Internal error
```json
{
  "error": "Internal server error"
}
```

---

### PUT /api/settings/timezone

Update the family timezone.

**Authentication**: Required (parent only — `RequireParent` middleware)

**Request body**:
```json
{
  "timezone": "America/Chicago"
}
```

| Field    | Type   | Required | Validation                              |
|----------|--------|----------|-----------------------------------------|
| timezone | string | Yes      | Must be a valid IANA timezone identifier |

**Response 200**: Success
```json
{
  "message": "Timezone updated",
  "timezone": "America/Chicago"
}
```

**Response 400**: Invalid timezone
```json
{
  "error": "Invalid timezone",
  "message": "\"Fake/Timezone\" is not a valid IANA timezone identifier"
}
```

**Response 401**: Unauthenticated
```json
{
  "error": "Unauthorized"
}
```

**Response 403**: Not a parent
```json
{
  "error": "Forbidden"
}
```

**Response 500**: Internal error
```json
{
  "error": "Internal server error"
}
```
