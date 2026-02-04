# Data Model: Allowance Scheduling

**Feature**: 003-allowance-scheduling
**Date**: 2026-02-04

## Entity Overview

```
┌─────────────┐       ┌─────────────┐       ┌──────────────────┐
│   Family    │──1:N──│   Parent    │       │                  │
└─────────────┘       └─────────────┘       │                  │
      │                     │               │                  │
      │                     │ creates       │                  │
      │                     ▼               │                  │
      │              ┌──────────────┐       │                  │
      └──────1:N─────│    Child     │───────│  has schedules   │
                     │              │       │                  │
                     │ balance_cents│       │                  │
                     └──────────────┘       │                  │
                           │                │                  │
                           │ 1:N            │                  │
                           ▼                │                  │
                ┌─────────────────────┐     │                  │
                │ AllowanceSchedule   │◀────┘                  │
                │                     │  created by parent     │
                │ frequency           │                        │
                │ day_of_week/month   │                        │
                │ amount_cents        │                        │
                │ next_run_at         │                        │
                └─────────────────────┘                        │
                           │                                   │
                           │ generates (0:N)                   │
                           ▼                                   │
                     ┌──────────────┐                          │
                     │ Transaction  │◀─────────────────────────┘
                     │              │  manual deposits also
                     │ schedule_id? │  (optional reference)
                     └──────────────┘
```

## Schema Changes

### New: `allowance_schedules` table

```sql
CREATE TABLE IF NOT EXISTS allowance_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id INTEGER NOT NULL,
    parent_id INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly')),
    day_of_week INTEGER CHECK(day_of_week >= 0 AND day_of_week <= 6),
    day_of_month INTEGER CHECK(day_of_month >= 1 AND day_of_month <= 31),
    note TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'paused')),
    next_run_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT,

    -- Ensure correct day field is set based on frequency
    CHECK(
        (frequency = 'weekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'biweekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'monthly' AND day_of_month IS NOT NULL)
    )
);

-- Index for efficient "what's due?" queries
CREATE INDEX idx_schedules_due ON allowance_schedules(status, next_run_at)
    WHERE status = 'active';

-- Index for listing schedules by family (parent's view)
CREATE INDEX idx_schedules_child ON allowance_schedules(child_id);
```

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Schedule ID |
| child_id | INTEGER | NOT NULL, FK(children) | Child receiving allowance |
| parent_id | INTEGER | NOT NULL, FK(parents) | Parent who created schedule |
| amount_cents | INTEGER | NOT NULL | Allowance amount in cents |
| frequency | TEXT | NOT NULL, CHECK | 'weekly', 'biweekly', or 'monthly' |
| day_of_week | INTEGER | CHECK 0-6 | For weekly/biweekly: 0=Sunday, 6=Saturday |
| day_of_month | INTEGER | CHECK 1-31 | For monthly: day of month (clamped to last day) |
| note | TEXT | NULLABLE | Optional description (appears on transactions) |
| status | TEXT | NOT NULL, DEFAULT 'active' | 'active' or 'paused' |
| next_run_at | DATETIME | NULLABLE | When this schedule should next execute |
| created_at | DATETIME | NOT NULL | Creation timestamp |
| updated_at | DATETIME | NOT NULL | Last modification timestamp |

### Modify: `transactions` table

Add optional schedule reference:

```sql
-- Migration: Add schedule_id column
ALTER TABLE transactions ADD COLUMN schedule_id INTEGER REFERENCES allowance_schedules(id) ON DELETE SET NULL;
```

Update transaction_type CHECK constraint to include 'allowance':

```sql
-- Note: SQLite doesn't support modifying CHECK constraints, so we need to:
-- 1. Accept 'allowance' as valid (implement in application layer)
-- 2. Or recreate table (only if strict validation needed)

-- Application-layer validation will accept: 'deposit', 'withdrawal', 'allowance'
```

**Updated transactions table structure:**
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Transaction ID |
| child_id | INTEGER | NOT NULL, FK(children) | Child account affected |
| parent_id | INTEGER | NOT NULL, FK(parents) | Parent who made/owns the schedule |
| amount_cents | INTEGER | NOT NULL | Amount in cents (always positive) |
| transaction_type | TEXT | NOT NULL | 'deposit', 'withdrawal', or 'allowance' |
| note | TEXT | NULLABLE | Optional description |
| **schedule_id** | **INTEGER** | **NULLABLE, FK(schedules)** | **Reference to schedule (NEW)** |
| created_at | DATETIME | NOT NULL | Transaction timestamp |

## Go Model Definitions

### AllowanceSchedule Model

```go
// backend/internal/store/schedule.go

type Frequency string

const (
    FrequencyWeekly   Frequency = "weekly"
    FrequencyBiweekly Frequency = "biweekly"
    FrequencyMonthly  Frequency = "monthly"
)

type ScheduleStatus string

const (
    ScheduleStatusActive ScheduleStatus = "active"
    ScheduleStatusPaused ScheduleStatus = "paused"
)

type AllowanceSchedule struct {
    ID          int64          `json:"id"`
    ChildID     int64          `json:"child_id"`
    ParentID    int64          `json:"parent_id"`
    AmountCents int64          `json:"amount_cents"`
    Frequency   Frequency      `json:"frequency"`
    DayOfWeek   *int           `json:"day_of_week,omitempty"`   // 0-6, nil for monthly
    DayOfMonth  *int           `json:"day_of_month,omitempty"`  // 1-31, nil for weekly/biweekly
    Note        *string        `json:"note,omitempty"`
    Status      ScheduleStatus `json:"status"`
    NextRunAt   *time.Time     `json:"next_run_at,omitempty"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}
```

### Extended Transaction Model

```go
// backend/internal/store/transaction.go (modify existing)

const (
    TransactionTypeDeposit    TransactionType = "deposit"
    TransactionTypeWithdrawal TransactionType = "withdrawal"
    TransactionTypeAllowance  TransactionType = "allowance"  // NEW
)

type Transaction struct {
    ID              int64           `json:"id"`
    ChildID         int64           `json:"child_id"`
    ParentID        int64           `json:"parent_id"`
    AmountCents     int64           `json:"amount_cents"`
    TransactionType TransactionType `json:"type"`
    Note            *string         `json:"note,omitempty"`
    ScheduleID      *int64          `json:"schedule_id,omitempty"`  // NEW
    CreatedAt       time.Time       `json:"created_at"`
}
```

### Response Models

```go
// For parent's schedule list view
type ScheduleListResponse struct {
    Schedules []ScheduleWithChild `json:"schedules"`
}

type ScheduleWithChild struct {
    AllowanceSchedule
    ChildFirstName string `json:"child_first_name"`
}

// For child's upcoming allowance view
type UpcomingAllowance struct {
    AmountCents int64     `json:"amount_cents"`
    NextDate    time.Time `json:"next_date"`
    Note        *string   `json:"note,omitempty"`
}

type UpcomingAllowancesResponse struct {
    Allowances []UpcomingAllowance `json:"allowances"`
}
```

## TypeScript Model Definitions

### Frontend Types

```typescript
// frontend/src/types.ts

export type Frequency = 'weekly' | 'biweekly' | 'monthly';
export type ScheduleStatus = 'active' | 'paused';

export interface AllowanceSchedule {
  id: number;
  child_id: number;
  parent_id: number;
  amount_cents: number;
  frequency: Frequency;
  day_of_week?: number;   // 0-6 for weekly/biweekly
  day_of_month?: number;  // 1-31 for monthly
  note?: string;
  status: ScheduleStatus;
  next_run_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ScheduleWithChild extends AllowanceSchedule {
  child_first_name: string;
}

export interface ScheduleListResponse {
  schedules: ScheduleWithChild[];
}

export interface CreateScheduleRequest {
  child_id: number;
  amount_cents: number;
  frequency: Frequency;
  day_of_week?: number;
  day_of_month?: number;
  note?: string;
}

export interface UpdateScheduleRequest {
  amount_cents?: number;
  frequency?: Frequency;
  day_of_week?: number;
  day_of_month?: number;
  note?: string;
}

export interface UpcomingAllowance {
  amount_cents: number;
  next_date: string;
  note?: string;
}

export interface UpcomingAllowancesResponse {
  allowances: UpcomingAllowance[];
}
```

## Validation Rules

### Amount Validation
- Must be a positive integer > 0
- Maximum: 99999999 cents ($999,999.99) - same as manual deposits
- No decimal values accepted (API uses cents)

### Frequency Validation
- Must be one of: 'weekly', 'biweekly', 'monthly'
- Weekly/Biweekly requires `day_of_week` (0-6)
- Monthly requires `day_of_month` (1-31)

### Day Validation
- `day_of_week`: 0 = Sunday, 1 = Monday, ..., 6 = Saturday
- `day_of_month`: 1-31 (clamped to last day for short months at execution time)

### Note Validation
- Optional field
- Maximum length: 500 characters
- Trimmed of leading/trailing whitespace
- Empty string treated as NULL

## State Transitions

### Schedule Status State Machine

```
              ┌─────────────┐
              │   active    │ (initial state)
              └──────┬──────┘
                     │
          pause()    │    resume()
                     │
              ┌──────▼──────┐
              │   paused    │
              └─────────────┘
                     │
          delete()   │   delete()
                     │
              ┌──────▼──────┐
              │  (deleted)  │
              └─────────────┘
```

**Invariants:**
- Active schedules have `next_run_at` set to a future date
- Paused schedules may have `next_run_at = NULL` or unchanged
- Deleted schedules are removed from database (CASCADE handles transactions FK)
- A schedule can have multiple transactions (0:N relationship)

## Relationships

| From | To | Cardinality | Constraint |
|------|-----|-------------|------------|
| Child | AllowanceSchedule | 1:N | Child can have multiple schedules |
| Parent | AllowanceSchedule | 1:N | Parent creates/owns schedules |
| AllowanceSchedule | Transaction | 1:N | Schedule generates transactions (optional FK) |
| Family | Child | 1:N | Family has many children (existing) |
| Parent | Family | N:1 | Parent belongs to family (existing) |

## Indexes

| Table | Index | Columns | Purpose |
|-------|-------|---------|---------|
| allowance_schedules | idx_schedules_due | (status, next_run_at) WHERE status='active' | Find due schedules |
| allowance_schedules | idx_schedules_child | (child_id) | List schedules for a child |

## Migration Strategy

1. Create `allowance_schedules` table with all columns
2. Create indexes on allowance_schedules
3. Add `schedule_id` column to transactions table (nullable FK)
4. No data migration needed (new feature, starting fresh)

```sql
-- Full migration script
CREATE TABLE IF NOT EXISTS allowance_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id INTEGER NOT NULL,
    parent_id INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly')),
    day_of_week INTEGER CHECK(day_of_week >= 0 AND day_of_week <= 6),
    day_of_month INTEGER CHECK(day_of_month >= 1 AND day_of_month <= 31),
    note TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'paused')),
    next_run_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT,
    CHECK(
        (frequency = 'weekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'biweekly' AND day_of_week IS NOT NULL) OR
        (frequency = 'monthly' AND day_of_month IS NOT NULL)
    )
);

CREATE INDEX idx_schedules_due ON allowance_schedules(status, next_run_at);
CREATE INDEX idx_schedules_child ON allowance_schedules(child_id);

-- Add schedule reference to transactions (idempotent)
-- Note: Check if column exists before adding
```
