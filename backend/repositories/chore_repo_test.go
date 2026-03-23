package repositories

import (
	"testing"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createChoreTestFamily(t *testing.T, db *gorm.DB) *models.Family {
	t.Helper()
	fr := NewFamilyRepo(db)
	f, err := fr.Create("chore-test-family")
	require.NoError(t, err)
	return f
}

func createChoreTestParent(t *testing.T, db *gorm.DB, familyID int64) *models.Parent {
	t.Helper()
	pr := NewParentRepo(db)
	p, err := pr.Create("google-chore-test-123", "choretest@example.com", "Chore Parent")
	require.NoError(t, err)
	err = pr.SetFamilyID(p.ID, familyID)
	require.NoError(t, err)
	p.FamilyID = familyID
	return p
}

func createChoreTestChild(t *testing.T, db *gorm.DB, familyID int64, name string) *models.Child {
	t.Helper()
	cr := NewChildRepo(db)
	c, err := cr.Create(familyID, name, "password123", nil)
	require.NoError(t, err)
	return c
}

func TestChoreRepo_Create(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)

	desc := "Take out the trash"
	chore := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Take Out Trash",
		Description:       &desc,
		RewardCents:       250,
		Recurrence:        models.ChoreRecurrenceWeekly,
		DayOfWeek:         intP(3),
		IsActive:          true,
	}

	created, err := repo.Create(chore)
	require.NoError(t, err)
	assert.True(t, created.ID > 0)
	assert.Equal(t, fam.ID, created.FamilyID)
	assert.Equal(t, parent.ID, created.CreatedByParentID)
	assert.Equal(t, "Take Out Trash", created.Name)
	assert.Equal(t, "Take out the trash", *created.Description)
	assert.Equal(t, 250, created.RewardCents)
	assert.Equal(t, models.ChoreRecurrenceWeekly, created.Recurrence)
	assert.Equal(t, 3, *created.DayOfWeek)
	assert.Nil(t, created.DayOfMonth)
	assert.True(t, created.IsActive)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())
}

func TestChoreRepo_GetByID(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)

	chore := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Wash Dishes",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	}
	created, err := repo.Create(chore)
	require.NoError(t, err)

	// Fetch by ID
	fetched, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Wash Dishes", fetched.Name)
	assert.Equal(t, 100, fetched.RewardCents)
	assert.Equal(t, models.ChoreRecurrenceDaily, fetched.Recurrence)

	// Not found returns nil
	notFound, err := repo.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestChoreRepo_ListByFamily(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child1 := createChoreTestChild(t, db, fam.ID, "Alice")
	child2 := createChoreTestChild(t, db, fam.ID, "Bob")

	// Create two chores
	chore1 := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Vacuum",
		RewardCents:       300,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	}
	created1, err := repo.Create(chore1)
	require.NoError(t, err)

	chore2 := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Mow Lawn",
		RewardCents:       500,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	}
	created2, err := repo.Create(chore2)
	require.NoError(t, err)

	// Assign child1 to chore1, child2 to chore2
	_, err = repo.CreateAssignment(&models.ChoreAssignment{ChoreID: created1.ID, ChildID: child1.ID})
	require.NoError(t, err)
	_, err = repo.CreateAssignment(&models.ChoreAssignment{ChoreID: created2.ID, ChildID: child2.ID})
	require.NoError(t, err)

	// Create a pending_approval instance for chore1
	instance := &models.ChoreInstance{
		ChoreID:     created1.ID,
		ChildID:     child1.ID,
		RewardCents: 300,
		Status:      models.ChoreInstanceStatusPendingApproval,
	}
	err = db.Create(instance).Error
	require.NoError(t, err)

	// List
	results, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Results are ordered by created_at DESC, so chore2 is first
	assert.Equal(t, created2.ID, results[0].ID)
	assert.Equal(t, "Mow Lawn", results[0].Name)
	assert.Len(t, results[0].Assignments, 1)
	assert.Equal(t, child2.ID, results[0].Assignments[0].ChildID)
	assert.Equal(t, "Bob", results[0].Assignments[0].ChildName)
	assert.Equal(t, 0, results[0].PendingCount)

	assert.Equal(t, created1.ID, results[1].ID)
	assert.Equal(t, "Vacuum", results[1].Name)
	assert.Len(t, results[1].Assignments, 1)
	assert.Equal(t, child1.ID, results[1].Assignments[0].ChildID)
	assert.Equal(t, "Alice", results[1].Assignments[0].ChildName)
	assert.Equal(t, 1, results[1].PendingCount)
}

func TestChoreRepo_CreateAssignment(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "Charlie")

	chore := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Feed Dog",
		RewardCents:       150,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	}
	created, err := repo.Create(chore)
	require.NoError(t, err)

	// Create assignment
	assignment, err := repo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: created.ID,
		ChildID: child.ID,
	})
	require.NoError(t, err)
	assert.True(t, assignment.ID > 0)
	assert.Equal(t, created.ID, assignment.ChoreID)
	assert.Equal(t, child.ID, assignment.ChildID)
	assert.False(t, assignment.CreatedAt.IsZero())

	// Duplicate should return error (unique constraint)
	_, err = repo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: created.ID,
		ChildID: child.ID,
	})
	assert.Error(t, err)
}

func TestChoreRepo_DeleteAssignment(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "Dana")

	chore := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Clean Room",
		RewardCents:       200,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	}
	created, err := repo.Create(chore)
	require.NoError(t, err)

	_, err = repo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: created.ID,
		ChildID: child.ID,
	})
	require.NoError(t, err)

	// Delete assignment
	err = repo.DeleteAssignment(created.ID, child.ID)
	require.NoError(t, err)

	// Verify assignment is gone by listing
	results, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Len(t, results[0].Assignments, 0)
}

func TestChoreRepo_Delete(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "Eve")

	chore := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Rake Leaves",
		RewardCents:       400,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	}
	created, err := repo.Create(chore)
	require.NoError(t, err)

	// Add assignment
	_, err = repo.CreateAssignment(&models.ChoreAssignment{
		ChoreID: created.ID,
		ChildID: child.ID,
	})
	require.NoError(t, err)

	// Delete chore
	err = repo.Delete(created.ID)
	require.NoError(t, err)

	// Verify chore is gone
	fetched, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)

	// Verify assignment was cascaded (query directly)
	var count int64
	db.Model(&models.ChoreAssignment{}).Where("chore_id = ?", created.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Assignment should be deleted when chore is deleted (CASCADE)")
}

func TestChoreRepo_Update(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)

	chore := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Sweep Floor",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	}
	created, err := repo.Create(chore)
	require.NoError(t, err)

	// Update name and reward_cents
	created.Name = "Mop Floor"
	created.RewardCents = 200
	updated, err := repo.Update(created)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Mop Floor", updated.Name)
	assert.Equal(t, 200, updated.RewardCents)

	// Verify persisted
	fetched, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Mop Floor", fetched.Name)
	assert.Equal(t, 200, fetched.RewardCents)
}

func TestChoreRepo_SetActive(t *testing.T) {
	db := testDB(t)
	repo := NewChoreRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)

	chore := &models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Water Plants",
		RewardCents:       50,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	}
	created, err := repo.Create(chore)
	require.NoError(t, err)
	assert.True(t, created.IsActive)

	// Set inactive
	err = repo.SetActive(created.ID, false)
	require.NoError(t, err)

	fetched, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.False(t, fetched.IsActive)

	// Set active again
	err = repo.SetActive(created.ID, true)
	require.NoError(t, err)

	fetched, err = repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.True(t, fetched.IsActive)
}
