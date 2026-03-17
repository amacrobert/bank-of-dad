package repositories

import (
	"testing"
	"time"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createTestInterestSchedule(t *testing.T, db *gorm.DB, childID, parentID int64) *models.InterestSchedule {
	t.Helper()
	isr := NewInterestScheduleRepo(db)
	nextRun := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	s := &models.InterestSchedule{
		ChildID:    childID,
		ParentID:   parentID,
		Frequency:  models.FrequencyMonthly,
		DayOfMonth: intP(15),
		Status:     models.ScheduleStatusActive,
		NextRunAt:  &nextRun,
	}
	created, err := isr.Create(s)
	require.NoError(t, err)
	return created
}

func TestInterestScheduleRepo_Create_Monthly(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	s := &models.InterestSchedule{
		ChildID:    child.ID,
		ParentID:   parent.ID,
		Frequency:  models.FrequencyMonthly,
		DayOfMonth: intP(15),
		Status:     models.ScheduleStatusActive,
		NextRunAt:  &nextRun,
	}

	created, err := isr.Create(s)
	require.NoError(t, err)
	assert.True(t, created.ID > 0)
	assert.Equal(t, child.ID, created.ChildID)
	assert.Equal(t, parent.ID, created.ParentID)
	assert.Equal(t, models.FrequencyMonthly, created.Frequency)
	assert.Equal(t, 15, *created.DayOfMonth)
	assert.Nil(t, created.DayOfWeek)
	assert.Equal(t, models.ScheduleStatusActive, created.Status)
	assert.NotNil(t, created.NextRunAt)
}

func TestInterestScheduleRepo_Create_Weekly(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)
	s := &models.InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: models.FrequencyWeekly,
		DayOfWeek: intP(5), // Friday
		Status:    models.ScheduleStatusActive,
		NextRunAt: &nextRun,
	}

	created, err := isr.Create(s)
	require.NoError(t, err)
	assert.Equal(t, models.FrequencyWeekly, created.Frequency)
	assert.Equal(t, 5, *created.DayOfWeek)
	assert.Nil(t, created.DayOfMonth)
}

func TestInterestScheduleRepo_Create_Biweekly(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)
	s := &models.InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: models.FrequencyBiweekly,
		DayOfWeek: intP(3), // Wednesday
		Status:    models.ScheduleStatusActive,
		NextRunAt: &nextRun,
	}

	created, err := isr.Create(s)
	require.NoError(t, err)
	assert.Equal(t, models.FrequencyBiweekly, created.Frequency)
	assert.Equal(t, 3, *created.DayOfWeek)
}

func TestInterestScheduleRepo_Create_UniqueChild(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Create first schedule
	createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Second schedule for same child should fail (UNIQUE constraint on child_id)
	nextRun := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	s2 := &models.InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: models.FrequencyWeekly,
		DayOfWeek: intP(1),
		Status:    models.ScheduleStatusActive,
		NextRunAt: &nextRun,
	}
	_, err := isr.Create(s2)
	assert.Error(t, err, "should fail due to UNIQUE constraint on child_id")
}

func TestInterestScheduleRepo_GetByChildID(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	fetched, err := isr.GetByChildID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, child.ID, fetched.ChildID)
	assert.Equal(t, models.FrequencyMonthly, fetched.Frequency)
	assert.Equal(t, 15, *fetched.DayOfMonth)
}

func TestInterestScheduleRepo_GetByChildID_NotFound(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fetched, err := isr.GetByChildID(999)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestInterestScheduleRepo_Update(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Change from monthly to weekly
	created.Frequency = models.FrequencyWeekly
	created.DayOfWeek = intP(2) // Tuesday
	created.DayOfMonth = nil
	newNext := time.Date(2025, time.January, 14, 0, 0, 0, 0, time.UTC)
	created.NextRunAt = &newNext

	updated, err := isr.Update(created)
	require.NoError(t, err)
	assert.Equal(t, models.FrequencyWeekly, updated.Frequency)
	assert.Equal(t, 2, *updated.DayOfWeek)
	assert.Nil(t, updated.DayOfMonth)
}

func TestInterestScheduleRepo_Delete(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	err := isr.Delete(created.ID)
	require.NoError(t, err)

	fetched, err := isr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestInterestScheduleRepo_UpdateStatus(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Pause
	err := isr.UpdateStatus(created.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	fetched, err := isr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ScheduleStatusPaused, fetched.Status)

	// Resume
	err = isr.UpdateStatus(created.ID, models.ScheduleStatusActive)
	require.NoError(t, err)

	fetched, err = isr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ScheduleStatusActive, fetched.Status)
}

func TestInterestScheduleRepo_ListDue(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)

	cr := NewChildRepo(db)
	child1, err := cr.Create(fam.ID, "Child1", "pass123", nil)
	require.NoError(t, err)
	child2, err := cr.Create(fam.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	// Child1: schedule due in the past
	pastTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	s1 := &models.InterestSchedule{
		ChildID:    child1.ID,
		ParentID:   parent.ID,
		Frequency:  models.FrequencyMonthly,
		DayOfMonth: intP(1),
		Status:     models.ScheduleStatusActive,
		NextRunAt:  &pastTime,
	}
	_, err = isr.Create(s1)
	require.NoError(t, err)

	// Child2: schedule due in the future
	futureTime := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)
	s2 := &models.InterestSchedule{
		ChildID:    child2.ID,
		ParentID:   parent.ID,
		Frequency:  models.FrequencyMonthly,
		DayOfMonth: intP(31),
		Status:     models.ScheduleStatusActive,
		NextRunAt:  &futureTime,
	}
	_, err = isr.Create(s2)
	require.NoError(t, err)

	now := time.Date(2025, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := isr.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 1)
	assert.Equal(t, child1.ID, due[0].ChildID)
}

func TestInterestScheduleRepo_ListDue_SkipsPaused(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	pastTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	s := &models.InterestSchedule{
		ChildID:    child.ID,
		ParentID:   parent.ID,
		Frequency:  models.FrequencyMonthly,
		DayOfMonth: intP(1),
		Status:     models.ScheduleStatusActive,
		NextRunAt:  &pastTime,
	}
	created, err := isr.Create(s)
	require.NoError(t, err)

	// Pause it
	err = isr.UpdateStatus(created.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	now := time.Date(2025, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := isr.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 0)
}

func TestInterestScheduleRepo_ListDue_Empty(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	now := time.Date(2025, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := isr.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 0)
}

func TestInterestScheduleRepo_UpdateNextRunAt(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	newTime := time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC)
	err := isr.UpdateNextRunAt(created.ID, newTime)
	require.NoError(t, err)

	fetched, err := isr.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.NextRunAt)
	assert.Equal(t, newTime.UTC(), fetched.NextRunAt.UTC())
}

func TestInterestScheduleRepo_ListAllActiveWithTimezone(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)
	fr := NewFamilyRepo(db)
	cr := NewChildRepo(db)

	fam := createTestFamily(t, db)
	err := fr.UpdateTimezone(fam.ID, "America/New_York")
	require.NoError(t, err)
	parent := createTestParent(t, db, fam.ID)
	child1 := createTestChild(t, db, fam.ID)
	child2, err := cr.Create(fam.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	// Create an active schedule for child1
	nextRun := time.Date(2025, time.January, 15, 5, 0, 0, 0, time.UTC)
	s1 := &models.InterestSchedule{
		ChildID:    child1.ID,
		ParentID:   parent.ID,
		Frequency:  models.FrequencyMonthly,
		DayOfMonth: intP(15),
		Status:     models.ScheduleStatusActive,
		NextRunAt:  &nextRun,
	}
	_, err = isr.Create(s1)
	require.NoError(t, err)

	// Create a paused schedule for child2 — should NOT be returned
	s2 := &models.InterestSchedule{
		ChildID:   child2.ID,
		ParentID:  parent.ID,
		Frequency: models.FrequencyWeekly,
		DayOfWeek: intP(5),
		Status:    models.ScheduleStatusActive,
		NextRunAt: &nextRun,
	}
	created2, err := isr.Create(s2)
	require.NoError(t, err)
	err = isr.UpdateStatus(created2.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	results, err := isr.ListAllActiveWithTimezone()
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, child1.ID, results[0].ChildID)
	assert.Equal(t, "America/New_York", results[0].FamilyTimezone)
}

func TestInterestScheduleRepo_ListAllActiveWithTimezone_Empty(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	results, err := isr.ListAllActiveWithTimezone()
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestInterestScheduleRepo_CascadeDelete(t *testing.T) {
	db := testDB(t)
	isr := NewInterestScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestInterestSchedule(t, db, child.ID, parent.ID)

	// Delete the child
	db.Exec("DELETE FROM children WHERE id = ?", child.ID)

	// Interest schedule should be CASCADE deleted
	fetched, err := isr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched, "interest schedule should be deleted when child is deleted (CASCADE)")
}
