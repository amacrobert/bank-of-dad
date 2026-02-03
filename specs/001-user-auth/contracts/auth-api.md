# API Contracts: Authentication

**Feature**: 001-user-auth
**Base Path**: `/api`

## Parent Authentication (Google OAuth)

### GET /api/auth/google/login

Initiates Google OAuth flow. Redirects browser to Google's consent screen.

**User Story**: US1, US2

**Request**: No body. Browser navigation.

**Response**: HTTP 302 redirect to Google OAuth consent URL with state parameter.

**Headers Set**:
- `Set-Cookie: oauth_state={random}; HttpOnly; Secure; SameSite=Lax; Max-Age=600`

---

### GET /api/auth/google/callback?code={code}&state={state}

Handles Google OAuth callback. Exchanges code for token, creates or finds parent, creates session.

**User Story**: US1, US2

**Query Parameters**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| code | string | yes | Authorization code from Google |
| state | string | yes | CSRF state parameter (must match cookie) |

**Success Response (existing parent)**: HTTP 302 redirect to `/dashboard`
- Sets session cookie

**Success Response (new parent, needs slug)**: HTTP 302 redirect to `/setup`
- Sets temporary session cookie (parent created but no family yet)

**Error Responses**:
| Status | Condition |
|--------|-----------|
| 400 | Missing or invalid state parameter |
| 401 | Google token exchange failed |
| 500 | Internal server error |

---

### POST /api/auth/logout

Ends the current session.

**User Story**: US2, US4

**Request**: No body. Session cookie required.

**Response**:
```json
{ "message": "Logged out" }
```
Status: 200

**Headers Set**:
- `Set-Cookie: session=; HttpOnly; Secure; SameSite=Lax; Max-Age=0` (clear cookie)

**Error Responses**:
| Status | Condition |
|--------|-----------|
| 401 | No valid session |

---

### GET /api/auth/me

Returns the current user's identity and role.

**User Story**: US2, US4, US6

**Request**: No body. Session cookie required.

**Response (parent)**:
```json
{
  "user_type": "parent",
  "user_id": 1,
  "family_id": 1,
  "display_name": "Jane Smith",
  "email": "jane@example.com",
  "family_slug": "smith-family"
}
```

**Response (child)**:
```json
{
  "user_type": "child",
  "user_id": 5,
  "family_id": 1,
  "first_name": "Tommy",
  "family_slug": "smith-family"
}
```

Status: 200

**Error Responses**:
| Status | Condition |
|--------|-----------|
| 401 | No valid session or session expired |

---

## Child Authentication

### POST /api/auth/child/login

Authenticates a child via family slug, first name, and password.

**User Story**: US4

**Request**:
```json
{
  "family_slug": "smith-family",
  "first_name": "Tommy",
  "password": "secret123"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| family_slug | string | yes | Must match existing family |
| first_name | string | yes | Must match child in family |
| password | string | yes | Compared against bcrypt hash |

**Response (success)**:
```json
{
  "user_type": "child",
  "first_name": "Tommy",
  "family_slug": "smith-family"
}
```
Status: 200
Sets session cookie (24-hour TTL).

**Error Responses**:
| Status | Body | Condition |
|--------|------|-----------|
| 401 | `{ "error": "Invalid credentials", "message": "Hmm, that didn't work. Try again or ask your parent for help!" }` | Wrong name or password |
| 403 | `{ "error": "Account locked", "message": "Your account is locked. Ask your parent to help you reset your password." }` | Account locked after 5 failures |
| 404 | `{ "error": "Family not found" }` | No family with that slug |

---

## Family Management

### POST /api/families

Creates a new family with the given slug. Called during parent registration setup.

**User Story**: US1

**Auth**: Parent session required (parent must exist but not yet have a family)

**Request**:
```json
{
  "slug": "smith-family"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| slug | string | yes | 3-30 chars, lowercase alphanumeric + hyphens, unique |

**Response (success)**:
```json
{
  "id": 1,
  "slug": "smith-family"
}
```
Status: 201

**Error Responses**:
| Status | Body | Condition |
|--------|------|-----------|
| 400 | `{ "error": "Invalid slug format", "message": "..." }` | Slug fails validation (FR-003) |
| 409 | `{ "error": "Slug taken", "suggestions": ["smith-family-1", "the-smiths"] }` | Slug already exists (FR-004) |
| 401 | `{ "error": "Unauthorized" }` | No valid parent session |

---

### GET /api/families/:slug

Checks if a family exists by slug. Public endpoint (used for child login page).

**User Story**: US4

**Auth**: None required

**Response (exists)**:
```json
{
  "slug": "smith-family",
  "exists": true
}
```
Status: 200

**Response (not found)**:
```json
{
  "slug": "nonexistent",
  "exists": false
}
```
Status: 200

---

### GET /api/families/check-slug?slug={slug}

Validates a slug during registration. Returns availability and suggestions.

**User Story**: US1

**Auth**: Parent session required

**Query Parameters**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| slug | string | yes | Slug to check |

**Response**:
```json
{
  "slug": "smith-family",
  "available": false,
  "valid": true,
  "suggestions": ["smith-family-1", "the-smiths", "smith-bank"]
}
```
Status: 200

---

## Child Account Management

### POST /api/children

Creates a new child account in the parent's family.

**User Story**: US3

**Auth**: Parent session required

**Request**:
```json
{
  "first_name": "Tommy",
  "password": "secret123"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| first_name | string | yes | Unique within family (FR-007) |
| password | string | yes | Minimum 6 characters (FR-006) |

**Response (success)**:
```json
{
  "id": 5,
  "first_name": "Tommy",
  "family_slug": "smith-family",
  "login_url": "/smith-family"
}
```
Status: 201

**Error Responses**:
| Status | Body | Condition |
|--------|------|-----------|
| 400 | `{ "error": "Password too short", "message": "Password must be at least 6 characters." }` | FR-006 violation |
| 409 | `{ "error": "Name taken", "message": "A child named Tommy already exists in your family." }` | FR-007 violation |
| 401 | `{ "error": "Unauthorized" }` | No valid parent session |

---

### GET /api/children

Lists all children in the parent's family.

**User Story**: US3, US5

**Auth**: Parent session required

**Response**:
```json
{
  "children": [
    {
      "id": 5,
      "first_name": "Tommy",
      "is_locked": false,
      "created_at": "2026-01-29T10:00:00Z"
    },
    {
      "id": 6,
      "first_name": "Sarah",
      "is_locked": true,
      "created_at": "2026-01-29T11:00:00Z"
    }
  ]
}
```
Status: 200

---

### PUT /api/children/:id/password

Resets a child's password. Also unlocks the account if locked.

**User Story**: US5

**Auth**: Parent session required. Parent must own the child's family.

**Request**:
```json
{
  "password": "newsecret456"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| password | string | yes | Minimum 6 characters (FR-006) |

**Response**:
```json
{
  "message": "Password updated",
  "account_unlocked": true
}
```
Status: 200

**Error Responses**:
| Status | Body | Condition |
|--------|------|-----------|
| 400 | `{ "error": "Password too short" }` | FR-006 violation |
| 403 | `{ "error": "Forbidden" }` | Parent does not own this child's family |
| 404 | `{ "error": "Child not found" }` | No child with that ID |

---

### PUT /api/children/:id/name

Updates a child's display name.

**User Story**: US5

**Auth**: Parent session required. Parent must own the child's family.

**Request**:
```json
{
  "first_name": "Thomas"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| first_name | string | yes | Unique within family (FR-007) |

**Response**:
```json
{
  "message": "Name updated",
  "first_name": "Thomas"
}
```
Status: 200

**Error Responses**:
| Status | Body | Condition |
|--------|------|-----------|
| 409 | `{ "error": "Name taken" }` | FR-007 violation |
| 403 | `{ "error": "Forbidden" }` | Parent does not own this child's family |
| 404 | `{ "error": "Child not found" }` | No child with that ID |

---

## Health Check

### GET /api/health

**Auth**: None required

**Response**:
```json
{
  "status": "ok"
}
```
Status: 200

---

## Common Error Format

All error responses follow this structure:

```json
{
  "error": "Short error code",
  "message": "Human-readable message (optional, for display to users)"
}
```

## Authentication Flow

All authenticated endpoints expect the session cookie `session` to be present. The auth middleware:

1. Reads the `session` cookie
2. Looks up the token in the `sessions` table
3. Checks `expires_at > NOW()`
4. Attaches `user_type`, `user_id`, `family_id` to the request context
5. Returns 401 if any step fails
