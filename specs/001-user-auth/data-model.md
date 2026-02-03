# Data Model: User Authentication

**Feature**: 001-user-auth
**Date**: 2026-01-29
**Storage**: SQLite via `modernc.org/sqlite`

## Entity Relationship Overview

```
Parent (1) ──owns──> (1) Family (1) ──contains──> (0..*) Child
  │                       │                          │
  └──has──> (0..*) Session │                          └──has──> (0..*) Session
                          │
                          └──identified by──> slug (unique URL)
```

## Entities

### Parent

Represents a user who registers via Google OAuth and manages a family.

| Field | Type | Constraints | Source |
|-------|------|-------------|--------|
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Auto-generated |
| google_id | TEXT | NOT NULL, UNIQUE | Google OAuth userinfo |
| email | TEXT | NOT NULL | Google OAuth userinfo |
| display_name | TEXT | NOT NULL | Google OAuth userinfo |
| family_id | INTEGER | NOT NULL, FOREIGN KEY -> families(id) | Set during registration |
| created_at | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Auto-generated |

**Validation Rules**:
- `google_id` must be unique (FR-001, prevents duplicate registration)
- One parent per family (assumption: single parent per family initially)

**User Stories**: US1 (registration), US2 (login)

---

### Family

Represents an organizational unit with a unique URL slug.

| Field | Type | Constraints | Source |
|-------|------|-------------|--------|
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Auto-generated |
| slug | TEXT | NOT NULL, UNIQUE | Parent chooses during registration |
| created_at | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Auto-generated |

**Validation Rules** (FR-003):
- `slug` must match pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$` (or single char `^[a-z0-9]$`)
- `slug` length: 3-30 characters
- `slug` must be unique across all families (FR-004)

**User Stories**: US1 (creation), US4 (child login URL)

---

### Child

Represents a user created by a parent within a family.

| Field | Type | Constraints | Source |
|-------|------|-------------|--------|
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Auto-generated |
| family_id | INTEGER | NOT NULL, FOREIGN KEY -> families(id) | Inherited from parent's family |
| first_name | TEXT | NOT NULL | Parent provides |
| password_hash | TEXT | NOT NULL | bcrypt hash of parent-provided password |
| is_locked | BOOLEAN | NOT NULL, DEFAULT FALSE | Set after 5 failed login attempts |
| failed_login_attempts | INTEGER | NOT NULL, DEFAULT 0 | Incremented on failed login |
| created_at | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Auto-generated |
| updated_at | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Updated on name/password change |

**Validation Rules**:
- `first_name` must be unique within the same family (FR-007)
- Password minimum 6 characters before hashing (FR-006)
- `password_hash` uses bcrypt cost factor 12 (FR-008)
- Account locks after 5 consecutive failed attempts (FR-010)

**Unique Constraint**: `UNIQUE(family_id, first_name)`

**User Stories**: US3 (creation), US4 (login), US5 (credential management)

---

### Session

Represents an active authentication state for a user.

| Field | Type | Constraints | Source |
|-------|------|-------------|--------|
| token | TEXT | PRIMARY KEY | crypto/rand, 32 bytes, base64-encoded |
| user_type | TEXT | NOT NULL, CHECK IN ('parent', 'child') | Determined at login |
| user_id | INTEGER | NOT NULL | Parent.id or Child.id |
| family_id | INTEGER | NOT NULL, FOREIGN KEY -> families(id) | From user's family |
| created_at | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Auto-generated |
| expires_at | DATETIME | NOT NULL | created_at + TTL |

**TTL Rules**:
- Parent sessions: 7 days (FR-012)
- Child sessions: 24 hours (FR-013)

**Behavior**:
- Logout deletes the session row (FR-014)
- Expired sessions are rejected by middleware and lazily cleaned up
- Session token is sent as `HttpOnly`, `Secure`, `SameSite=Lax` cookie

**User Stories**: US2 (parent session), US4 (child session), US6 (persistence)

---

### AuthEvent (Audit Log)

Represents a logged authentication event for security auditing (FR-018).

| Field | Type | Constraints | Source |
|-------|------|-------------|--------|
| id | INTEGER | PRIMARY KEY, AUTOINCREMENT | Auto-generated |
| event_type | TEXT | NOT NULL | One of: login_success, login_failure, logout, account_created, account_locked, password_reset, name_updated |
| user_type | TEXT | NOT NULL | 'parent' or 'child' |
| user_id | INTEGER | NULL | NULL for failed logins of unknown users |
| family_id | INTEGER | NULL | NULL for failed logins of unknown families |
| ip_address | TEXT | NOT NULL | From request |
| details | TEXT | NULL | Non-sensitive context (e.g., "5th failed attempt") |
| created_at | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Auto-generated |

**Rules**:
- MUST NOT log passwords, tokens, or other sensitive data
- MUST log all events listed in FR-018

**User Stories**: Cross-cutting (all stories generate events)

---

## SQL Schema

```sql
CREATE TABLE families (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slug        TEXT    NOT NULL UNIQUE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE parents (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    google_id    TEXT    NOT NULL UNIQUE,
    email        TEXT    NOT NULL,
    display_name TEXT    NOT NULL,
    family_id    INTEGER NOT NULL REFERENCES families(id),
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE children (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    family_id             INTEGER NOT NULL REFERENCES families(id),
    first_name            TEXT    NOT NULL,
    password_hash         TEXT    NOT NULL,
    is_locked             INTEGER NOT NULL DEFAULT 0,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(family_id, first_name)
);

CREATE TABLE sessions (
    token      TEXT     PRIMARY KEY,
    user_type  TEXT     NOT NULL CHECK(user_type IN ('parent', 'child')),
    user_id    INTEGER  NOT NULL,
    family_id  INTEGER  NOT NULL REFERENCES families(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

CREATE TABLE auth_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type  TEXT     NOT NULL,
    user_type   TEXT     NOT NULL,
    user_id     INTEGER,
    family_id   INTEGER,
    ip_address  TEXT     NOT NULL,
    details     TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_auth_events_created ON auth_events(created_at);
CREATE INDEX idx_auth_events_family ON auth_events(family_id);
```

## State Transitions

### Child Account Locking

```
Active (failed_login_attempts < 5)
  │
  ├── login_failure → increment failed_login_attempts
  │   └── if failed_login_attempts >= 5 → Locked
  │
  └── login_success → reset failed_login_attempts to 0

Locked (is_locked = true)
  │
  └── parent resets password → set is_locked = false, failed_login_attempts = 0
```

### Session Lifecycle

```
Created (login success)
  │
  ├── request with valid token + not expired → Active (extend? no, fixed expiry)
  │
  ├── request with expired token → Expired (delete row, redirect to login)
  │
  └── explicit logout → Deleted (delete row)
```
