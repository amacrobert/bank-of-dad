package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intP(i int) *int       { return &i }
func strP(s string) *string { return &s }

func createTestSchedule(t *testing.T, db *DB, childID, parentID int64) *AllowanceSchedule {
	t.Helper()
	ss := NewScheduleStore(db)
	nextRun := time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC)
	s := &AllowanceSchedule{
		ChildID:     childID,
		ParentID:    parentID,
		AmountCents: 1000,
		Frequency:   FrequencyWeekly,
		DayOfWeek:   intP(5), // Friday
		Status:      ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}
	created, err := ss.Create(s)
	require.NoError(t, err)
	return created
}

func TestScheduleStore_Create(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC)
	s := &AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   FrequencyWeekly,
		DayOfWeek:   intP(5),
		Note:        strP("Weekly allowance"),
		Status:      ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}

	created, err := ss.Create(s)
	require.NoError(t, err)
	assert.True(t, created.ID > 0)
	assert.Equal(t, child.ID, created.ChildID)
	assert.Equal(t, parent.ID, created.ParentID)
	assert.Equal(t, int64(1000), created.AmountCents)
	assert.Equal(t, FrequencyWeekly, created.Frequency)
	assert.Equal(t, 5, *created.DayOfWeek)
	assert.Nil(t, created.DayOfMonth)
	assert.Equal(t, "Weekly allowance", *created.Note)
	assert.Equal(t, ScheduleStatusActive, created.Status)
	assert.NotNil(t, created.NextRunAt)
}

func TestScheduleStore_Create_Monthly(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	s := &AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 5000,
		Frequency:   FrequencyMonthly,
		DayOfMonth:  intP(15),
		Status:      ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}

	created, err := ss.Create(s)
	require.NoError(t, err)
	assert.Equal(t, FrequencyMonthly, created.Frequency)
	assert.Equal(t, 15, *created.DayOfMonth)
	assert.Nil(t, created.DayOfWeek)
}

func TestScheduleStore_GetByID(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	fetched, err := ss.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.AmountCents, fetched.AmountCents)
	assert.Equal(t, created.Frequency, fetched.Frequency)
}

func TestScheduleStore_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fetched, err := ss.GetByID(999)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestScheduleStore_ListByParentFamily(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child1 := createTestChild(t, db, fam.ID)
	child2, err := cs.Create(fam.ID, "Child2", "pass123")
	require.NoError(t, err)

	createTestSchedule(t, db, child1.ID, parent.ID)
	createTestSchedule(t, db, child2.ID, parent.ID)

	schedules, err := ss.ListByParentFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, schedules, 2)
	// Should have child name joined
	assert.NotEmpty(t, schedules[0].ChildFirstName)
}

func TestScheduleStore_Update(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Update amount
	created.AmountCents = 2000
	newNote := "Updated allowance"
	created.Note = &newNote
	updated, err := ss.Update(created)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), updated.AmountCents)
	assert.Equal(t, "Updated allowance", *updated.Note)
}

func TestScheduleStore_Delete(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	err := ss.Delete(created.ID)
	require.NoError(t, err)

	fetched, err := ss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestScheduleStore_UpdateNextRunAt(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	newTime := time.Date(2026, time.February, 14, 0, 0, 0, 0, time.UTC)
	err := ss.UpdateNextRunAt(created.ID, newTime)
	require.NoError(t, err)

	fetched, err := ss.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.NextRunAt)
	assert.Equal(t, newTime.UTC(), fetched.NextRunAt.UTC())
}

func TestScheduleStore_UpdateStatus_Pause(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	err := ss.UpdateStatus(created.ID, ScheduleStatusPaused)
	require.NoError(t, err)

	fetched, err := ss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, ScheduleStatusPaused, fetched.Status)
}

func TestScheduleStore_UpdateStatus_Resume(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Pause then resume
	err := ss.UpdateStatus(created.ID, ScheduleStatusPaused)
	require.NoError(t, err)
	err = ss.UpdateStatus(created.ID, ScheduleStatusActive)
	require.NoError(t, err)

	fetched, err := ss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, ScheduleStatusActive, fetched.Status)
}

func TestScheduleStore_ListDue(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child1 := createTestChild(t, db, fam.ID)
	child2, err := cs.Create(fam.ID, "Child2", "pass123")
	require.NoError(t, err)

	// Create schedule due in the past (child1)
	pastTime := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	s := &AllowanceSchedule{
		ChildID:     child1.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   FrequencyWeekly,
		DayOfWeek:   intP(5),
		Status:      ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	_, err = ss.Create(s)
	require.NoError(t, err)

	// Create schedule due in the future (child2)
	futureTime := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)
	s2 := &AllowanceSchedule{
		ChildID:     child2.ID,
		ParentID:    parent.ID,
		AmountCents: 2000,
		Frequency:   FrequencyWeekly,
		DayOfWeek:   intP(1),
		Status:      ScheduleStatusActive,
		NextRunAt:   &futureTime,
	}
	_, err = ss.Create(s2)
	require.NoError(t, err)

	now := time.Date(2026, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := ss.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 1)
	assert.Equal(t, int64(1000), due[0].AmountCents)
}

func TestScheduleStore_ListDue_SkipsPaused(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	pastTime := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	s := &AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   FrequencyWeekly,
		DayOfWeek:   intP(5),
		Status:      ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := ss.Create(s)
	require.NoError(t, err)

	// Pause it
	err = ss.UpdateStatus(created.ID, ScheduleStatusPaused)
	require.NoError(t, err)

	now := time.Date(2026, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := ss.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 0)
}

func TestScheduleStore_ListActiveByChild(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Active schedule should be returned
	active, err := ss.ListActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, int64(1000), active[0].AmountCents)

	// Pause the schedule — should no longer appear in active list
	err = ss.UpdateStatus(created.ID, ScheduleStatusPaused)
	require.NoError(t, err)

	active, err = ss.ListActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, active, 0)
}

// T007 (006): Test for ScheduleStore.GetByChildID

func TestScheduleStore_GetByChildID(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// No schedule yet — should return nil
	fetched, err := ss.GetByChildID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)

	// Create a schedule
	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Should return the schedule
	fetched, err = ss.GetByChildID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, child.ID, fetched.ChildID)
	assert.Equal(t, int64(1000), fetched.AmountCents)
	assert.Equal(t, FrequencyWeekly, fetched.Frequency)
}

func TestScheduleStore_GetByChildID_ReturnsPausedToo(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Pause it
	err := ss.UpdateStatus(created.ID, ScheduleStatusPaused)
	require.NoError(t, err)

	// GetByChildID should still return it (any status)
	fetched, err := ss.GetByChildID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, ScheduleStatusPaused, fetched.Status)
}

// T040: Verify ON DELETE CASCADE removes schedules when child is deleted
func TestScheduleStore_CascadeDeleteOnChildRemoval(t *testing.T) {
	db := testDB(t)
	ss := NewScheduleStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Create schedule for the child
	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Verify schedule exists
	sched, err := ss.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, sched)

	// Delete the child
	_, err = db.Write.Exec("DELETE FROM children WHERE id = ?", child.ID)
	require.NoError(t, err)

	// Verify child is gone
	deletedChild, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, deletedChild)

	// Verify schedule was CASCADE deleted
	deletedSched, err := ss.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, deletedSched, "Schedule should be deleted when child is deleted (CASCADE)")
}
