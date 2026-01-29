# Research: User Authentication

**Feature**: 001-user-auth
**Date**: 2026-01-29

## 1. Database Storage

**Decision**: SQLite via `modernc.org/sqlite` (pure Go, CGO-free)

**Rationale**: Embedded database requires no separate server process or Docker container. Pure Go driver works with the existing `CGO_ENABLED=0` Dockerfile. WAL mode supports concurrent reads with a single writer, sufficient for 100 concurrent users. A single `.db` file simplifies deployment and backup.

**Alternatives Considered**:
- **PostgreSQL**: Requires separate container, connection credentials, network config. Unnecessary operational complexity at this scale.
- **`mattn/go-sqlite3`**: Requires CGO, would break existing Dockerfile and complicate cross-compilation.
- **File-based storage (JSON)**: No query capability, no transactions, no concurrent access safety.

**Configuration**:
- Write pool: `SetMaxOpenConns(1)` to avoid `SQLITE_BUSY`
- Read pool: `SetMaxOpenConns(runtime.NumCPU())`
- PRAGMAs: `journal_mode=WAL`, `synchronous=NORMAL`, `busy_timeout=5000`, `foreign_keys=ON`

---

## 2. Google OAuth 2.0

**Decision**: `golang.org/x/oauth2` with `golang.org/x/oauth2/google`

**Rationale**: Quasi-standard library maintained by the Go team. Minimal footprint, 46k+ importers, handles CSRF state parameter and token exchange correctly. After token exchange, use Google's userinfo endpoint with standard `net/http` to retrieve Google ID, email, and display name -- no additional Google API client library needed.

**Alternatives Considered**:
- **Raw HTTP implementation**: Error-prone for CSRF protection, token exchange, and refresh. Violates "clear over clever."
- **Third-party libraries (goth, authboss)**: Unnecessary abstraction for a single-provider setup. Violates YAGNI and minimal dependencies.

**Flow**:
1. Frontend redirects to `/api/auth/google/login`
2. Backend generates state token, stores in session/cookie, redirects to Google
3. Google redirects back to `/api/auth/google/callback` with auth code
4. Backend exchanges code for token, fetches userinfo, creates/finds parent, creates session
5. Backend redirects to frontend with session cookie set

---

## 3. Password Hashing

**Decision**: `golang.org/x/crypto/bcrypt` with cost factor 12

**Rationale**: Two-function API (`GenerateFromPassword`, `CompareHashAndPassword`) is the simplest correct approach. Same `golang.org/x/crypto` package as argon2 so dependency footprint is identical. Cost factor 12 provides ~200-300ms hashing time, appropriate for interactive authentication. Sufficient security for child passwords (min 6 chars) in a family educational app.

**Alternatives Considered**:
- **Argon2id**: Superior algorithm (memory-hard, GPU-resistant) but requires manual salt generation, 4 tuning parameters, and custom encoding. Added complexity provides no meaningful additional protection for this threat model.
- **scrypt**: Less intuitive API, fewer guarantees than Argon2id. If upgrading beyond bcrypt, Argon2id would be the better choice.
- **SHA-256 + salt**: Cryptographically inappropriate for password hashing (too fast, vulnerable to brute force).

---

## 4. Session Management

**Decision**: Server-side token table in SQLite with secure HTTP-only cookies

**Rationale**: Generate 32-byte cryptographically random token (`crypto/rand`), base64-encode, store in `sessions` table with `expires_at`. Send to client as `HttpOnly`, `Secure`, `SameSite=Lax` cookie. Zero external dependencies. Different TTLs per row (7 days for parents, 24 hours for children). Logout deletes the row -- instant revocation with no token blacklisting.

**Alternatives Considered**:
- **JWT (stateless tokens)**: Cannot be revoked without server-side blacklist, defeating their purpose. Spec requires explicit logout (FR-014). Adds signing key management and JWT library dependency.
- **Third-party session libraries (gorilla/sessions, scs)**: Unnecessary dependency for ~50 lines of implementation code.
- **In-memory session store (Go map)**: Sessions lost on restart. Unacceptable for 7-day parent sessions.

**Session cookie settings**:
- `HttpOnly: true` (no JavaScript access)
- `Secure: true` (HTTPS only; relaxed in development)
- `SameSite: Lax` (CSRF protection while allowing OAuth redirects)
- `Path: /`

---

## 5. Testing Framework

**Decision**: Go standard `testing` + `github.com/stretchr/testify` (assert/require packages only)

**Rationale**: Standard `testing` is the non-negotiable foundation. Testify's `assert` and `require` packages reduce assertion boilerplate (1 line vs 3 lines per check) and produce clearer failure messages. Critical for TDD workflow with hundreds of assertions. Do NOT use testify's `mock` or `suite` packages -- use idiomatic Go interfaces and `TestMain` instead.

**Alternatives Considered**:
- **Standard library only**: Viable but increases test verbosity significantly for a TDD-mandated project.
- **Ginkgo/Gomega**: Heavy BDD framework with its own DSL. Violates simplicity.

**Test organization**:
- `backend/internal/*_test.go`: Unit tests (same package)
- `backend/tests/contract/`: Contract tests against API endpoints
- `backend/tests/integration/`: End-to-end user journey tests

---

## 6. Frontend Routing

**Decision**: `react-router-dom` v6+ (declarative mode)

**Rationale**: Bank of Dad needs ~6 routes including dynamic family slugs (`/:familySlug`). react-router-dom's API is minimal and familiar. No build tooling changes required. `useParams()` handles the family URL pattern naturally.

**Alternatives Considered**:
- **TanStack Router**: Superior type safety and data loading, but overkill for ~6 routes. Steeper setup and learning curve.
- **Manual `window.location`**: Would require reimplementing browser history management.

---

## 7. Frontend HTTP Client

**Decision**: Native `fetch` API with a thin typed wrapper

**Rationale**: Zero dependencies, built into every modern browser. A ~30-line wrapper handles JSON parsing, error checking, and typing. Constitution says "prefer standard library solutions" -- `fetch` is the web platform's standard.

**Alternatives Considered**:
- **Axios**: Adds ~35 KB dependency for features not needed (interceptors, upload progress). Violates minimal dependencies.
- **ky**: Lightweight (~3 KB) but still an unnecessary dependency for the current scope.

---

## Summary

| Topic | Decision | New Dependency | Type |
|-------|----------|---------------|------|
| Database | SQLite (`modernc.org/sqlite`) | 1 Go module | Pure Go, CGO-free |
| Google OAuth | `golang.org/x/oauth2` | 1 Go module | Quasi-stdlib |
| Password Hashing | `golang.org/x/crypto/bcrypt` | 1 Go module | Quasi-stdlib |
| Sessions | Server-side tokens + cookies | 0 | stdlib only |
| Testing | `testify/assert` + `testify/require` | 1 Go module | Assertions only |
| Frontend Routing | `react-router-dom` | 1 npm package | Declarative mode |
| Frontend HTTP | Native `fetch` wrapper | 0 | Browser built-in |

**Total new Go dependencies**: 4 (2 quasi-stdlib, 1 pure Go SQLite, 1 test-only)
**Total new npm dependencies**: 1
