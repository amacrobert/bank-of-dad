# Data Model: Free Tier Child Account Limits

## Schema Changes

### Migration: Add `is_disabled` to `children`

**Table**: `children`

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `is_disabled` | `BOOLEAN NOT NULL` | `FALSE` | Whether the child account is disabled due to free tier limits |

**Up migration**:
```sql
ALTER TABLE children ADD COLUMN is_disabled BOOLEAN NOT NULL DEFAULT FALSE;
```

**Down migration**:
```sql
ALTER TABLE children DROP COLUMN is_disabled;
```

No index needed on `is_disabled` alone — it will be filtered via JOINs in existing queries that already index on `child_id` or `family_id`.

## Entity Changes

### Child (modified)

| Field | Type | Change | Notes |
|-------|------|--------|-------|
| `is_disabled` | `bool` | **Added** | New field. Separate from `is_locked` (failed login lockout). |

All existing fields remain unchanged.

### Family (unchanged)

The `account_type` field (`"free"` or `"plus"`) already exists and determines the child limit. No schema changes to the `families` table.

## State Transitions

### Child `is_disabled` State Machine

```
                   ┌─────────────────────────────────────────┐
                   │                                         │
                   ▼                                         │
  ┌──────────┐  create (free, >2 enabled)  ┌──────────┐     │
  │  enabled  │ ◄──────────────────────────│ disabled  │     │
  │(is_disabled│   upgrade to Plus          │(is_disabled│    │
  │  = false) │   OR child deletion         │  = true)  │    │
  │           │   brings count ≤2           │           │    │
  └──────────┘                              └──────────┘    │
       │                                         ▲           │
       │   downgrade from Plus                   │           │
       │   (child is beyond earliest 2)          │           │
       └─────────────────────────────────────────┘           │
                                                              │
       create (free, ≤2 enabled OR Plus)                     │
       ──────────────────────► enabled ──────────────────────┘
```

**Triggers for disable**:
1. Child created on free tier when 2+ enabled children already exist
2. Family downgrades from Plus (children beyond earliest 2 by ID)

**Triggers for enable**:
1. Family upgrades to Plus (`checkout.session.completed`)
2. Enabled child deleted on free tier, bringing enabled count below 2

## Query Impact

### Queries requiring `is_disabled` filter

| Store | Method | Filter Logic |
|-------|--------|-------------|
| `AllowanceScheduleStore` | `ListDue` | Add `AND c.is_disabled = FALSE` to WHERE clause |
| `AllowanceScheduleStore` | `ListAllActiveWithTimezone` | Add `AND c.is_disabled = FALSE` to WHERE clause |
| `InterestScheduleStore` | `ListDue` | Add `AND c.is_disabled = FALSE` to WHERE clause |
| `InterestStore` | `ListDueForInterest` | Add `AND c.is_disabled = FALSE` to WHERE clause |

### Queries NOT requiring `is_disabled` filter

| Store | Method | Reason |
|-------|--------|--------|
| `ChildStore` | `ListByFamily` | Returns all children (needed for settings page) |
| `ChildStore` | `GetByID` | Used by handlers that do their own disabled check |
| `ChildStore` | `GetByFamilyAndName` | Used by login, which does its own disabled check |

## New Store Methods

### ChildStore additions

| Method | Signature | Description |
|--------|-----------|-------------|
| `EnableAllChildren` | `(familyID int64) error` | Sets `is_disabled = FALSE` for all children in family |
| `DisableExcessChildren` | `(familyID int64, limit int) error` | Disables children beyond the earliest `limit` (by ID) |
| `CountEnabledByFamily` | `(familyID int64) (int, error)` | Counts children where `is_disabled = FALSE` |
| `ReconcileChildLimits` | `(familyID int64, limit int) error` | Enables earliest `limit` children, disables the rest. Called after deletion and subscription changes. |
