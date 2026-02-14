# API Contracts: Stateless Authentication

**Feature**: 012-stateless-auth
**Date**: 2026-02-14

## Authentication Pattern

All authenticated endpoints require:
```
Authorization: Bearer <access_token>
```

Access tokens are JWTs (HS256, 15-minute expiry). Refresh tokens are opaque strings stored in localStorage.

---

## Modified Endpoints

### POST /api/auth/child/login

Child credential login. Returns token pair instead of setting cookies.

**Request**:
```json
{
  "family_slug": "string",
  "first_name": "string",
  "password": "string"
}
```

**Response 200**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "base64url-random-string",
  "user": {
    "user_type": "child",
    "user_id": 456,
    "family_id": 42,
    "first_name": "Emma",
    "family_slug": "macrobert"
  }
}
```

**Response 401** (invalid credentials):
```json
{
  "error": "Invalid credentials"
}
```

**Response 403** (account locked):
```json
{
  "error": "Account is locked. Ask your parent to reset your password."
}
```

**Changes from current**: Response now includes `access_token` and `refresh_token` fields. No `Set-Cookie` header. User info nested under `user` key.

---

### GET /api/auth/google/login

Unchanged. Redirects to Google OAuth consent screen. Sets `oauth_state` cookie (backend-only, not cross-domain).

---

### GET /api/auth/google/callback

Google OAuth callback. Redirects to frontend with tokens in query params instead of setting cookies.

**Behavior**:
1. Validates `oauth_state` cookie
2. Exchanges code with Google
3. Creates or finds parent user
4. Issues access + refresh tokens
5. Redirects to: `FRONTEND_URL/auth/callback?access_token=<jwt>&refresh_token=<token>`
   - New user or no family: `FRONTEND_URL/auth/callback?access_token=<jwt>&refresh_token=<token>&redirect=/setup`
   - Existing user with family: `FRONTEND_URL/auth/callback?access_token=<jwt>&refresh_token=<token>&redirect=/dashboard`

**Changes from current**: No `Set-Cookie: session=...`. Tokens delivered via URL query parameters. Frontend extracts tokens and clears URL.

---

### GET /api/auth/me

Returns current user info. Now reads Bearer token instead of session cookie.

**Request**: `Authorization: Bearer <access_token>`

**Response 200** (parent):
```json
{
  "user_type": "parent",
  "user_id": 123,
  "family_id": 42,
  "display_name": "Andrew",
  "email": "andrew@example.com",
  "family_slug": "macrobert"
}
```

**Response 200** (child):
```json
{
  "user_type": "child",
  "user_id": 456,
  "family_id": 42,
  "first_name": "Emma",
  "family_slug": "macrobert",
  "avatar": "bear"
}
```

**Response 401**: Invalid or expired token.

**Changes from current**: Reads `Authorization: Bearer` header instead of `session` cookie. No other changes to response shape.

---

### POST /api/auth/logout

Revokes the refresh token. Now accepts refresh token in request body instead of reading session cookie.

**Request**: `Authorization: Bearer <access_token>`
```json
{
  "refresh_token": "base64url-random-string"
}
```

**Response 200**:
```json
{
  "message": "Logged out successfully"
}
```

**Changes from current**: Accepts refresh_token in body. Deletes refresh token from DB. No `Set-Cookie` with MaxAge=-1. Access token continues to work until its 15-minute expiry, but no new access tokens can be obtained.

---

### POST /api/families

Create a family. Returns a new access token with updated `family_id`.

**Request**: `Authorization: Bearer <access_token>`
```json
{
  "name": "Macrobert",
  "slug": "macrobert"
}
```

**Response 201**:
```json
{
  "id": 42,
  "name": "Macrobert",
  "slug": "macrobert",
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Changes from current**: Response now includes `access_token` with updated `family_id` claim. No longer calls `sessionStore.UpdateFamilyID()`.

---

## New Endpoints

### POST /api/auth/refresh

Exchange a valid refresh token for a new access token + refresh token pair (rotation).

**Request**:
```json
{
  "refresh_token": "base64url-random-string"
}
```

**Response 200**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "new-base64url-random-string"
}
```

**Response 401** (expired, revoked, or invalid refresh token):
```json
{
  "error": "Invalid or expired refresh token"
}
```

**Notes**:
- No `Authorization` header required (the access token may be expired).
- Old refresh token is deleted from DB, new one is inserted (rotation).
- New access token contains current user data (user_type, user_id, family_id looked up from DB).
- Rate limited: 10 requests/minute per IP (same as child login).

---

## CORS Changes

**Headers** (no changes to header names, but behavior note):
```
Access-Control-Allow-Origin: <FRONTEND_URL>
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: false
```

**Change**: `Access-Control-Allow-Credentials` changes from `true` to `false`. Cookies are no longer used for authentication, so credential sharing is unnecessary. The `Authorization` header is sent explicitly by the frontend, not automatically by the browser.

---

## Removed Cookie Operations

| Operation | File | Line | Status |
|-----------|------|------|--------|
| Set `session` cookie | google.go:159-167 | Removed |
| Set `session` cookie | child.go:138-146 | Removed |
| Read `session` cookie | middleware/auth.go:24 | Replaced with Bearer token |
| Clear `session` cookie | middleware/auth.go:33-39 | Removed |
| Clear `session` cookie | handlers.go:93-101 | Removed |
| Read `session` cookie | handlers.go:88 | Replaced with refresh_token from body |
| Read `session` cookie | family/handlers.go:84-86 | Replaced with new access token in response |
| Set `oauth_state` cookie | google.go:65-73 | **Kept** (backend-only) |
| Read `oauth_state` cookie | google.go:81 | **Kept** (backend-only) |
| Clear `oauth_state` cookie | google.go:94-100 | **Kept** (backend-only) |
