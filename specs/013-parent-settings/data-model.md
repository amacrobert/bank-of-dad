# Data Model: Parent Settings Page

**Feature Branch**: `013-parent-settings`
**Date**: 2026-02-16

## Schema Changes

### Migration 003: Add timezone to families

**Table**: `families` (existing)

| Column    | Type         | Nullable | Default              | Notes                          |
|-----------|--------------|----------|----------------------|--------------------------------|
| id        | SERIAL       | NOT NULL | auto-increment       | Primary key (existing)         |
| slug      | TEXT         | NOT NULL | —                    | Unique family identifier (existing) |
| timezone  | TEXT         | NOT NULL | 'America/New_York'   | IANA timezone identifier (NEW) |
| created_at| TIMESTAMPTZ  | NOT NULL | NOW()                | (existing)                     |

**Migration SQL (up)**:
```sql
ALTER TABLE families ADD COLUMN timezone TEXT NOT NULL DEFAULT 'America/New_York';
```

**Migration SQL (down)**:
```sql
ALTER TABLE families DROP COLUMN timezone;
```

## Entity Changes

### Family (updated)

```
Family {
  ID        int64
  Slug      string
  Timezone  string    // IANA timezone identifier, e.g. "America/New_York"
  CreatedAt time.Time
}
```

**Validation rules**:
- `Timezone` must be a valid IANA timezone identifier (validated via `time.LoadLocation()`)
- `Timezone` must not be empty

## No New Tables

This feature adds a single column to an existing table. No new tables, indexes, or constraints are needed.
