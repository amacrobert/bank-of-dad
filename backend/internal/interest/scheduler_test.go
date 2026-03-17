package interest

import (
	"testing"
	"time"

	"bank-of-dad/internal/testutil"
	"bank-of-dad/models"
	"bank-of-dad/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// =====================================================
// Tests for loadTimezone
// =====================================================

func TestLoadTimezone_ValidTimezone(t *testing.T) {
	loc := loadTimezone("America/New_York")
	expected, _ := time.LoadLocation("America/New_York")
	assert.Equal(t, expected, loc)
}

func TestLoadTimezone_EmptyString(t *testing.T) {
	loc := loadTimezone("")
	assert.Equal(t, time.UTC, loc)
}

func TestLoadTimezone_InvalidTimezone(t *testing.T) {
	loc := loadTimezone("Not/A/Timezone")
	assert.Equal(t, time.UTC, loc)
}

// T017: Tests for ProcessDue

func TestProcessDue_AppliesInterest(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	// Set up: deposit $100 and set 10% rate
	_, _, err := txRepo.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child.ID, 1000)
	require.NoError(t, err)

	scheduler := NewScheduler(interestRepo)
	scheduler.ProcessDue()

	// Verify balance increased: 10000 * 1000 / 12 / 10000 = 83.33 → 83
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10083), balance)

	// Verify interest transaction was created
	txns, err := txRepo.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, txns, 2) // deposit + interest
	assert.Equal(t, models.TransactionTypeInterest, txns[0].TransactionType)
}

func TestProcessDue_SkipsAlreadyAccruedThisMonth(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	_, _, err := txRepo.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child.ID, 1000)
	require.NoError(t, err)

	scheduler := NewScheduler(interestRepo)

	// First accrual
	scheduler.ProcessDue()
	balanceAfterFirst, err := cs.GetBalance(child.ID)
	require.NoError(t, err)

	// Second accrual — should be skipped (same month)
	scheduler.ProcessDue()
	balanceAfterSecond, err := cs.GetBalance(child.ID)
	require.NoError(t, err)

	assert.Equal(t, balanceAfterFirst, balanceAfterSecond, "balance should not change on duplicate accrual")
}

func TestProcessDue_SkipsZeroBalance(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	cs := repositories.NewChildRepo(db)

	// Set rate but no balance
	err := interestRepo.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	scheduler := NewScheduler(interestRepo)
	scheduler.ProcessDue()

	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), balance, "balance should remain zero")
}

func TestProcessDue_SkipsZeroRate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	// Deposit but no rate set (default 0)
	_, _, err := txRepo.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)

	scheduler := NewScheduler(interestRepo)
	scheduler.ProcessDue()

	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000), balance, "balance should not change with zero rate")
}

func TestProcessDue_MultipleChildren(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	cs := repositories.NewChildRepo(db)
	child1, err := cs.Create(family.ID, "Child1", "pass123", nil)
	require.NoError(t, err)
	child2, err := cs.Create(family.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	interestRepo := repositories.NewInterestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)

	// Both children have balance and rate
	_, _, err = txRepo.Deposit(child1.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	_, _, err = txRepo.Deposit(child2.ID, parent.ID, 20000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child1.ID, 500) // 5%
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child2.ID, 1000) // 10%
	require.NoError(t, err)

	scheduler := NewScheduler(interestRepo)
	scheduler.ProcessDue()

	// Child1: 10000 * 500 / 12 / 10000 = 41.67 → 42
	balance1, err := cs.GetBalance(child1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10042), balance1)

	// Child2: 20000 * 1000 / 12 / 10000 = 166.67 → 167
	balance2, err := cs.GetBalance(child2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(20167), balance2)
}

// T018: Test for partial failure

func TestProcessDue_PartialFailure(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	cs := repositories.NewChildRepo(db)
	child1, err := cs.Create(family.ID, "Child1", "pass123", nil)
	require.NoError(t, err)
	child2, err := cs.Create(family.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	interestRepo := repositories.NewInterestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)

	// Child1: $1 at 5% → interest rounds to 0, will be skipped (error from ApplyInterest)
	_, _, err = txRepo.Deposit(child1.ID, parent.ID, 100, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child1.ID, 500)
	require.NoError(t, err)

	// Child2: $100 at 5% → 42 cents interest, should succeed
	_, _, err = txRepo.Deposit(child2.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child2.ID, 500)
	require.NoError(t, err)

	scheduler := NewScheduler(interestRepo)
	scheduler.ProcessDue()

	// Child1 should be unchanged (interest rounds to 0)
	balance1, err := cs.GetBalance(child1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100), balance1)

	// Child2 should have interest applied
	balance2, err := cs.GetBalance(child2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10042), balance2)
}

func TestScheduler_StartAndStop(t *testing.T) {
	db := testutil.SetupTestDB(t)
	interestRepo := repositories.NewInterestRepo(db)

	scheduler := NewScheduler(interestRepo)
	stop := make(chan struct{})

	scheduler.Start(100*time.Millisecond, stop)

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop should not hang
	close(stop)
	time.Sleep(50 * time.Millisecond) // Give goroutine time to exit
}

// T024: Tests for schedule-based ProcessDueSchedules

func createTestInterestSchedule(t *testing.T, db *gorm.DB, childID, parentID int64, freq models.Frequency, dayOfWeek, dayOfMonth *int, nextRunAt time.Time) *models.InterestSchedule {
	t.Helper()
	iss := repositories.NewInterestScheduleRepo(db)
	sched := &models.InterestSchedule{
		ChildID:    childID,
		ParentID:   parentID,
		Frequency:  freq,
		DayOfWeek:  dayOfWeek,
		DayOfMonth: dayOfMonth,
		Status:     models.ScheduleStatusActive,
		NextRunAt:  &nextRunAt,
	}
	created, err := iss.Create(sched)
	require.NoError(t, err)
	return created
}

func TestProcessDueSchedules_WeeklyProration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	iss := repositories.NewInterestScheduleRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	// Deposit $1000, set 5% rate
	_, _, err := txRepo.Deposit(child.ID, parent.ID, 100000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Create weekly interest schedule due in the past
	dow := 5 // Friday
	pastDue := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC) // a past Friday
	createTestInterestSchedule(t, db, child.ID, parent.ID, models.FrequencyWeekly, &dow, nil, pastDue)

	scheduler := NewScheduler(interestRepo)
	scheduler.SetInterestScheduleStore(iss)
	scheduler.ProcessDueSchedules()

	// Weekly: 100000 * 500 / 52 / 10000 = 96.15 → 96
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100096), balance)
}

func TestProcessDueSchedules_BiweeklyProration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	iss := repositories.NewInterestScheduleRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	_, _, err := txRepo.Deposit(child.ID, parent.ID, 100000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	dow := 5
	pastDue := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	createTestInterestSchedule(t, db, child.ID, parent.ID, models.FrequencyBiweekly, &dow, nil, pastDue)

	scheduler := NewScheduler(interestRepo)
	scheduler.SetInterestScheduleStore(iss)
	scheduler.ProcessDueSchedules()

	// Biweekly: 100000 * 500 / 26 / 10000 = 192.31 → 192
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100192), balance)
}

func TestProcessDueSchedules_MonthlyProration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	iss := repositories.NewInterestScheduleRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	_, _, err := txRepo.Deposit(child.ID, parent.ID, 100000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	dom := 15
	pastDue := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	createTestInterestSchedule(t, db, child.ID, parent.ID, models.FrequencyMonthly, nil, &dom, pastDue)

	scheduler := NewScheduler(interestRepo)
	scheduler.SetInterestScheduleStore(iss)
	scheduler.ProcessDueSchedules()

	// Monthly: 100000 * 500 / 12 / 10000 = 416.67 → 417
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100417), balance)
}

func TestProcessDueSchedules_UpdatesNextRunAt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	iss := repositories.NewInterestScheduleRepo(db)
	txRepo := repositories.NewTransactionRepo(db)

	_, _, err := txRepo.Deposit(child.ID, parent.ID, 100000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	dow := 5 // Friday
	pastDue := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	sched := createTestInterestSchedule(t, db, child.ID, parent.ID, models.FrequencyWeekly, &dow, nil, pastDue)

	scheduler := NewScheduler(interestRepo)
	scheduler.SetInterestScheduleStore(iss)
	scheduler.ProcessDueSchedules()

	// Verify next_run_at was advanced
	updated, err := iss.GetByID(sched.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.NextRunAt)
	assert.True(t, updated.NextRunAt.After(pastDue), "next_run_at should be after the original due date")
}

func TestProcessDueSchedules_SkipsPaused(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	iss := repositories.NewInterestScheduleRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	_, _, err := txRepo.Deposit(child.ID, parent.ID, 100000, "")
	require.NoError(t, err)
	err = interestRepo.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	dow := 5
	pastDue := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	sched := createTestInterestSchedule(t, db, child.ID, parent.ID, models.FrequencyWeekly, &dow, nil, pastDue)

	// Pause it
	err = iss.UpdateStatus(sched.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	scheduler := NewScheduler(interestRepo)
	scheduler.SetInterestScheduleStore(iss)
	scheduler.ProcessDueSchedules()

	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100000), balance, "balance should be unchanged for paused schedule")
}

func TestProcessDueSchedules_SkipsZeroRate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	interestRepo := repositories.NewInterestRepo(db)
	iss := repositories.NewInterestScheduleRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	cs := repositories.NewChildRepo(db)

	_, _, err := txRepo.Deposit(child.ID, parent.ID, 100000, "")
	require.NoError(t, err)
	// No interest rate set (default 0)

	dow := 5
	pastDue := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	createTestInterestSchedule(t, db, child.ID, parent.ID, models.FrequencyWeekly, &dow, nil, pastDue)

	scheduler := NewScheduler(interestRepo)
	scheduler.SetInterestScheduleStore(iss)
	scheduler.ProcessDueSchedules()

	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100000), balance, "balance should be unchanged with zero rate")
}

// =====================================================
// Tests for RecalculateAllNextRuns
// =====================================================

func TestScheduler_RecalculateAllNextRuns(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	// Set family timezone to America/New_York
	fs := repositories.NewFamilyRepo(db)
	err := fs.UpdateTimezone(family.ID, "America/New_York")
	require.NoError(t, err)

	interestRepo := repositories.NewInterestRepo(db)
	iss := repositories.NewInterestScheduleRepo(db)

	// Create an active monthly schedule with UTC-midnight next_run_at
	dom := 15
	utcMidnight := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	createTestInterestSchedule(t, db, child.ID, parent.ID, models.FrequencyMonthly, nil, &dom, utcMidnight)

	scheduler := NewScheduler(interestRepo)
	scheduler.SetInterestScheduleStore(iss)
	scheduler.RecalculateAllNextRuns()

	// Verify next_run_at was recalculated to midnight in America/New_York
	results, err := iss.ListAllActiveWithTimezone()
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, results[0].NextRunAt)

	est, _ := time.LoadLocation("America/New_York")
	localTime := results[0].NextRunAt.In(est)
	assert.Equal(t, 0, localTime.Hour(), "should be midnight in family timezone")
	assert.Equal(t, 0, localTime.Minute())
}

func TestScheduler_RecalculateAllNextRuns_NilStore(t *testing.T) {
	db := testutil.SetupTestDB(t)
	interestRepo := repositories.NewInterestRepo(db)
	scheduler := NewScheduler(interestRepo)
	// interestScheduleRepo is nil — should not panic
	scheduler.RecalculateAllNextRuns()
}
