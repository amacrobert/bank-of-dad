package repositories

import (
	"testing"
	"time"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func intP(i int) *int       { return &i }
func strP(s string) *string { return &s }

func createTestFamily(t *testing.T, db *gorm.DB) *models.Family {
	t.Helper()
	fr := NewFamilyRepo(db)
	f, err := fr.Create("test-family")
	require.NoError(t, err)
	return f
}

func createTestParent(t *testing.T, db *gorm.DB, familyID int64) *models.Parent {
	t.Helper()
	pr := NewParentRepo(db)
	p, err := pr.Create("google-test-123", "test@example.com", "Test Parent")
	require.NoError(t, err)
	err = pr.SetFamilyID(p.ID, familyID)
	require.NoError(t, err)
	p.FamilyID = familyID
	return p
}

func createTestChild(t *testing.T, db *gorm.DB, familyID int64) *models.Child {
	t.Helper()
	cr := NewChildRepo(db)
	c, err := cr.Create(familyID, "TestChild", "password123", nil)
	require.NoError(t, err)
	return c
}

func createTestSchedule(t *testing.T, db *gorm.DB, childID, parentID int64) *models.AllowanceSchedule {
	t.Helper()
	sr := NewScheduleRepo(db)
	nextRun := time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC)
	s := &models.AllowanceSchedule{
		ChildID:     childID,
		ParentID:    parentID,
		AmountCents: 1000,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   intP(5), // Friday
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}
	created, err := sr.Create(s)
	require.NoError(t, err)
	return created
}

func TestScheduleRepo_Create(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC)
	s := &models.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   intP(5),
		Note:        strP("Weekly allowance"),
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}

	created, err := sr.Create(s)
	require.NoError(t, err)
	assert.True(t, created.ID > 0)
	assert.Equal(t, child.ID, created.ChildID)
	assert.Equal(t, parent.ID, created.ParentID)
	assert.Equal(t, int64(1000), created.AmountCents)
	assert.Equal(t, models.FrequencyWeekly, created.Frequency)
	assert.Equal(t, 5, *created.DayOfWeek)
	assert.Nil(t, created.DayOfMonth)
	assert.Equal(t, "Weekly allowance", *created.Note)
	assert.Equal(t, models.ScheduleStatusActive, created.Status)
	assert.NotNil(t, created.NextRunAt)
}

func TestScheduleRepo_Create_Monthly(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	nextRun := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	s := &models.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 5000,
		Frequency:   models.FrequencyMonthly,
		DayOfMonth:  intP(15),
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}

	created, err := sr.Create(s)
	require.NoError(t, err)
	assert.Equal(t, models.FrequencyMonthly, created.Frequency)
	assert.Equal(t, 15, *created.DayOfMonth)
	assert.Nil(t, created.DayOfWeek)
}

func TestScheduleRepo_GetByID(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	fetched, err := sr.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.AmountCents, fetched.AmountCents)
	assert.Equal(t, created.Frequency, fetched.Frequency)
}

func TestScheduleRepo_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fetched, err := sr.GetByID(999)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestScheduleRepo_ListByParentFamily(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)
	cr := NewChildRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child1 := createTestChild(t, db, fam.ID)
	child2, err := cr.Create(fam.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	createTestSchedule(t, db, child1.ID, parent.ID)
	createTestSchedule(t, db, child2.ID, parent.ID)

	schedules, err := sr.ListByParentFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, schedules, 2)
	// Should have child name joined
	assert.NotEmpty(t, schedules[0].ChildFirstName)
}

func TestScheduleRepo_Update(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Update amount
	created.AmountCents = 2000
	newNote := "Updated allowance"
	created.Note = &newNote
	updated, err := sr.Update(created)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), updated.AmountCents)
	assert.Equal(t, "Updated allowance", *updated.Note)
}

func TestScheduleRepo_Delete(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	err := sr.Delete(created.ID)
	require.NoError(t, err)

	fetched, err := sr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestScheduleRepo_UpdateNextRunAt(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	newTime := time.Date(2026, time.February, 14, 0, 0, 0, 0, time.UTC)
	err := sr.UpdateNextRunAt(created.ID, newTime)
	require.NoError(t, err)

	fetched, err := sr.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.NextRunAt)
	assert.Equal(t, newTime.UTC(), fetched.NextRunAt.UTC())
}

func TestScheduleRepo_UpdateStatus_Pause(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	err := sr.UpdateStatus(created.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	fetched, err := sr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ScheduleStatusPaused, fetched.Status)
}

func TestScheduleRepo_UpdateStatus_Resume(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Pause then resume
	err := sr.UpdateStatus(created.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)
	err = sr.UpdateStatus(created.ID, models.ScheduleStatusActive)
	require.NoError(t, err)

	fetched, err := sr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ScheduleStatusActive, fetched.Status)
}

func TestScheduleRepo_ListDue(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)
	cr := NewChildRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child1 := createTestChild(t, db, fam.ID)
	child2, err := cr.Create(fam.ID, "Child2", "pass123", nil)
	require.NoError(t, err)

	// Create schedule due in the past (child1)
	pastTime := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	s := &models.AllowanceSchedule{
		ChildID:     child1.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   intP(5),
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	_, err = sr.Create(s)
	require.NoError(t, err)

	// Create schedule due in the future (child2)
	futureTime := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)
	s2 := &models.AllowanceSchedule{
		ChildID:     child2.ID,
		ParentID:    parent.ID,
		AmountCents: 2000,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   intP(1),
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &futureTime,
	}
	_, err = sr.Create(s2)
	require.NoError(t, err)

	now := time.Date(2026, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := sr.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 1)
	assert.Equal(t, int64(1000), due[0].AmountCents)
}

func TestScheduleRepo_ListDue_SkipsPaused(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	pastTime := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	s := &models.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   intP(5),
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &pastTime,
	}
	created, err := sr.Create(s)
	require.NoError(t, err)

	// Pause it
	err = sr.UpdateStatus(created.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	now := time.Date(2026, time.February, 5, 12, 0, 0, 0, time.UTC)
	due, err := sr.ListDue(now)
	require.NoError(t, err)
	assert.Len(t, due, 0)
}

func TestScheduleRepo_ListActiveByChild(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Active schedule should be returned
	active, err := sr.ListActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, int64(1000), active[0].AmountCents)

	// Pause the schedule — should no longer appear in active list
	err = sr.UpdateStatus(created.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	active, err = sr.ListActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, active, 0)
}

func TestScheduleRepo_GetByChildID(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// No schedule yet — should return nil
	fetched, err := sr.GetByChildID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)

	// Create a schedule
	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Should return the schedule
	fetched, err = sr.GetByChildID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, child.ID, fetched.ChildID)
	assert.Equal(t, int64(1000), fetched.AmountCents)
	assert.Equal(t, models.FrequencyWeekly, fetched.Frequency)
}

func TestScheduleRepo_GetByChildID_ReturnsPausedToo(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Pause it
	err := sr.UpdateStatus(created.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	// GetByChildID should still return it (any status)
	fetched, err := sr.GetByChildID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, models.ScheduleStatusPaused, fetched.Status)
}

func TestScheduleRepo_ListAllActiveWithTimezone(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)
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
	nextRun := time.Date(2026, time.February, 7, 5, 0, 0, 0, time.UTC)
	active := &models.AllowanceSchedule{
		ChildID:     child1.ID,
		ParentID:    parent.ID,
		AmountCents: 1000,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   intP(5),
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}
	_, err = sr.Create(active)
	require.NoError(t, err)

	// Create a paused schedule for child2 — should NOT be returned
	paused := &models.AllowanceSchedule{
		ChildID:     child2.ID,
		ParentID:    parent.ID,
		AmountCents: 2000,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   intP(1),
		Status:      models.ScheduleStatusActive,
		NextRunAt:   &nextRun,
	}
	created2, err := sr.Create(paused)
	require.NoError(t, err)
	err = sr.UpdateStatus(created2.ID, models.ScheduleStatusPaused)
	require.NoError(t, err)

	results, err := sr.ListAllActiveWithTimezone()
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, int64(1000), results[0].AmountCents)
	assert.Equal(t, "America/New_York", results[0].FamilyTimezone)
}

func TestScheduleRepo_ListAllActiveWithTimezone_Empty(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	results, err := sr.ListAllActiveWithTimezone()
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestScheduleRepo_CascadeDeleteOnChildRemoval(t *testing.T) {
	db := testDB(t)
	sr := NewScheduleRepo(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Create schedule for the child
	created := createTestSchedule(t, db, child.ID, parent.ID)

	// Verify schedule exists
	sched, err := sr.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, sched)

	// Delete the child
	db.Exec("DELETE FROM children WHERE id = ?", child.ID)

	// Verify schedule was CASCADE deleted
	deletedSched, err := sr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, deletedSched, "Schedule should be deleted when child is deleted (CASCADE)")
}
