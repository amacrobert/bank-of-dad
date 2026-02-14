# Quickstart: Stateless Authentication

**Feature**: 012-stateless-auth
**Date**: 2026-02-14

## Prerequisites

- Go 1.24+
- Node.js 18+
- PostgreSQL 17 (via Docker Compose)
- Existing `bankofdad` and `bankofdad_test` databases

## New Dependencies

### Backend
```bash
cd backend
go get github.com/golang-jwt/jwt/v5
```

### Frontend
No new dependencies.

## New Environment Variables

| Variable     | Required | Default | Description                                    |
|-------------|----------|---------|------------------------------------------------|
| JWT_SECRET  | Yes      | —       | Base64-encoded secret key (min 32 bytes decoded)|

### Removed Environment Variables

| Variable       | Reason                              |
|---------------|-------------------------------------|
| COOKIE_DOMAIN | Cookies no longer used for auth     |
| COOKIE_SECURE | Cookies no longer used for auth     |

### Generate JWT Secret (one time)

```bash
openssl rand -base64 64
```

Add to `.env` (local dev) or deployment secrets.

## Database Migration

```bash
cd backend
go run cmd/migrate/main.go up
```

This runs `002_stateless_auth.up.sql` which:
1. Creates `refresh_tokens` table
2. Drops `sessions` table

**Warning**: All existing sessions are invalidated. Users must log in again.

## Running Tests

```bash
cd backend
go test -p 1 ./...
```

The `-p 1` flag is required because all packages share the `bankofdad_test` database.

### Test Environment

Tests require `JWT_SECRET` to be set. Use any value for testing:
```bash
export JWT_SECRET=$(openssl rand -base64 64)
```

Or set `TEST_JWT_SECRET` in test helpers.

## Auth Flow Overview

### Parent Login (Google OAuth)
```
1. Browser → GET /api/auth/google/login
2. Browser → Google consent screen
3. Google → GET /api/auth/google/callback?code=...
4. Backend → 302 FRONTEND_URL/auth/callback?access_token=...&refresh_token=...
5. Frontend extracts tokens from URL, stores in localStorage
6. Frontend → GET /api/auth/me (Authorization: Bearer <access_token>)
```

### Child Login
```
1. Frontend → POST /api/auth/child/login {family_slug, first_name, password}
2. Backend → {access_token, refresh_token, user}
3. Frontend stores tokens in localStorage
```

### Token Refresh
```
1. Frontend detects access token expired (401 response or JWT exp claim)
2. Frontend → POST /api/auth/refresh {refresh_token}
3. Backend → {access_token, refresh_token} (rotated)
4. Frontend replaces stored tokens
```

### Logout
```
1. Frontend → POST /api/auth/logout {refresh_token} (Authorization: Bearer <access_token>)
2. Backend deletes refresh token from DB
3. Frontend clears localStorage
```

## File Changes Summary

### Backend — New Files
- `internal/auth/jwt.go` — JWT signing, validation, claims
- `internal/store/refresh_token.go` — Refresh token CRUD
- `migrations/002_stateless_auth.up.sql`
- `migrations/002_stateless_auth.down.sql`

### Backend — Modified Files
- `internal/middleware/auth.go` — Bearer token instead of cookie
- `internal/auth/google.go` — Redirect with tokens instead of Set-Cookie
- `internal/auth/child.go` — Return tokens in response body
- `internal/auth/handlers.go` — Refresh endpoint, update logout/me
- `internal/family/handlers.go` — Return new access token on family creation
- `internal/config/config.go` — Add JWT_SECRET, remove cookie settings
- `main.go` — Wire JWT, refresh token store; remove session store/cleanup

### Backend — Removed Files
- `internal/store/session.go` — Replaced by refresh_token.go + jwt.go

### Frontend — New Files
- `src/auth.ts` — Token storage, refresh logic, auth helpers

### Frontend — Modified Files
- `src/api.ts` — Bearer token instead of credentials:include
- `src/pages/GoogleCallback.tsx` — Extract tokens from URL params
- `src/pages/FamilyLogin.tsx` — Store tokens from response
- `src/pages/SetupPage.tsx` — Store updated access token from family creation
- `src/pages/ParentDashboard.tsx` — Use auth module for token state
- `src/pages/ChildDashboard.tsx` — Use auth module for token state
- `src/components/Layout.tsx` — Logout clears tokens from localStorage
