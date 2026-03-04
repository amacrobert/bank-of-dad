# Data Model: Savings Goals

**Branch**: `025-savings-goals` | **Date**: 2026-03-02

## New Entities

### savings_goals

Represents a child's savings target with a name, target amount, and progress tracking.

| Field          | Type                     | Constraints                                         | Description                                |
| -------------- | ------------------------ | --------------------------------------------------- | ------------------------------------------ |
| id             | SERIAL PRIMARY KEY       |                                                     | Unique identifier                          |
| child_id       | INTEGER NOT NULL         | REFERENCES children(id) ON DELETE CASCADE           | Owning child                               |
| name           | TEXT NOT NULL             | max 50 characters (app-enforced)                    | Goal name (e.g., "New Skateboard")         |
| target_cents   | BIGINT NOT NULL          | CHECK(target_cents > 0)                             | Target amount in cents                     |
| saved_cents    | BIGINT NOT NULL DEFAULT 0| CHECK(saved_cents >= 0)                             | Amount currently saved toward goal         |
| emoji          | TEXT                     | nullable                                            | Optional emoji icon for the goal           |
| target_date    | DATE                     | nullable                                            | Optional target completion date            |
| status         | TEXT NOT NULL DEFAULT 'active' | CHECK(status IN ('active', 'completed'))       | Goal lifecycle state                       |
| completed_at   | TIMESTAMPTZ              | nullable                                            | When the goal was achieved                 |
| created_at     | TIMESTAMPTZ NOT NULL     | DEFAULT NOW()                                       | Creation timestamp                         |
| updated_at     | TIMESTAMPTZ NOT NULL     | DEFAULT NOW()                                       | Last modification timestamp                |

**Indexes**:
- `idx_savings_goals_child_status` ON savings_goals(child_id, status) — fast lookup of active goals per child
- `idx_savings_goals_child_created` ON savings_goals(child_id, created_at DESC) — ordered listing

**State Transitions**:
- `active` → `completed`: When `saved_cents >= target_cents`, set `status = 'completed'` and `completed_at = NOW()`
- No reverse transition: completed goals stay completed

### goal_allocations

Records each transfer of funds to/from a savings goal. Provides audit trail per FR-013.

| Field          | Type                     | Constraints                                         | Description                                |
| -------------- | ------------------------ | --------------------------------------------------- | ------------------------------------------ |
| id             | SERIAL PRIMARY KEY       |                                                     | Unique identifier                          |
| goal_id        | INTEGER NOT NULL         | REFERENCES savings_goals(id) ON DELETE CASCADE      | Target savings goal                        |
| child_id       | INTEGER NOT NULL         | REFERENCES children(id) ON DELETE CASCADE           | Owning child (denormalized for queries)    |
| amount_cents   | BIGINT NOT NULL          | CHECK(amount_cents != 0)                            | Positive = allocation, negative = de-allocation |
| created_at     | TIMESTAMPTZ NOT NULL     | DEFAULT NOW()                                       | When the allocation occurred               |

**Indexes**:
- `idx_goal_allocations_goal` ON goal_allocations(goal_id, created_at DESC) — allocation history per goal

## Modified Entities

### children (existing)

No schema changes. The `balance_cents` column continues to represent the total balance (available + saved toward goals).

**Computed value** (not stored):
- `available_balance_cents` = `balance_cents` - SUM(`saved_cents` from active savings_goals for this child)
- Computed on read via a query or application logic

## Relationships

```
children (1) ──── (0..5) savings_goals
                           │
savings_goals (1) ──── (0..N) goal_allocations
```

- A child can have 0–5 **active** savings goals (app-enforced limit)
- A child can have unlimited **completed** savings goals
- Each savings goal has 0 or more allocation records
- Deleting a child cascades to their goals and allocations
- Deleting a goal cascades to its allocations

## Validation Rules

| Rule                                    | Enforced at | Description                                                  |
| --------------------------------------- | ----------- | ------------------------------------------------------------ |
| Goal name required, max 50 chars        | App         | Handler validates before insert                              |
| Target amount > 0                       | DB + App    | CHECK constraint + handler validation                        |
| Max 5 active goals per child            | App         | Handler counts active goals before creating                  |
| Allocation cannot exceed available balance | App      | Handler checks (balance - sum of active saved_cents) >= allocation |
| Saved cents cannot go negative          | DB + App    | CHECK constraint + handler validates de-allocation amount    |
| Status must be 'active' or 'completed'  | DB          | CHECK constraint                                            |

## Migration Plan

**Migration 007**: `007_savings_goals.up.sql`

Creates both tables with constraints and indexes. Down migration drops both tables.

No changes to existing tables are required. The available balance is computed, not stored.
