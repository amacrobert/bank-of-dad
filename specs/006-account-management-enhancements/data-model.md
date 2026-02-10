# Data Model: Account Management Enhancements

## Modified Entities

### AllowanceSchedule (existing table: `allowance_schedules`)

**Change**: Add UNIQUE constraint on `child_id` to enforce one allowance per child.

**Migration**:
1. Check for duplicate `child_id` values in `allowance_schedules`
2. If duplicates exist, keep the newest schedule (highest `id`) per child, delete others
3. Add UNIQUE constraint on `child_id`

```sql
-- After deduplication:
CREATE UNIQUE INDEX IF NOT EXISTS idx_allowance_schedules_child_id
ON allowance_schedules(child_id);
```

**Existing fields** (unchanged):
| Field | Type | Description |
|-------|------|-------------|
| id | INTEGER PK | Auto-increment |
| child_id | INTEGER FK | References children.id (now UNIQUE) |
| parent_id | INTEGER FK | References parents.id |
| amount_cents | INTEGER | Allowance amount in cents |
| frequency | TEXT | 'weekly', 'biweekly', 'monthly' |
| day_of_week | INTEGER NULL | 0=Sunday..6=Saturday (for weekly/biweekly) |
| day_of_month | INTEGER NULL | 1-31 (for monthly) |
| note | TEXT NULL | Optional note |
| status | TEXT | 'active' or 'paused' |
| next_run_at | DATETIME NULL | Next scheduled execution |
| created_at | DATETIME | Creation timestamp |
| updated_at | DATETIME | Last update timestamp |

### ScheduleStore (existing)

**New method**:
- `GetByChildID(childID int64) (*AllowanceSchedule, error)` — Returns the single schedule for a child (any status), or nil if none exists.

## New Entities

### InterestSchedule (new table: `interest_schedules`)

Represents when interest is calculated and credited for a child. One per child.

| Field | Type | Description |
|-------|------|-------------|
| id | INTEGER PK | Auto-increment |
| child_id | INTEGER FK UNIQUE | References children.id |
| parent_id | INTEGER FK | References parents.id |
| frequency | TEXT | 'weekly', 'biweekly', 'monthly' |
| day_of_week | INTEGER NULL | 0=Sunday..6=Saturday (for weekly/biweekly) |
| day_of_month | INTEGER NULL | 1-31 (for monthly) |
| status | TEXT | 'active' or 'paused' |
| next_run_at | DATETIME NULL | Next scheduled accrual |
| created_at | DATETIME | Creation timestamp |
| updated_at | DATETIME | Last update timestamp |

```sql
CREATE TABLE IF NOT EXISTS interest_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id INTEGER NOT NULL UNIQUE REFERENCES children(id) ON DELETE CASCADE,
    parent_id INTEGER NOT NULL REFERENCES parents(id),
    frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly')),
    day_of_week INTEGER,
    day_of_month INTEGER,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'paused')),
    next_run_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key differences from allowance_schedules**:
- No `amount_cents` — interest amount is calculated from balance and rate
- No `note` — interest transactions auto-generate notes with rate info
- `child_id` is UNIQUE from creation (one schedule per child)

### InterestScheduleStore (new)

| Method | Description |
|--------|-------------|
| `Create(sched *InterestSchedule)` | Create interest schedule for a child |
| `GetByChildID(childID int64)` | Get the schedule for a child (any status) |
| `Update(sched *InterestSchedule)` | Update frequency/day fields, recalculate next_run_at |
| `Delete(id int64)` | Remove a schedule |
| `ListDue(now time.Time)` | List all active schedules where next_run_at <= now |
| `UpdateNextRunAt(id int64, nextRunAt time.Time)` | Set next execution time after accrual |
| `UpdateStatus(id int64, status ScheduleStatus)` | Pause/resume |

## Modified Responses

### BalanceResponse (existing)

**New field**:
| Field | Type | Description |
|-------|------|-------------|
| next_interest_at | string (ISO 8601) NULL | Next scheduled interest payment date, from interest_schedules.next_run_at |

## Entity Relationships

```
children (1) ──── (0..1) allowance_schedules   [UNIQUE constraint on child_id]
children (1) ──── (0..1) interest_schedules     [UNIQUE constraint on child_id]
children (1) ──── (many) transactions           [unchanged]
```

## Migration Notes

### Deduplication Strategy for allowance_schedules

Since the app is a personal/family tool, duplicate allowances per child are unlikely but must be handled:

```sql
-- Delete all but the newest schedule per child (keep highest id)
DELETE FROM allowance_schedules
WHERE id NOT IN (
    SELECT MAX(id) FROM allowance_schedules GROUP BY child_id
);
```

### Interest Scheduler Changes

The existing interest scheduler (`interest.Scheduler`) currently:
1. Runs on a fixed interval (1 hour)
2. Calls `ListDueForInterest()` which checks `last_interest_at` against current month

**New behavior**:
1. Still runs on a fixed interval (1 hour) as a polling mechanism
2. Calls `InterestScheduleStore.ListDue(time.Now())` to find schedules where `next_run_at <= now`
3. For each due schedule, calls `ApplyInterest` with the appropriate `periodsPerYear` based on frequency
4. After successful accrual, calls `UpdateNextRunAt` to set the next occurrence

**Removed**: The `last_interest_at` column on `children` is no longer needed for duplicate prevention — the schedule's `next_run_at` serves that purpose. However, it can be retained for audit purposes.

### ApplyInterest Changes

Current signature: `ApplyInterest(childID, parentID int64, rateBps int) error`

New signature: `ApplyInterest(childID, parentID int64, rateBps int, periodsPerYear int) error`

The `periodsPerYear` parameter replaces the hardcoded `/12` divisor:
- Weekly: 52
- Biweekly: 26
- Monthly: 12
