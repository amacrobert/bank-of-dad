package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestInterestSchedule(t *testing.T, db *sql.DB, childID, parentID int64) *InterestSchedule {
	t.Helper()
	iss := NewInterestScheduleStore(db)
	nextRun := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	s := &InterestSchedule{
		ChildID:    childID,
		ParentID:   parentID,
		Frequency:  FrequencyMonthly,
		DayOfMonth: intP(15),
		Status:     ScheduleStatusActive,
		NextRunAt:  &nextRun,
	}
	created, err := iss.Create(s)
	require.NoError(t, err)
	return created
}

// T004: InterestScheduleStore CRUD tests

func TestInterestScheduleStore_Create_Monthly(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	s := &InterestSchedule{
		ChildID:    child.ID,
		ParentID:   parent.ID,
		Frequency:  FrequencyMonthly,
		DayOfMonth: intP(15),
		Status:     ScheduleStatusActive,
		NextRunAt:  &nextRun,
	}

	created, err := iss.Create(s)
	require.NoError(t, err)
	assert.True(t, created.ID > 0)
	assert.Equal(t, child.ID, created.ChildID)
	assert.Equal(t, parent.ID, created.ParentID)
	assert.Equal(t, FrequencyMonthly, created.Frequency)
	assert.Equal(t, 15, *created.DayOfMonth)
	assert.Nil(t, created.DayOfWeek)
	assert.Equal(t, ScheduleStatusActive, created.Status)
	assert.NotNil(t, created.NextRunAt)
}

func TestInterestScheduleStore_Create_Weekly(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)
	s := &InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: FrequencyWeekly,
		DayOfWeek: intP(5), // Friday
		Status:    ScheduleStatusActive,
		NextRunAt: &nextRun,
	}

	created, err := iss.Create(s)
	require.NoError(t, err)
	assert.Equal(t, FrequencyWeekly, created.Frequency)
	assert.Equal(t, 5, *created.DayOfWeek)
	assert.Nil(t, created.DayOfMonth)
}

func TestInterestScheduleStore_Create_Biweekly(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)
	s := &InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: FrequencyBiweekly,
		DayOfWeek: intP(3), // Wednesday
		Status:    ScheduleStatusActive,
		NextRunAt: &nextRun,
	}

	created, err := iss.Create(s)
	require.NoError(t, err)
	assert.Equal(t, FrequencyBiweekly, created.Frequency)
	assert.Equal(t, 3, *created.DayOfWeek)
}

func TestInterestScheduleStore_Create_UniqueChild(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Create first schedule
	createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Second schedule for same child should fail (UNIQUE constraint on child_id)
	nextRun := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	s2 := &InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: FrequencyWeekly,
		DayOfWeek: intP(1),
		Status:    ScheduleStatusActive,
		NextRunAt: &nextRun,
	}
	_, err := iss.Create(s2)
	assert.Error(t, err, "should fail due to UNIQUE constraint on child_id")
}

func TestInterestScheduleStore_GetByChildID(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	fetched, err := iss.GetByChildID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, child.ID, fetched.ChildID)
	assert.Equal(t, FrequencyMonthly, fetched.Frequency)
	assert.Equal(t, 15, *fetched.DayOfMonth)
}

func TestInterestScheduleStore_GetByChildID_NotFound(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fetched, err := iss.GetByChildID(999)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestInterestScheduleStore_Update(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Change from monthly to weekly
	created.Frequency = FrequencyWeekly
	created.DayOfWeek = intP(2) // Tuesday
	created.DayOfMonth = nil
	newNext := time.Date(2025, time.January, 14, 0, 0, 0, 0, time.UTC)
	created.NextRunAt = &newNext

	updated, err := iss.Update(created)
	require.NoError(t, err)
	assert.Equal(t, FrequencyWeekly, updated.Frequency)
	assert.Equal(t, 2, *updated.DayOfWeek)
	assert.Nil(t, updated.DayOfMonth)
}

func TestInterestScheduleStore_Delete(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	err := iss.Delete(created.ID)
	require.NoError(t, err)

	fetched, err := iss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestInterestScheduleStore_UpdateStatus(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Pause
	err := iss.UpdateStatus(created.ID, ScheduleStatusPaused)
	require.NoError(t, err)

	fetched, err := iss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, ScheduleStatusPaused, fetched.Status)

	// Resume
	err = iss.UpdateStatus(created.ID, ScheduleStatusActive)
	require.NoError(t, err)

	fetched, err = iss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, ScheduleStatusActive, fetched.Status)
}

// T005: InterestScheduleStore ListDue and UpdateNextRunAt tests

func TestInterestScheduleStore_ListDue(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)

	cs := NewChildStore(db)
	child1, err := cs.Create(fam.ID, "Child1", "pass123", nil)
	require.NoError(t, err)
	child2, err := cs.Create(fam.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	// Child1: schedule due in the past
	pastTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	s1 := &InterestSchedule{
		ChildID:    child1.ID,
		ParentID:   parent.ID,
		Frequency:  FrequencyMonthly,
		DayOfMonth: intP(1),
		Status:     ScheduleStatusActive,
		NextRunAt:  &pastTime,
	}
	_, err = iss.Create(s1)
	require.NoError(t, err)

	// Child2: schedule due in the future
	futureTime := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)
	s2 := &InterestSchedule{
		ChildID:    child2.ID,
		ParentID:   parent.ID,
		Frequency:  FrequencyMonthly,
		DayOfMonth: intP(31),
		Status:     ScheduleStatusActive,
		NextRunAt:  &futureTime,
	}
	_, err = iss.Create(s2)
	require.NoError(t, err)

	now := time.Date(2025, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := iss.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 1)
	assert.Equal(t, child1.ID, due[0].ChildID)
}

func TestInterestScheduleStore_ListDue_SkipsPaused(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	pastTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	s := &InterestSchedule{
		ChildID:    child.ID,
		ParentID:   parent.ID,
		Frequency:  FrequencyMonthly,
		DayOfMonth: intP(1),
		Status:     ScheduleStatusActive,
		NextRunAt:  &pastTime,
	}
	created, err := iss.Create(s)
	require.NoError(t, err)

	// Pause it
	err = iss.UpdateStatus(created.ID, ScheduleStatusPaused)
	require.NoError(t, err)

	now := time.Date(2025, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := iss.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 0)
}

func TestInterestScheduleStore_ListDue_Empty(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	now := time.Date(2025, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := iss.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 0)
}

func TestInterestScheduleStore_UpdateNextRunAt(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	newTime := time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC)
	err := iss.UpdateNextRunAt(created.ID, newTime)
	require.NoError(t, err)

	fetched, err := iss.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.NextRunAt)
	assert.Equal(t, newTime.UTC(), fetched.NextRunAt.UTC())
}

func TestInterestScheduleStore_ListAllActiveWithTimezone(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)
	fs := NewFamilyStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	err := fs.UpdateTimezone(fam.ID, "America/New_York")
	require.NoError(t, err)
	parent := createTestParent(t, db, fam.ID)
	child1 := createTestChild(t, db, fam.ID)
	child2, err := cs.Create(fam.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	// Create an active schedule for child1
	nextRun := time.Date(2025, time.January, 15, 5, 0, 0, 0, time.UTC)
	s1 := &InterestSchedule{
		ChildID:    child1.ID,
		ParentID:   parent.ID,
		Frequency:  FrequencyMonthly,
		DayOfMonth: intP(15),
		Status:     ScheduleStatusActive,
		NextRunAt:  &nextRun,
	}
	_, err = iss.Create(s1)
	require.NoError(t, err)

	// Create a paused schedule for child2 — should NOT be returned
	s2 := &InterestSchedule{
		ChildID:    child2.ID,
		ParentID:   parent.ID,
		Frequency:  FrequencyWeekly,
		DayOfWeek:  intP(5),
		Status:     ScheduleStatusActive,
		NextRunAt:  &nextRun,
	}
	created2, err := iss.Create(s2)
	require.NoError(t, err)
	err = iss.UpdateStatus(created2.ID, ScheduleStatusPaused)
	require.NoError(t, err)

	results, err := iss.ListAllActiveWithTimezone()
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, child1.ID, results[0].ChildID)
	assert.Equal(t, "America/New_York", results[0].FamilyTimezone)
}

func TestInterestScheduleStore_ListAllActiveWithTimezone_Empty(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	results, err := iss.ListAllActiveWithTimezone()
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestInterestScheduleStore_CascadeDelete(t *testing.T) {
	db := testDB(t)
	iss := NewInterestScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Delete the child
	_, err := db.Exec("DELETE FROM children WHERE id = $1", child.ID)
	require.NoError(t, err)

	// Interest schedule should be CASCADE deleted
	fetched, err := iss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched, "interest schedule should be deleted when child is deleted (CASCADE)")
}
