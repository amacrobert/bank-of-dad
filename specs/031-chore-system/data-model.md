# Data Model: Chore & Task System

**Feature Branch**: `031-chore-system`
**Created**: 2026-03-22

## New Entities

### Chore

A task template defined by a parent, belonging to a family.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | serial PK | auto-increment | Unique identifier |
| family_id | FK → families | NOT NULL | Owning family |
| created_by_parent_id | FK → parents | NOT NULL | Parent who created the chore |
| name | varchar(100) | NOT NULL | Chore name (e.g., "Mow the lawn") |
| description | varchar(500) | nullable | Optional details |
| reward_cents | integer | NOT NULL, >= 0 | Reward amount in cents ($0.00 allowed) |
| recurrence | varchar(10) | NOT NULL, default 'one_time' | One of: one_time, daily, weekly, monthly |
| day_of_week | smallint | nullable | 0=Sun..6=Sat (for weekly recurrence) |
| day_of_month | smallint | nullable | 1-31 (for monthly recurrence) |
| is_active | boolean | NOT NULL, default true | Whether new instances are generated |
| created_at | timestamptz | NOT NULL, auto | Creation timestamp |
| updated_at | timestamptz | NOT NULL, auto | Last update timestamp |

**Indexes**:
- `idx_chores_family_id` on `family_id`

**Validations**:
- `name` must be 1-100 characters, trimmed
- `reward_cents` must be 0 to 99,999,999 (consistent with MaxAmountCents)
- `recurrence` must be one of the allowed values
- `day_of_week` required when recurrence is "weekly"
- `day_of_month` required when recurrence is "monthly"

---

### Chore Assignment

Join table linking a chore to one or more children.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | serial PK | auto-increment | Unique identifier |
| chore_id | FK → chores | NOT NULL, ON DELETE CASCADE | Parent chore |
| child_id | FK → children | NOT NULL | Assigned child |
| created_at | timestamptz | NOT NULL, auto | When assignment was made |

**Indexes**:
- `idx_chore_assignments_chore_id` on `chore_id`
- `idx_chore_assignments_child_id` on `child_id`
- `uq_chore_assignments_chore_child` UNIQUE on `(chore_id, child_id)`

**Validations**:
- Child must belong to the same family as the chore
- No duplicate assignments (enforced by unique constraint)

---

### Chore Instance

A specific occurrence of a chore for a specific child. Created once for one-time chores, or per-period for recurring chores.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | serial PK | auto-increment | Unique identifier |
| chore_id | FK → chores | NOT NULL | Parent chore template |
| child_id | FK → children | NOT NULL | Assigned child |
| reward_cents | integer | NOT NULL | Snapshotted reward at instance creation |
| status | varchar(20) | NOT NULL, default 'available' | One of: available, pending_approval, approved, rejected, expired |
| period_start | date | nullable | Start of the recurrence period (null for one-time) |
| period_end | date | nullable | End of the recurrence period (null for one-time) |
| completed_at | timestamptz | nullable | When child marked as complete |
| reviewed_at | timestamptz | nullable | When parent approved/rejected |
| reviewed_by_parent_id | FK → parents | nullable | Parent who reviewed |
| rejection_reason | varchar(500) | nullable | Optional reason for rejection |
| transaction_id | FK → transactions | nullable | Resulting deposit transaction (on approval) |
| created_at | timestamptz | NOT NULL, auto | Instance creation timestamp |
| updated_at | timestamptz | NOT NULL, auto | Last update timestamp |

**Indexes**:
- `idx_chore_instances_chore_id` on `chore_id`
- `idx_chore_instances_child_id` on `child_id`
- `idx_chore_instances_status` on `status`
- `uq_chore_instances_chore_child_period` UNIQUE on `(chore_id, child_id, period_start)` WHERE `period_start IS NOT NULL`

**Validations**:
- `status` must be one of the allowed values
- `completed_at` set when status transitions to pending_approval
- `reviewed_at` and `reviewed_by_parent_id` set when status transitions to approved or rejected
- `transaction_id` set only when status is approved and reward_cents > 0

---

## State Transitions

### Chore Instance Status

```
                    ┌──────────────┐
                    │   available  │ ◄─── created by scheduler or on chore creation
                    └──────┬───────┘
                           │ child marks complete
                           ▼
                    ┌──────────────────┐
              ┌───► │ pending_approval  │
              │     └────────┬─────────┘
              │              │
              │    ┌─────────┴─────────┐
              │    │                   │
              │    ▼                   ▼
              │ ┌──────────┐    ┌───────────┐
              │ │ rejected │    │  approved  │ ──► deposit transaction created
              │ └────┬─────┘    └───────────┘
              │      │
              └──────┘  (returns to available)

                    ┌──────────────┐
                    │   available  │
                    └──────┬───────┘
                           │ period elapses (recurring only)
                           ▼
                    ┌──────────────┐
                    │   expired    │
                    └──────────────┘
```

**Rejection handling**: When a chore instance is rejected, its status resets to "available" and `completed_at` is cleared. The child can mark it complete again. The `rejection_reason` is preserved on the instance until the child re-completes it (at which point it's cleared).

---

## Modified Entities

### Transaction (existing)

**Change**: Add new transaction type constant.

| Change | Detail |
|--------|--------|
| New constant | `TransactionTypeChore TransactionType = "chore"` |
| New method | `DepositChore(childID, parentID, amountCents, choreInstanceID int64, note string)` on TransactionRepo |

No schema migration needed — `transaction_type` is a VARCHAR column. The `schedule_id` field is not used for chore transactions; instead, the chore instance stores the `transaction_id` reference.

---

## Relationships

```
families 1──────────* chores
parents  1──────────* chores (created_by)
chores   1──────────* chore_assignments
children 1──────────* chore_assignments
chores   1──────────* chore_instances
children 1──────────* chore_instances
parents  1──────────* chore_instances (reviewed_by)
transactions 1─────? chore_instances (on approval)
```
