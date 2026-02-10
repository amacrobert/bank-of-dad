package allowance

import (
	"testing"
	"time"

	"bank-of-dad/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================
// T011: Tests for Scheduler.ProcessDueSchedules
// =====================================================

func TestScheduler_ProcessDueSchedules_CreatesAllowanceTransaction(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	// Create a schedule that's already due
	pastTime := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	note := "Weekly allowance"
	sched := &store.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   store.FrequencyWeekly,
		DayOfWeek:   intPtr(5),
		Note:        &note,
		Status:      store.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := schedStore.Create(sched)
	require.NoError(t, err)

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// Verify transaction was created
	txns, err := txStore.ListByChild(child.ID)
	require.NoError(t, err)
	require.Len(t, txns, 1)
	assert.Equal(t, int64(1000), txns[0].AmountCents)
	assert.Equal(t, store.TransactionTypeAllowance, txns[0].TransactionType)
	assert.NotNil(t, txns[0].ScheduleID)
	assert.Equal(t, created.ID, *txns[0].ScheduleID)
	assert.Equal(t, "Weekly allowance", *txns[0].Note)

	// Verify child balance was updated
	balance, err := childStore.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), balance)
}

func TestScheduler_ProcessDueSchedules_AdvancesNextRunAt(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	pastTime := time.Date(2026, time.January, 30, 0, 0, 0, 0, time.UTC) // Friday
	sched := &store.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   store.FrequencyWeekly,
		DayOfWeek:   intPtr(5), // Friday
		Status:      store.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := schedStore.Create(sched)
	require.NoError(t, err)

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// Verify next_run_at was advanced to next Friday (Jan 30 + 7 = Feb 6)
	updated, err := schedStore.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.NextRunAt)
	assert.Equal(t, time.Date(2026, time.February, 6, 0, 0, 0, 0, time.UTC), updated.NextRunAt.UTC())
}

func TestScheduler_ProcessDueSchedules_SkipsPaused(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	pastTime := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	sched := &store.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   store.FrequencyWeekly,
		DayOfWeek:   intPtr(5),
		Status:      store.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := schedStore.Create(sched)
	require.NoError(t, err)

	// Pause the schedule
	err = schedStore.UpdateStatus(created.ID, store.ScheduleStatusPaused)
	require.NoError(t, err)

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// No transaction should have been created
	txns, err := txStore.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, txns, 0)
}

func TestScheduler_ProcessDueSchedules_MultipleDue(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child1 := createTestChild(t, db, family.ID, "Emma")
	child2 := createTestChild(t, db, family.ID, "Liam")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	pastTime := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)

	// Create due schedules for two different children (one per child)
	for _, tc := range []struct {
		childID int64
		amount  int64
	}{
		{child1.ID, 1000},
		{child2.ID, 2000},
	} {
		sched := &store.AllowanceSchedule{
			ChildID:     tc.childID,
			ParentID:    parent.ID,
			AmountCents: tc.amount,
			Frequency:   store.FrequencyWeekly,
			DayOfWeek:   intPtr(5),
			Status:      store.ScheduleStatusActive,
			NextRunAt:   &pastTime,
		}
		_, err := schedStore.Create(sched)
		require.NoError(t, err)
	}

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// Each child should have one transaction
	txns1, err := txStore.ListByChild(child1.ID)
	require.NoError(t, err)
	assert.Len(t, txns1, 1)

	txns2, err := txStore.ListByChild(child2.ID)
	require.NoError(t, err)
	assert.Len(t, txns2, 1)

	// Balances should match their allowance amounts
	balance1, err := childStore.GetBalance(child1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), balance1)

	balance2, err := childStore.GetBalance(child2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), balance2)
}

func TestScheduler_ProcessDueSchedules_HandlesMissed(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	// Schedule that was due far in the past (simulating downtime)
	pastTime := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	sched := &store.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 500,
		Frequency:   store.FrequencyWeekly,
		DayOfWeek:   intPtr(1),
		Status:      store.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	_, err := schedStore.Create(sched)
	require.NoError(t, err)

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// Transaction should still be created
	txns, err := txStore.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, txns, 1)
	assert.Equal(t, int64(500), txns[0].AmountCents)
}

// =====================================================
// T026: Tests for biweekly and monthly scheduler execution
// =====================================================

func TestScheduler_ProcessDueSchedules_BiweeklyAdvances14Days(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	pastTime := time.Date(2026, time.January, 19, 0, 0, 0, 0, time.UTC) // Monday
	sched := &store.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 2000,
		Frequency:   store.FrequencyBiweekly,
		DayOfWeek:   intPtr(1), // Monday
		Status:      store.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := schedStore.Create(sched)
	require.NoError(t, err)

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// Verify next_run_at advanced by 14 days
	updated, err := schedStore.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.NextRunAt)
	assert.Equal(t, time.Date(2026, time.February, 2, 0, 0, 0, 0, time.UTC), updated.NextRunAt.UTC())

	// Verify transaction created
	txns, err := txStore.ListByChild(child.ID)
	require.NoError(t, err)
	require.Len(t, txns, 1)
	assert.Equal(t, int64(2000), txns[0].AmountCents)
}

func TestScheduler_ProcessDueSchedules_MonthlyAdvancesToNextMonth(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	pastTime := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	sched := &store.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 5000,
		Frequency:   store.FrequencyMonthly,
		DayOfMonth:  intPtr(15),
		Status:      store.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := schedStore.Create(sched)
	require.NoError(t, err)

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// Verify next_run_at advanced to Feb 15
	updated, err := schedStore.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.NextRunAt)
	assert.Equal(t, time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC), updated.NextRunAt.UTC())
}

func TestScheduler_ProcessDueSchedules_Monthly31stClampsToFeb28(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	schedStore := store.NewScheduleStore(db)
	txStore := store.NewTransactionStore(db)
	childStore := store.NewChildStore(db)

	// Schedule due on Jan 31
	pastTime := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	sched := &store.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 5000,
		Frequency:   store.FrequencyMonthly,
		DayOfMonth:  intPtr(31),
		Status:      store.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := schedStore.Create(sched)
	require.NoError(t, err)

	scheduler := NewScheduler(schedStore, txStore, childStore)
	scheduler.ProcessDueSchedules()

	// Verify next_run_at clamped to Feb 28 (2026 is not a leap year)
	updated, err := schedStore.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.NextRunAt)
	assert.Equal(t, time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC), updated.NextRunAt.UTC())
}
