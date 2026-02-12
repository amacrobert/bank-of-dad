# Data Model: Child Avatars

**Feature**: 010-child-avatars
**Date**: 2026-02-11

## Entity Changes

### Children (modified)

| Field | Type | Nullable | Default | Notes |
|-------|------|----------|---------|-------|
| id | INTEGER | No | AUTOINCREMENT | PK, existing |
| family_id | INTEGER | No | — | FK → families, existing |
| first_name | TEXT | No | — | Existing |
| password_hash | TEXT | No | — | Existing |
| is_locked | INTEGER | No | 0 | Existing |
| failed_login_attempts | INTEGER | No | 0 | Existing |
| balance_cents | INTEGER | No | 0 | Existing |
| interest_rate_bps | INTEGER | No | 0 | Existing |
| last_interest_at | DATETIME | Yes | NULL | Existing |
| **avatar** | **TEXT** | **Yes** | **NULL** | **New — emoji string, e.g. "🌻"** |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Existing |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Existing |

### Migration

Add column using existing `addColumnIfNotExists` helper:

```
avatar TEXT
```

No constraints, no default, nullable. Existing rows get NULL (no avatar).

## Go Struct Change

```go
type Child struct {
    // ... existing fields ...
    Avatar *string  // nil = no avatar set
}
```

Uses `database/sql.NullString` for scanning from DB, mapped to `*string` on the struct.

## TypeScript Type Changes

```typescript
// types.ts
export interface Child {
  // ... existing fields ...
  avatar?: string;  // undefined = no avatar set
}
```

## Avatar Set (Frontend Constant)

```typescript
export const AVATAR_EMOJIS = [
  "🌻", "🌿", "🍂", "🌸",
  "🌊", "🌙", "⭐", "🦋",
  "🐝", "🍄", "🐸", "🦊",
  "🐻", "🐰", "🐢", "🎨",
] as const;
```

Defined once in the `AvatarPicker` component. Not stored in the database or served from the API.
