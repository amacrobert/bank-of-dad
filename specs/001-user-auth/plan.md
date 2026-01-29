# Implementation Plan: User Authentication

**Branch**: `001-user-auth` | **Date**: 2026-01-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-user-auth/spec.md`

## Summary

Implement user authentication for Bank of Dad: Google OAuth for parent registration/login, password-based child login via family-specific URLs, session management with role-based TTLs, and child account CRUD by parents. Built with Go 1.21 backend (SQLite storage, server-side sessions) and React + TypeScript frontend (Vite, react-router-dom).

## Technical Context

**Language/Version**: Go 1.21 (backend), TypeScript 5.3 + React 18.2 (frontend)
**Primary Dependencies**: `golang.org/x/oauth2` (Google OAuth), `golang.org/x/crypto/bcrypt` (password hashing), `modernc.org/sqlite` (database), `react-router-dom` (frontend routing), `testify` (Go test assertions)
**Storage**: SQLite via `modernc.org/sqlite` (pure Go, CGO-free, WAL mode)
**Testing**: Go standard `testing` + `github.com/stretchr/testify` (assert/require only); frontend: Vite built-in
**Target Platform**: Web browsers (Docker-deployed Linux server)
**Project Type**: Web application (separate backend + frontend)
**Performance Goals**: 100 concurrent authenticated users without degradation (SC-004)
**Constraints**: Registration < 2 min (SC-001), child login < 30 seconds (SC-003), account lockout within 1 second of 5th failure (SC-007)
**Scale/Scope**: Small family app; ~6 pages, 4 entities, ~12 API endpoints

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development

- **Contract tests for all API endpoints**: PASS - All API endpoints will have contract tests in `backend/tests/contract/`
- **Integration tests for user journeys**: PASS - Each user story gets integration tests in `backend/tests/integration/`
- **Unit tests for business logic**: PASS - Service and store layers tested via `_test.go` files
- **TDD workflow**: PASS - Tasks will be ordered tests-first per constitution mandate

### II. Security-First Design

- **Data protection**: PASS - Child passwords hashed with bcrypt (cost 12); sessions use `HttpOnly`, `Secure`, `SameSite=Lax` cookies; SQLite file permissions restricted
- **Authentication**: PASS - All endpoints except health check and public family URL lookup require authentication via session middleware
- **Authorization**: PASS - Session tokens store `user_type`, `user_id`, `family_id`; middleware enforces parent/child role separation
- **Input validation**: PASS - URL slug validation (FR-003), password requirements (FR-006), unique name enforcement (FR-007)
- **Logging**: PASS - All auth events logged without sensitive data (FR-018)

### III. Simplicity

- **YAGNI**: PASS - No JWT, no Redis, no external session store, no multi-provider OAuth
- **Minimal dependencies**: PASS - 3 Go modules (2 quasi-stdlib), 1 npm package, native `fetch` for HTTP
- **Clear over clever**: PASS - Server-side sessions in SQLite, standard OAuth2 flow, bcrypt 2-function API

**GATE RESULT**: PASS - No violations. All three principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/001-user-auth/
├── plan.md              # This file
├── research.md          # Phase 0: technology decisions
├── data-model.md        # Phase 1: entity definitions
├── quickstart.md        # Phase 1: test scenarios
├── contracts/           # Phase 1: API endpoint definitions
└── tasks.md             # Phase 2: implementation tasks (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── main.go                          # HTTP server entry point, router setup
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── Dockerfile                       # Production multi-stage build
├── internal/
│   ├── auth/
│   │   ├── google.go                # Google OAuth handlers (login, callback)
│   │   ├── child.go                 # Child login/logout handlers
│   │   ├── session.go               # Session middleware, cookie management
│   │   └── handlers.go              # Shared auth utilities
│   ├── family/
│   │   ├── handlers.go              # Family CRUD handlers (slug, children)
│   │   └── validation.go            # Slug and name validation
│   ├── store/
│   │   ├── sqlite.go                # SQLite connection, migrations, pragmas
│   │   ├── parent.go                # Parent CRUD operations
│   │   ├── family.go                # Family CRUD operations
│   │   ├── child.go                 # Child CRUD operations
│   │   └── session.go               # Session CRUD operations
│   ├── middleware/
│   │   ├── auth.go                  # Authentication middleware
│   │   ├── cors.go                  # CORS middleware
│   │   └── logging.go               # Request/auth event logging
│   └── config/
│       └── config.go                # Environment configuration
├── tests/
│   ├── contract/                    # API contract tests
│   │   ├── auth_test.go
│   │   ├── family_test.go
│   │   └── child_test.go
│   └── integration/                 # User journey tests
│       ├── parent_registration_test.go
│       ├── parent_login_test.go
│       ├── child_account_test.go
│       └── child_login_test.go

frontend/
├── index.html                       # HTML entry point
├── package.json                     # Dependencies
├── vite.config.ts                   # Vite config (proxy to backend)
├── tsconfig.json                    # TypeScript config
├── Dockerfile                       # Production build
├── Dockerfile.dev                   # Development with HMR
├── nginx.conf                       # SPA routing
├── src/
│   ├── main.tsx                     # React DOM mount
│   ├── App.tsx                      # Router setup, route definitions
│   ├── api.ts                       # Typed fetch wrapper
│   ├── types.ts                     # Shared TypeScript types
│   ├── pages/
│   │   ├── HomePage.tsx             # Landing / marketing page
│   │   ├── ParentDashboard.tsx      # Parent family view
│   │   ├── ChildDashboard.tsx       # Child account view
│   │   ├── FamilyLogin.tsx          # Child login via /:familySlug
│   │   ├── GoogleCallback.tsx       # OAuth callback handler
│   │   └── NotFound.tsx             # 404 page
│   └── components/
│       ├── AddChildForm.tsx         # Create child account form
│       ├── ChildList.tsx            # List children in family
│       ├── ManageChild.tsx          # Reset password, update name
│       ├── SlugPicker.tsx           # Choose family bank URL slug
│       ├── ProtectedRoute.tsx       # Auth guard component
│       └── Layout.tsx               # Shared layout (nav, logout)
```

**Structure Decision**: Web application layout matching the existing `backend/` and `frontend/` directory structure already in the repository. Go code uses `internal/` package convention for encapsulation. Frontend uses `pages/` for route-level components and `components/` for shared UI.

## Complexity Tracking

> No constitution violations identified. No entries needed.
