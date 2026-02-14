# Data Model: Stateless Authentication

**Feature**: 012-stateless-auth
**Date**: 2026-02-14

## Entity Changes

### New: refresh_tokens

Stores hashed refresh tokens for session management and revocation.

| Column      | Type                     | Constraints                                      |
|-------------|--------------------------|--------------------------------------------------|
| id          | SERIAL                   | PRIMARY KEY                                      |
| token_hash  | TEXT                     | NOT NULL, UNIQUE                                 |
| user_type   | TEXT                     | NOT NULL, CHECK(user_type IN ('parent', 'child'))|
| user_id     | INTEGER                  | NOT NULL                                         |
| family_id   | INTEGER                  | NOT NULL                                         |
| expires_at  | TIMESTAMPTZ              | NOT NULL                                         |
| created_at  | TIMESTAMPTZ              | NOT NULL, DEFAULT NOW()                          |

**Indexes**:
- `idx_refresh_tokens_token_hash` on `token_hash` (unique constraint provides this)
- `idx_refresh_tokens_expires` on `expires_at` (for cleanup queries)
- `idx_refresh_tokens_user` on `(user_type, user_id)` (for revoking all tokens on account lockout)

**Notes**:
- `token_hash` stores SHA-256 hash of the raw token (not the raw token itself). If the DB is compromised, tokens cannot be extracted.
- Raw token is sent to the client and stored in localStorage.
- Refresh token rotation: on each refresh, old row is deleted and new row is inserted.

### Removed: sessions

The `sessions` table is dropped. JWT access tokens replace server-side session validation.

| Column (removed) | Type (removed) |
|-------------------|----------------|
| token             | TEXT           |
| user_type         | TEXT           |
| user_id           | INTEGER        |
| family_id         | INTEGER        |
| created_at        | TIMESTAMPTZ    |
| expires_at        | TIMESTAMPTZ    |

### Unchanged: auth_events

No changes. Continues to log all authentication events.

### Unchanged: parents, children, families

No schema changes to user/family tables.

## JWT Access Token Claims

The access token is a signed JWT (HS256) with these claims:

| Claim       | Type    | Description                              |
|-------------|---------|------------------------------------------|
| sub         | string  | Subject identifier: `"parent:123"` or `"child:456"` |
| user_type   | string  | `"parent"` or `"child"`                  |
| user_id     | int     | Parent ID or Child ID                    |
| family_id   | int     | Family ID (0 if no family yet)           |
| iat         | int     | Issued at (Unix timestamp)               |
| exp         | int     | Expires at (Unix timestamp, iat + 15 min)|

**Not stored server-side.** Validated by verifying the HS256 signature and checking expiry.

## Migration

**Migration file**: `002_stateless_auth.up.sql`

```sql
-- Create refresh_tokens table
CREATE TABLE refresh_tokens (
    id         SERIAL PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    user_type  TEXT NOT NULL CHECK(user_type IN ('parent', 'child')),
    user_id    INTEGER NOT NULL,
    family_id  INTEGER NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_type, user_id);

-- Drop sessions table
DROP TABLE IF EXISTS sessions;
```

**Rollback file**: `002_stateless_auth.down.sql`

```sql
-- Recreate sessions table
CREATE TABLE sessions (
    token      TEXT PRIMARY KEY,
    user_type  TEXT NOT NULL CHECK(user_type IN ('parent', 'child')),
    user_id    INTEGER NOT NULL,
    family_id  INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- Drop refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens;
```

## State Transitions

### Refresh Token Lifecycle

```
Created (login)
    │
    ├── Used (refresh) → Deleted, new token created (rotation)
    │
    ├── Revoked (logout) → Deleted
    │
    └── Expired → Deleted by cleanup job
```

### Access Token Lifecycle

```
Issued (login or refresh)
    │
    ├── Valid → Accepted by middleware (stateless check)
    │
    └── Expired → Rejected by middleware (401)
```

No server-side state for access tokens. Expiry is the only invalidation mechanism.
