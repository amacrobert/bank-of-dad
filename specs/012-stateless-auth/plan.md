# Implementation Plan: Stateless Authentication

**Branch**: `012-stateless-auth` | **Date**: 2026-02-14 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/012-stateless-auth/spec.md`

## Summary

Replace cookie-based session authentication with stateless JWT access tokens and database-tracked refresh tokens. This enables the frontend and backend to be hosted on different domains without relying on cross-domain cookies. Access tokens (15-min JWT, HS256) are validated statelessly on every request. Refresh tokens (opaque, DB-tracked) support session extension and server-side revocation on logout.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `golang-jwt/jwt/v5` (new), `jackc/pgx/v5`, `testify`, react-router-dom, Vite
**Storage**: PostgreSQL 17 — new `refresh_tokens` table, drop `sessions` table
**Testing**: `go test -p 1 ./...` (backend), existing frontend patterns
**Target Platform**: Linux server (Docker/Fly.io), web browsers
**Project Type**: Web application (separate backend + frontend)
**Performance Goals**: Stateless JWT validation (no DB lookup per request); DB lookup only on token refresh (~1/15 min)
**Constraints**: Must work cross-domain (backend and frontend on different origins)
**Scale/Scope**: ~15 backend files changed, ~8 frontend files changed, 1 new migration

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development

- **Contract tests**: New `POST /api/auth/refresh` endpoint needs contract tests. Modified login endpoints need updated contract tests for token response shape.
- **Integration tests**: End-to-end auth flows (Google OAuth callback → token issuance, child login → token issuance, token refresh → rotation, logout → revocation).
- **Unit tests**: JWT signing/validation, refresh token store CRUD, Bearer token extraction middleware.
- **Status**: PASS — all code paths will have tests. Existing session_test.go replaced by refresh_token_test.go and jwt_test.go.

### II. Security-First Design

- **Token signing**: HS256 with 64-byte secret, validated on every request. Algorithm confusion prevented by checking signing method in parser.
- **Refresh token storage**: SHA-256 hash stored in DB (not raw token). DB compromise does not expose usable tokens.
- **Refresh token rotation**: Old token deleted on each refresh — stolen tokens have limited replay window.
- **Token revocation**: Refresh token deleted from DB on logout. Access token expires naturally in ≤15 minutes.
- **OAuth state**: CSRF protection via oauth_state cookie preserved (backend-only, no cross-domain issue).
- **CORS**: Authorization header already allowed. Credentials mode changes to false (no cookies needed).
- **Audit logging**: All auth events continue to be logged (login, logout, failure, lockout).
- **Status**: PASS — security posture maintained or improved.

### III. Simplicity

- **One new dependency**: `golang-jwt/jwt/v5` — well-justified (standard JWT library, 13k+ importers).
- **No new abstractions**: Refresh token store follows existing store pattern. JWT utilities are a single file.
- **Clean cutover**: No dual cookie+token period. Existing sessions invalidated on deploy.
- **localStorage**: Simplest cross-domain storage that survives page refresh.
- **Status**: PASS — minimal complexity added for the cross-domain requirement.

### Post-Phase 1 Re-check

- **Test-First**: All new entities (refresh_tokens, JWT claims) have corresponding test files. ✅
- **Security-First**: Token hashing, rotation, algorithm validation all present in design. ✅
- **Simplicity**: No unnecessary abstractions. Store pattern matches existing codebase. ✅

## Project Structure

### Documentation (this feature)

```text
specs/012-stateless-auth/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── auth-api.md
├── checklists/
│   └── requirements.md
└── tasks.md              # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── auth/
│   │   ├── jwt.go              # NEW — JWT signing, validation, claims struct
│   │   ├── jwt_test.go         # NEW — JWT unit tests
│   │   ├── google.go           # MODIFY — redirect with tokens instead of Set-Cookie
│   │   ├── child.go            # MODIFY — return tokens in response body
│   │   └── handlers.go         # MODIFY — refresh endpoint, update logout/me
│   ├── middleware/
│   │   └── auth.go             # MODIFY — Bearer token instead of cookie
│   ├── store/
│   │   ├── refresh_token.go      # NEW — refresh token CRUD (hash, create, validate, delete)
│   │   ├── refresh_token_test.go # NEW — refresh token store tests
│   │   ├── session.go            # DELETE — replaced by refresh_token.go + jwt.go
│   │   └── session_test.go       # DELETE — replaced by refresh_token_test.go
│   ├── family/
│   │   └── handlers.go         # MODIFY — return new access token on family creation
│   └── config/
│       └── config.go           # MODIFY — add JWT_SECRET, remove cookie settings
├── migrations/
│   ├── 002_stateless_auth.up.sql    # NEW
│   └── 002_stateless_auth.down.sql  # NEW
└── main.go                     # MODIFY — wire JWT, refresh store; remove session store/cleanup

frontend/
├── src/
│   ├── auth.ts                   # NEW — token storage, refresh, auth helpers
│   ├── api.ts                    # MODIFY — Bearer token instead of credentials:include
│   ├── pages/
│   │   ├── GoogleCallback.tsx    # MODIFY — extract tokens from URL params
│   │   ├── FamilyLogin.tsx       # MODIFY — store tokens from response
│   │   ├── SetupPage.tsx         # MODIFY — store updated token from family creation
│   │   ├── ParentDashboard.tsx   # MODIFY — use auth module
│   │   └── ChildDashboard.tsx    # MODIFY — use auth module
│   └── components/
│       └── Layout.tsx            # MODIFY — logout clears localStorage
```

**Structure Decision**: Existing web application structure (backend/ + frontend/) maintained. New files follow established patterns: stores in `internal/store/`, auth utilities in `internal/auth/`, frontend modules in `src/`.

## Complexity Tracking

No constitution violations. All design decisions align with the three core principles.
