# Research: Allowance Scheduling

**Feature**: 003-allowance-scheduling
**Date**: 2026-02-04

## Research Summary

No NEEDS CLARIFICATION items in Technical Context. All technical decisions align with existing codebase patterns discovered during exploration. The existing session cleanup goroutine pattern provides a proven foundation for scheduled task execution.

---

## Decision 1: Background Scheduler Architecture

**Decision**: Use a single background goroutine with a 1-minute ticker to process due schedules

**Rationale**:
- Follows existing pattern in session.go `StartCleanupLoop()`
- No external dependencies required (uses Go stdlib `time` package)
- Sufficient granularity for "within 1 hour" requirement (SC-002)
- Simple to implement, test, and reason about
- Graceful shutdown via channel pattern already established

**Alternatives Considered**:
- External job scheduler (cron, etc.): Adds deployment complexity, violates Simplicity
- Database-triggered approach: SQLite doesn't support triggers that call Go code
- Per-schedule goroutine: Memory overhead for potentially many schedules, harder to manage
- Longer polling interval (hourly): Would not meet SC-002 requirement

**Implementation Notes**:
```go
func (s *Scheduler) StartScheduleProcessor(interval time.Duration, stop <-chan struct{}) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        // Process immediately on start (catch any missed while down)
        s.ProcessDueSchedules()
        for {
            select {
            case <-ticker.C:
                s.ProcessDueSchedules()
            case <-stop:
                return
            }
        }
    }()
}
```

---

## Decision 2: Schedule Execution Tracking

**Decision**: Store `next_run_at` timestamp on each schedule; update after successful execution

**Rationale**:
- Simple query: `SELECT * FROM schedules WHERE status = 'active' AND next_run_at <= ?`
- Efficient with index on `(status, next_run_at)`
- Handles missed runs: after downtime, all schedules with `next_run_at < now` are processed
- No complex date math at query time; calculation happens once when scheduling

**Alternatives Considered**:
- Store only frequency + day, calculate at runtime: Complex queries, error-prone
- Last-run timestamp with offset: Requires date math in query, harder to index
- Separate "pending executions" table: Over-engineered for this scale

**Implementation Notes**:
- On schedule creation: calculate and set `next_run_at`
- On successful execution: calculate next occurrence and update `next_run_at`
- On pause: set `next_run_at = NULL` or far future
- On resume: recalculate `next_run_at` from today

---

## Decision 3: Frequency Representation

**Decision**: Store frequency as TEXT enum ('weekly', 'biweekly', 'monthly') with separate day fields

**Rationale**:
- Clear, readable values in database
- Easy validation with CHECK constraint
- Different day semantics per frequency handled by separate fields:
  - Weekly: `day_of_week` (0-6, Sunday=0)
  - Biweekly: `day_of_week` + anchor date
  - Monthly: `day_of_month` (1-31)

**Alternatives Considered**:
- Single integer "interval_days": Doesn't handle months well (28-31 days vary)
- Cron expression string: Overly complex for 3 simple frequencies
- Separate boolean columns: Inflexible, harder to extend

**Implementation Notes**:
- Weekly: day_of_week = 0-6 (Sunday-Saturday)
- Biweekly: day_of_week + uses next_run_at to maintain 14-day cycle
- Monthly: day_of_month = 1-31 (clamped to last day for short months)

---

## Decision 4: Handling Monthly Edge Cases

**Decision**: For months with fewer days than specified, use the last day of that month

**Rationale**:
- Spec explicitly requires this (FR-009, acceptance scenario for 31st)
- Common pattern in financial scheduling
- User expectation: "31st" means "end of month" in short months

**Alternatives Considered**:
- Skip the deposit: Violates user expectation and spec
- Carry forward to next month: Creates confusion, doesn't match spec
- Require only valid-for-all-months days (1-28): Too restrictive

**Implementation Notes**:
```go
func nextMonthlyDate(dayOfMonth int, after time.Time) time.Time {
    year, month, _ := after.Date()
    // Start with next month
    month++
    if month > 12 {
        month = 1
        year++
    }
    // Clamp to last day of month if needed
    lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
    if dayOfMonth > lastDay {
        dayOfMonth = lastDay
    }
    return time.Date(year, month, dayOfMonth, 0, 0, 0, 0, time.UTC)
}
```

---

## Decision 5: Transaction Linkage

**Decision**: Add optional `schedule_id` foreign key to transactions table; add 'allowance' transaction type

**Rationale**:
- Clear audit trail: which schedule created this transaction
- Enables queries like "show all deposits from this schedule"
- New transaction type distinguishes automatic from manual deposits
- Preserves existing transaction structure; nullable FK for backward compatibility

**Alternatives Considered**:
- Note field only: Loses structured relationship, harder to query
- Separate scheduled_transactions table: Over-engineered, breaks existing transaction list
- Store schedule info in transaction note: Unstructured, fragile

**Implementation Notes**:
- Migration: `ALTER TABLE transactions ADD COLUMN schedule_id INTEGER REFERENCES allowance_schedules(id) ON DELETE SET NULL`
- Modify TransactionType to include `TransactionTypeAllowance = "allowance"`
- When schedule executes deposit, include schedule_id in transaction

---

## Decision 6: Child Visibility of Upcoming Allowance

**Decision**: Add API endpoint for child to get their upcoming allowances; calculate from active schedules

**Rationale**:
- Satisfies US4 and FR-011
- Simple query: active schedules for child, sorted by next_run_at
- Returns list (child may have multiple schedules)
- Frontend displays nearest one prominently

**Alternatives Considered**:
- Calculate only on frontend: Duplicates date logic, inconsistent with server
- Store "next allowance summary" separately: Redundant data, sync issues

**Implementation Notes**:
- Endpoint: `GET /api/children/{id}/upcoming-allowances`
- Returns: `[{ amount_cents, next_date, note }, ...]` sorted by next_date
- Child can only see their own; parent can see any child in family

---

## Decision 7: Pause/Resume Behavior

**Decision**: Paused schedules do not accumulate; on resume, calculate next occurrence from current date

**Rationale**:
- Matches spec assumption: "Paused schedules do not accumulate missed deposits"
- Simple mental model for parents
- Avoids surprise bulk deposits on resume

**Alternatives Considered**:
- Accumulate missed deposits: Complex, potentially large unexpected deposits
- Resume from where paused: Confusing if paused for months

**Implementation Notes**:
- On pause: set `status = 'paused'`, optionally clear `next_run_at`
- On resume: set `status = 'active'`, recalculate `next_run_at` from today
- Scheduler query ignores paused schedules

---

## Decision 8: Missed Schedule Recovery

**Decision**: On startup and each tick, process all schedules where `next_run_at <= now()`

**Rationale**:
- Satisfies FR-010: "Process any missed scheduled deposits when recovering from downtime"
- Simple implementation: same query handles normal and catch-up processing
- Each missed schedule processed exactly once (next_run_at updated after execution)

**Alternatives Considered**:
- Separate "recovery mode" on startup: Unnecessary complexity
- Limit catch-up to N deposits: Could lose allowances, violates spec

**Implementation Notes**:
- Query: `WHERE status = 'active' AND next_run_at <= ?`
- Process each, update next_run_at, repeat until none due
- Log when processing schedules that were significantly overdue

---

## Existing Code Patterns to Follow

### Store Pattern (from schedule.go - to be created)
```go
type ScheduleStore struct {
    db *DB
}

func NewScheduleStore(db *DB) *ScheduleStore {
    return &ScheduleStore{db: db}
}

func (s *ScheduleStore) Create(schedule *AllowanceSchedule) (*AllowanceSchedule, error) {
    // Calculate next_run_at
    // INSERT INTO allowance_schedules
    // Return created schedule with ID
}

func (s *ScheduleStore) ListDue(now time.Time) ([]AllowanceSchedule, error) {
    // SELECT * FROM allowance_schedules WHERE status = 'active' AND next_run_at <= ?
}
```

### Scheduler Pattern (following session.go)
```go
type Scheduler struct {
    scheduleStore   *ScheduleStore
    transactionStore *TransactionStore
}

func (s *Scheduler) ProcessDueSchedules() {
    schedules, _ := s.scheduleStore.ListDue(time.Now())
    for _, sched := range schedules {
        s.executeSchedule(sched)
    }
}

func (s *Scheduler) executeSchedule(sched AllowanceSchedule) error {
    // Create transaction via TransactionStore.Deposit()
    // Update schedule's next_run_at
}
```

### Handler Pattern (following balance/handler.go)
```go
type Handler struct {
    scheduleStore *store.ScheduleStore
    childStore    *store.ChildStore
}

func (h *Handler) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Validate parent owns child
    // Create schedule
    // Return schedule JSON
}
```

---

## Dependencies

No new dependencies required. Feature uses:
- `database/sql` (existing)
- `modernc.org/sqlite` (existing)
- `testify` (existing)
- `time` (Go stdlib, already used)
- React (existing)
- react-router-dom (existing)
