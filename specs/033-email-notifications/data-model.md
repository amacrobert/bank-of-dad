# Data Model: Email Notifications

**Feature Branch**: `033-email-notifications`  
**Date**: 2026-04-02

## Schema Changes

### Modified Table: `parents`

Three new boolean columns added to the existing `parents` table.

| Column | Type | Default | Nullable | Description |
|--------|------|---------|----------|-------------|
| notify_withdrawal_requests | BOOLEAN | TRUE | NOT NULL | Whether parent receives emails for child withdrawal requests |
| notify_chore_completions | BOOLEAN | TRUE | NOT NULL | Whether parent receives emails for child chore completions |
| notify_decisions | BOOLEAN | TRUE | NOT NULL | Whether parent receives emails when another parent approves/denies |

### Migration: `014_notification_preferences`

**Up**:
```sql
ALTER TABLE parents
    ADD COLUMN notify_withdrawal_requests BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN notify_chore_completions BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN notify_decisions BOOLEAN NOT NULL DEFAULT TRUE;
```

**Down**:
```sql
ALTER TABLE parents
    DROP COLUMN notify_withdrawal_requests,
    DROP COLUMN notify_chore_completions,
    DROP COLUMN notify_decisions;
```

## Updated Entity: Parent

```
Parent
├── id (PK)
├── google_id (unique, not null)
├── email (not null)
├── display_name (not null)
├── family_id (FK → families.id)
├── notify_withdrawal_requests (bool, default true)   ← NEW
├── notify_chore_completions (bool, default true)      ← NEW
├── notify_decisions (bool, default true)              ← NEW
└── created_at
```

## Relationships

- **Parent → Family**: Many-to-one (unchanged). Used to find all parents in a family for notification dispatch.
- **Parent ↔ Notification Preferences**: Embedded in parent record (no separate entity needed for 3 fixed preference types).

## State Transitions

No new state machines. Notifications are triggered by existing state transitions:

| Existing Transition | Triggers Notification To |
|---|---|
| Withdrawal request: → `pending` | All opted-in parents in family |
| Chore instance: → `pending_approval` | All opted-in parents in family (batched) |
| Withdrawal request: → `approved` | Other opted-in parents in family |
| Withdrawal request: → `denied` | Other opted-in parents in family |
| Chore instance: → `approved` | Other opted-in parents in family |
| Chore instance: → `rejected` | Other opted-in parents in family |

## Validation Rules

- `notify_withdrawal_requests`, `notify_chore_completions`, `notify_decisions` must be boolean values.
- All three columns default to `true` for new and existing parents (migration applies default).
- Unsubscribe action sets all three to `false` atomically.
