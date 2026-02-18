# Data Model: Child Visual Themes

**Feature**: 017-child-visual-themes
**Date**: 2026-02-18

## Schema Changes

### Migration 004: Add theme column to children

```sql
-- 004_child_theme.up.sql
ALTER TABLE children ADD COLUMN theme TEXT;

-- 004_child_theme.down.sql
ALTER TABLE children DROP COLUMN theme;
```

**Notes**:
- Nullable TEXT, no default. NULL = "sapling" (handled in application layer).
- Matches the `avatar` column pattern exactly.
- No index needed — theme is only read alongside the full child row, never queried independently.

## Entity Changes

### Children (modified)

| Column              | Type         | Nullable | Default  | Change   |
|---------------------|--------------|----------|----------|----------|
| id                  | SERIAL       | NO       | auto     | existing |
| family_id           | INTEGER (FK) | NO       |          | existing |
| first_name          | TEXT         | NO       |          | existing |
| password_hash       | TEXT         | NO       |          | existing |
| is_locked           | BOOLEAN      | NO       | FALSE    | existing |
| failed_login_attempts | INTEGER    | NO       | 0        | existing |
| balance_cents       | BIGINT       | NO       | 0        | existing |
| interest_rate_bps   | INTEGER      | NO       | 0        | existing |
| last_interest_at    | TIMESTAMPTZ  | YES      |          | existing |
| avatar              | TEXT         | YES      |          | existing |
| **theme**           | **TEXT**     | **YES**  |          | **NEW**  |
| created_at          | TIMESTAMPTZ  | NO       | NOW()    | existing |
| updated_at          | TIMESTAMPTZ  | NO       | NOW()    | existing |

### Valid theme values

| Slug       | Display Name | Description            |
|------------|-------------|------------------------|
| `NULL`     | Sapling     | Default theme (green)  |
| `sapling`  | Sapling     | Explicit default       |
| `piggybank`| Piggy Bank  | Rose/pink theme        |
| `rainbow`  | Rainbow     | Purple/lavender theme  |

Application layer treats `NULL` and `"sapling"` identically.

## Go Struct Change

```go
type Child struct {
    ID                  int64
    FamilyID            int64
    FirstName           string
    PasswordHash        string
    IsLocked            bool
    FailedLoginAttempts int
    BalanceCents        int64
    Avatar              *string  // existing
    Theme               *string  // NEW — nullable, nil = "sapling"
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```

## Frontend Type Change

```typescript
export interface ChildUser {
    user_type: "child";
    user_id: number;
    family_id: number;
    first_name: string;
    family_slug: string;
    avatar?: string | null;
    theme?: string | null;  // NEW — null/undefined = "sapling"
}
```

## Relationships

```
families 1──∞ children (existing)
                └── theme: TEXT (new column, per-child preference)
```

No new tables. No new foreign keys. No new indexes.
