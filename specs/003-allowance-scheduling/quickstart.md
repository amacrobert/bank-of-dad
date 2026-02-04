# Quickstart: Allowance Scheduling

**Feature**: 003-allowance-scheduling
**Date**: 2026-02-04

## Overview

This feature adds automatic recurring allowance deposits. Parents create schedules that automatically deposit money to their children's accounts on a weekly, biweekly, or monthly basis. A background goroutine processes due schedules.

## Key Concepts

### Frequency Types
- **weekly**: Deposits every week on a specific day (0=Sunday, 6=Saturday)
- **biweekly**: Deposits every two weeks on a specific day
- **monthly**: Deposits on a specific day of month (1-31, clamped for short months)

### Schedule Status
- **active**: Schedule will execute at `next_run_at`
- **paused**: Schedule is suspended; no deposits occur

### next_run_at Field
- Stores the next scheduled execution time
- Updated after each successful deposit
- Query: `WHERE status = 'active' AND next_run_at <= now()`

## Development Workflow

### TDD Cycle
1. Write failing test
2. Run test, confirm red
3. Implement minimal code
4. Run test, confirm green
5. Refactor if needed

### Running Tests

```bash
# Backend tests
cd backend && go test ./... -v

# Specific package
cd backend && go test ./internal/store -v -run TestSchedule
cd backend && go test ./internal/allowance -v

# Frontend type check
cd frontend && npm run build
```

### Running the App

```bash
# Start all services
docker compose up --build

# Backend only (for testing)
cd backend && go run main.go

# Frontend only
cd frontend && npm run dev
```

## Implementation Patterns

### Store Pattern (follow existing)

```go
// backend/internal/store/schedule.go
type ScheduleStore struct {
    db *DB
}

func NewScheduleStore(db *DB) *ScheduleStore {
    return &ScheduleStore{db: db}
}

func (s *ScheduleStore) Create(schedule *AllowanceSchedule) (*AllowanceSchedule, error) {
    // Calculate next_run_at before insert
    schedule.NextRunAt = calculateNextRun(schedule)

    result, err := s.db.Write.Exec(`
        INSERT INTO allowance_schedules
        (child_id, parent_id, amount_cents, frequency, day_of_week, day_of_month, note, status, next_run_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, 'active', ?)
    `, schedule.ChildID, schedule.ParentID, schedule.AmountCents,
       schedule.Frequency, schedule.DayOfWeek, schedule.DayOfMonth,
       schedule.Note, schedule.NextRunAt)
    // ...
}

func (s *ScheduleStore) ListDue(now time.Time) ([]AllowanceSchedule, error) {
    rows, err := s.db.Read.Query(`
        SELECT id, child_id, parent_id, amount_cents, frequency,
               day_of_week, day_of_month, note, status, next_run_at,
               created_at, updated_at
        FROM allowance_schedules
        WHERE status = 'active' AND next_run_at <= ?
    `, now)
    // ...
}
```

### Handler Pattern (follow existing)

```go
// backend/internal/allowance/handler.go
type Handler struct {
    scheduleStore *store.ScheduleStore
    childStore    *store.ChildStore
}

func NewHandler(scheduleStore *store.ScheduleStore, childStore *store.ChildStore) *Handler {
    return &Handler{
        scheduleStore: scheduleStore,
        childStore:    childStore,
    }
}

func (h *Handler) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
    session := middleware.GetSession(r.Context())
    if session == nil || session.Role != "parent" {
        http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
        return
    }
    // Parse, validate, create schedule
}
```

### Background Scheduler Pattern (follow session cleanup)

```go
// backend/internal/allowance/scheduler.go
type Scheduler struct {
    scheduleStore    *store.ScheduleStore
    transactionStore *store.TransactionStore
    childStore       *store.ChildStore
}

func (s *Scheduler) Start(interval time.Duration, stop <-chan struct{}) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        // Process immediately on start (catch missed)
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

func (s *Scheduler) ProcessDueSchedules() {
    schedules, err := s.scheduleStore.ListDue(time.Now())
    if err != nil {
        log.Printf("Error listing due schedules: %v", err)
        return
    }
    for _, sched := range schedules {
        if err := s.executeSchedule(sched); err != nil {
            log.Printf("Error executing schedule %d: %v", sched.ID, err)
        }
    }
}

func (s *Scheduler) executeSchedule(sched store.AllowanceSchedule) error {
    // 1. Create deposit transaction
    // 2. Update child balance
    // 3. Calculate and update next_run_at
    return nil
}
```

### Date Calculation Helpers

```go
// Calculate next run time based on frequency
func calculateNextRun(sched *AllowanceSchedule) time.Time {
    now := time.Now().UTC()

    switch sched.Frequency {
    case FrequencyWeekly:
        return nextWeeklyDate(*sched.DayOfWeek, now)
    case FrequencyBiweekly:
        return nextBiweeklyDate(*sched.DayOfWeek, now)
    case FrequencyMonthly:
        return nextMonthlyDate(*sched.DayOfMonth, now)
    }
    return now
}

func nextWeeklyDate(dayOfWeek int, after time.Time) time.Time {
    // Find next occurrence of dayOfWeek after 'after'
    daysUntil := (dayOfWeek - int(after.Weekday()) + 7) % 7
    if daysUntil == 0 {
        daysUntil = 7 // Next week if today is the day
    }
    return time.Date(after.Year(), after.Month(), after.Day()+daysUntil,
                     0, 0, 0, 0, time.UTC)
}

func nextMonthlyDate(dayOfMonth int, after time.Time) time.Time {
    year, month, _ := after.Date()

    // Try this month first
    target := time.Date(year, month, min(dayOfMonth, daysInMonth(year, month)),
                        0, 0, 0, 0, time.UTC)
    if target.After(after) {
        return target
    }

    // Otherwise next month
    month++
    if month > 12 {
        month = 1
        year++
    }
    return time.Date(year, month, min(dayOfMonth, daysInMonth(year, month)),
                     0, 0, 0, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
    return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
```

### Frontend API Functions

```typescript
// frontend/src/api.ts
export async function getSchedules(): Promise<ScheduleListResponse> {
  const res = await fetch('/api/schedules', { credentials: 'include' });
  if (!res.ok) throw new Error('Failed to fetch schedules');
  return res.json();
}

export async function createSchedule(req: CreateScheduleRequest): Promise<AllowanceSchedule> {
  const res = await fetch('/api/schedules', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || 'Failed to create schedule');
  }
  return res.json();
}

export async function pauseSchedule(id: number): Promise<AllowanceSchedule> {
  const res = await fetch(`/api/schedules/${id}/pause`, {
    method: 'POST',
    credentials: 'include',
  });
  if (!res.ok) throw new Error('Failed to pause schedule');
  return res.json();
}

export async function getUpcomingAllowances(childId: number): Promise<UpcomingAllowancesResponse> {
  const res = await fetch(`/api/children/${childId}/upcoming-allowances`, {
    credentials: 'include',
  });
  if (!res.ok) throw new Error('Failed to fetch upcoming allowances');
  return res.json();
}
```

### Test Patterns

```go
// Store test pattern
func TestScheduleStore_Create(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    store := NewScheduleStore(db)

    schedule := &AllowanceSchedule{
        ChildID:     1,
        ParentID:    1,
        AmountCents: 1000,
        Frequency:   FrequencyWeekly,
        DayOfWeek:   intPtr(5), // Friday
    }

    created, err := store.Create(schedule)
    require.NoError(t, err)
    assert.Equal(t, int64(1000), created.AmountCents)
    assert.Equal(t, FrequencyWeekly, created.Frequency)
    assert.NotNil(t, created.NextRunAt)
}

// Handler test pattern
func TestHandler_CreateSchedule(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

    body := `{"child_id":1,"amount_cents":1000,"frequency":"weekly","day_of_week":5}`
    req := httptest.NewRequest("POST", "/api/schedules", strings.NewReader(body))
    req = req.WithContext(middleware.WithSession(req.Context(), parentSession))

    w := httptest.NewRecorder()
    handler.HandleCreateSchedule(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
}

// Scheduler test pattern
func TestScheduler_ProcessDueSchedules(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Create a schedule that's due
    schedStore := store.NewScheduleStore(db)
    schedule := createTestSchedule(t, db, time.Now().Add(-time.Hour))

    scheduler := NewScheduler(schedStore, store.NewTransactionStore(db), store.NewChildStore(db))
    scheduler.ProcessDueSchedules()

    // Verify transaction was created
    txns, _ := store.NewTransactionStore(db).ListByChild(schedule.ChildID)
    assert.Len(t, txns, 1)
    assert.Equal(t, schedule.AmountCents, txns[0].AmountCents)
    assert.Equal(t, "allowance", string(txns[0].TransactionType))
}
```

## Common Validation Rules

| Field | Rule | Error Code |
|-------|------|------------|
| amount_cents | 1 to 99999999 | invalid_amount |
| frequency | 'weekly', 'biweekly', 'monthly' | invalid_frequency |
| day_of_week | 0-6 (required for weekly/biweekly) | invalid_day |
| day_of_month | 1-31 (required for monthly) | invalid_day |
| note | max 500 chars, optional | invalid_note |
| child_id | must be in parent's family | not_found |

## File Locations

| Component | Path |
|-----------|------|
| Schedule store | `backend/internal/store/schedule.go` |
| Schedule store tests | `backend/internal/store/schedule_test.go` |
| API handlers | `backend/internal/allowance/handler.go` |
| Handler tests | `backend/internal/allowance/handler_test.go` |
| Background scheduler | `backend/internal/allowance/scheduler.go` |
| Scheduler tests | `backend/internal/allowance/scheduler_test.go` |
| API routes (main) | `backend/main.go` |
| Frontend types | `frontend/src/types.ts` |
| Frontend API | `frontend/src/api.ts` |
| Schedule list component | `frontend/src/components/ScheduleList.tsx` |
| Schedule form component | `frontend/src/components/ScheduleForm.tsx` |
| Upcoming allowance display | `frontend/src/components/UpcomingAllowance.tsx` |

## Migration

```sql
-- Run via migration or init script
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

-- Add schedule reference to transactions
ALTER TABLE transactions ADD COLUMN schedule_id INTEGER REFERENCES allowance_schedules(id) ON DELETE SET NULL;
```
