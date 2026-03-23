package chore

import (
	"testing"
	"time"

	"bank-of-dad/internal/testutil"
	"bank-of-dad/models"
	"bank-of-dad/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculatePeriodBounds_Daily(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 3, 15, 14, 30, 0, 0, loc)

	start, end := CalculatePeriodBounds(models.ChoreRecurrenceDaily, nil, nil, now, loc)

	assert.Equal(t, time.Date(2026, 3, 15, 0, 0, 0, 0, loc), start)
	assert.Equal(t, time.Date(2026, 3, 15, 23, 59, 59, 0, loc), end)
}

func TestCalculatePeriodBounds_Weekly(t *testing.T) {
	loc := time.UTC
	// 2026-03-15 is a Sunday (weekday 0)
	now := time.Date(2026, 3, 18, 10, 0, 0, 0, loc) // Wednesday

	dow := 1 // Monday
	start, end := CalculatePeriodBounds(models.ChoreRecurrenceWeekly, &dow, nil, now, loc)

	// Week starting Monday March 16
	assert.Equal(t, time.Date(2026, 3, 16, 0, 0, 0, 0, loc), start)
	assert.Equal(t, time.Date(2026, 3, 22, 23, 59, 59, 0, loc), end)
}

func TestCalculatePeriodBounds_Weekly_SameDayAsStart(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, loc) // Monday

	dow := 1 // Monday
	start, end := CalculatePeriodBounds(models.ChoreRecurrenceWeekly, &dow, nil, now, loc)

	assert.Equal(t, time.Date(2026, 3, 16, 0, 0, 0, 0, loc), start)
	assert.Equal(t, time.Date(2026, 3, 22, 23, 59, 59, 0, loc), end)
}

func TestCalculatePeriodBounds_Monthly(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 3, 15, 14, 30, 0, 0, loc)

	start, end := CalculatePeriodBounds(models.ChoreRecurrenceMonthly, nil, nil, now, loc)

	assert.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, loc), start)
	assert.Equal(t, time.Date(2026, 3, 31, 23, 59, 59, 0, loc), end)
}

func TestCalculatePeriodBounds_Timezone(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	// UTC time that's still "yesterday" in ET
	now := time.Date(2026, 3, 16, 3, 0, 0, 0, time.UTC) // 11pm ET on March 15

	start, end := CalculatePeriodBounds(models.ChoreRecurrenceDaily, nil, nil, now, loc)

	assert.Equal(t, time.Date(2026, 3, 15, 0, 0, 0, 0, loc), start)
	assert.Equal(t, time.Date(2026, 3, 15, 23, 59, 59, 0, loc), end)
}

func TestGenerateInstances_CreatesForRecurringChore(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	dow := 1
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Weekly Vacuum",
		RewardCents:       300,
		Recurrence:        models.ChoreRecurrenceWeekly,
		DayOfWeek:         &dow,
		IsActive:          true,
	})
	require.NoError(t, err)

	_, err = choreRepo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: chore.ID,
		ChildID: child.ID,
	})
	require.NoError(t, err)

	scheduler := NewScheduler(choreRepo, instanceRepo, repositories.NewChildRepo(db), repositories.NewFamilyRepo(db))
	scheduler.GenerateInstances()

	// Verify instance was created
	available, _, _, err := instanceRepo.ListByChild(child.ID)
	require.NoError(t, err)
	require.Len(t, available, 1)
	assert.Equal(t, chore.ID, available[0].ChoreID)
	assert.Equal(t, 300, available[0].RewardCents)
	assert.NotNil(t, available[0].PeriodStart)
	assert.NotNil(t, available[0].PeriodEnd)
}

func TestGenerateInstances_SkipsExistingPeriod(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Bob")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Daily Dishes",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	_, err = choreRepo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: chore.ID,
		ChildID: child.ID,
	})
	require.NoError(t, err)

	scheduler := NewScheduler(choreRepo, instanceRepo, repositories.NewChildRepo(db), repositories.NewFamilyRepo(db))

	// Run twice
	scheduler.GenerateInstances()
	scheduler.GenerateInstances()

	// Should still only have 1 instance
	available, _, _, err := instanceRepo.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, available, 1)
}

func TestGenerateInstances_SkipsInactiveChores(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Charlie")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Inactive Chore",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	// Deactivate the chore after creation
	err = choreRepo.SetActive(chore.ID, false)
	require.NoError(t, err)

	_, err = choreRepo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: chore.ID,
		ChildID: child.ID,
	})
	require.NoError(t, err)

	scheduler := NewScheduler(choreRepo, instanceRepo, repositories.NewChildRepo(db), repositories.NewFamilyRepo(db))
	scheduler.GenerateInstances()

	available, _, _, err := instanceRepo.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, available, 0)
}

func TestGenerateInstances_SkipsDisabledChildren(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Dana")

	// Disable child
	db.Model(&models.Child{}).Where("id = ?", child.ID).Update("is_disabled", true)

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Daily Task",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	_, err = choreRepo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: chore.ID,
		ChildID: child.ID,
	})
	require.NoError(t, err)

	scheduler := NewScheduler(choreRepo, instanceRepo, repositories.NewChildRepo(db), repositories.NewFamilyRepo(db))
	scheduler.GenerateInstances()

	available, _, _, err := instanceRepo.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, available, 0)
}

func TestExpireInstances(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Eve")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Expired Chore",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	// Create instance with period_end in the past
	pastStart := time.Now().AddDate(0, 0, -2)
	pastEnd := time.Now().AddDate(0, 0, -1)
	_, err = instanceRepo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 100,
		Status:      models.ChoreInstanceStatusAvailable,
		PeriodStart: &pastStart,
		PeriodEnd:   &pastEnd,
	})
	require.NoError(t, err)

	scheduler := NewScheduler(choreRepo, instanceRepo, repositories.NewChildRepo(db), repositories.NewFamilyRepo(db))
	scheduler.ExpireInstances()

	// Verify instance is now expired
	instance, err := instanceRepo.GetByID(1)
	require.NoError(t, err)
	require.NotNil(t, instance)
	assert.Equal(t, models.ChoreInstanceStatusExpired, instance.Status)
}
