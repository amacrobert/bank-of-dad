# API Contracts: Child Visual Themes

**Feature**: 017-child-visual-themes
**Date**: 2026-02-18

## Modified Endpoints

### GET /api/auth/me (child response — modified)

Returns the child's theme in the existing response payload.

**Auth**: Any authenticated user (existing)

**Response** (child user, 200):
```json
{
    "user_type": "child",
    "user_id": 123,
    "family_id": 456,
    "first_name": "Alice",
    "family_slug": "smith-family",
    "avatar": "fox",
    "theme": "piggybank"
}
```

**Notes**:
- `theme` is `null` or absent when no theme has been set (treated as "sapling" by frontend).
- `theme` values: `"sapling"`, `"piggybank"`, `"rainbow"`, or `null`.
- Parent response is unchanged.

---

## New Endpoints

### PUT /api/child/settings/theme

Updates the authenticated child's theme preference.

**Auth**: `requireAuth` — handler validates `user_type === "child"`

**Request**:
```json
{
    "theme": "rainbow"
}
```

| Field   | Type   | Required | Validation                                          |
|---------|--------|----------|-----------------------------------------------------|
| `theme` | string | Yes      | Must be one of: `"sapling"`, `"piggybank"`, `"rainbow"` |

**Response** (200):
```json
{
    "message": "Theme updated",
    "theme": "rainbow"
}
```

**Error Responses**:

| Status | Condition                      | Body                                          |
|--------|--------------------------------|-----------------------------------------------|
| 400    | Invalid or missing theme value | `{"error": "Invalid theme. Must be one of: sapling, piggybank, rainbow"}` |
| 401    | Not authenticated              | `{"error": "Unauthorized"}`                   |
| 403    | User is not a child            | `{"error": "Only child users can set themes"}` |
| 500    | Database error                 | `{"error": "Failed to update theme"}`         |

---

## Route Registration

```
mux.Handle("PUT /api/child/settings/theme", requireAuth(http.HandlerFunc(childSettingsHandlers.HandleUpdateTheme)))
```

**Note**: Uses `requireAuth` (not `requireParent`) because children set their own theme.
