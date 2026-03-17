package repositories

import (
	"testing"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- helpers ---

func strPtr(s string) *string  { return &s }
func int64Ptr(i int64) *int64  { return &i }

func createTestFamilyWithParentAndChild(t *testing.T, db *gorm.DB) (*models.Family, *models.Parent, *models.Child) {
	t.Helper()
	fam := models.Family{Slug: "test-family"}
	require.NoError(t, db.Create(&fam).Error)

	parent := models.Parent{GoogleID: "google-" + t.Name(), Email: t.Name() + "@test.com", DisplayName: "Test Parent", FamilyID: fam.ID}
	require.NoError(t, db.Create(&parent).Error)

	child := models.Child{FamilyID: fam.ID, FirstName: "Saver", PasswordHash: "hash123"}
	require.NoError(t, db.Create(&child).Error)

	return &fam, &parent, &child
}

func depositForChild(t *testing.T, db *gorm.DB, childID, parentID, amountCents int64) {
	t.Helper()
	tx := models.Transaction{
		ChildID:         childID,
		ParentID:        parentID,
		AmountCents:     amountCents,
		TransactionType: models.TransactionTypeDeposit,
	}
	require.NoError(t, db.Create(&tx).Error)
	require.NoError(t, db.Model(&models.Child{}).Where("id = ?", childID).
		Update("balance_cents", gorm.Expr("balance_cents + ?", amountCents)).Error)
}

// --- TestSavingsGoalRepo_Create ---

func TestSavingsGoalRepo_Create_AllFields(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	emoji := "🎮"
	goal, err := repo.Create(child.ID, "New Bike", 50000, &emoji)
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.NotZero(t, goal.ID)
	assert.Equal(t, child.ID, goal.ChildID)
	assert.Equal(t, "New Bike", goal.Name)
	assert.Equal(t, int64(50000), goal.TargetCents)
	assert.Equal(t, int64(0), goal.SavedCents)
	require.NotNil(t, goal.Emoji)
	assert.Equal(t, "🎮", *goal.Emoji)
	assert.Equal(t, "active", goal.Status)
	assert.Nil(t, goal.CompletedAt)
	assert.False(t, goal.CreatedAt.IsZero())
	assert.False(t, goal.UpdatedAt.IsZero())
}

func TestSavingsGoalRepo_Create_RequiredFieldsOnly(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	goal, err := repo.Create(child.ID, "Piggy Bank", 1000, nil)
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.NotZero(t, goal.ID)
	assert.Equal(t, child.ID, goal.ChildID)
	assert.Equal(t, "Piggy Bank", goal.Name)
	assert.Equal(t, int64(1000), goal.TargetCents)
	assert.Equal(t, int64(0), goal.SavedCents)
	assert.Nil(t, goal.Emoji)
	assert.Equal(t, "active", goal.Status)
	assert.Nil(t, goal.CompletedAt)
	assert.False(t, goal.CreatedAt.IsZero())
	assert.False(t, goal.UpdatedAt.IsZero())
}

// --- TestSavingsGoalRepo_GetByID ---

func TestSavingsGoalRepo_GetByID_Found(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	emoji := "🚀"
	created, err := repo.Create(child.ID, "Rocket Ship", 99999, &emoji)
	require.NoError(t, err)

	found, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.ChildID, found.ChildID)
	assert.Equal(t, "Rocket Ship", found.Name)
	assert.Equal(t, int64(99999), found.TargetCents)
	assert.Equal(t, int64(0), found.SavedCents)
	require.NotNil(t, found.Emoji)
	assert.Equal(t, "🚀", *found.Emoji)
	assert.Equal(t, "active", found.Status)
	assert.Nil(t, found.CompletedAt)
	assert.False(t, found.CreatedAt.IsZero())
	assert.False(t, found.UpdatedAt.IsZero())
}

func TestSavingsGoalRepo_GetByID_NotFound(t *testing.T) {
	db := testDB(t)

	repo := NewSavingsGoalRepo(db)

	found, err := repo.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, found)
}

// --- TestSavingsGoalRepo_ListByChild ---

func TestSavingsGoalRepo_ListByChild_Ordering(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	goal1, err := repo.Create(child.ID, "Goal A", 1000, nil)
	require.NoError(t, err)
	goal2, err := repo.Create(child.ID, "Goal B", 2000, nil)
	require.NoError(t, err)
	goal3, err := repo.Create(child.ID, "Goal C", 3000, nil)
	require.NoError(t, err)

	// Mark goal2 as completed via direct SQL
	require.NoError(t, db.Exec(`UPDATE savings_goals SET status = 'completed', completed_at = NOW() WHERE id = ?`, goal2.ID).Error)

	goals, err := repo.ListByChild(child.ID)
	require.NoError(t, err)
	require.Len(t, goals, 3)

	// Active goals first, ordered by created_at ASC
	assert.Equal(t, goal1.ID, goals[0].ID)
	assert.Equal(t, "active", goals[0].Status)
	assert.Equal(t, goal3.ID, goals[1].ID)
	assert.Equal(t, "active", goals[1].Status)
	// Completed goal last
	assert.Equal(t, goal2.ID, goals[2].ID)
	assert.Equal(t, "completed", goals[2].Status)
}

func TestSavingsGoalRepo_ListByChild_Empty(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	goals, err := repo.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, goals, 0)
}

// --- TestSavingsGoalRepo_CountActiveByChild ---

func TestSavingsGoalRepo_CountActiveByChild(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	_, err := repo.Create(child.ID, "Goal A", 1000, nil)
	require.NoError(t, err)
	goal2, err := repo.Create(child.ID, "Goal B", 2000, nil)
	require.NoError(t, err)
	_, err = repo.Create(child.ID, "Goal C", 3000, nil)
	require.NoError(t, err)

	// Mark one as completed
	require.NoError(t, db.Exec(`UPDATE savings_goals SET status = 'completed', completed_at = NOW() WHERE id = ?`, goal2.ID).Error)

	count, err := repo.CountActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSavingsGoalRepo_CountActiveByChild_NoGoals(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	count, err := repo.CountActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// --- TestSavingsGoalRepo_Allocate ---

func TestSavingsGoalRepo_Allocate_Success(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "New Bike", 50000, nil)
	require.NoError(t, err)

	updated, err := repo.Allocate(goal.ID, child.ID, 2000)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int64(2000), updated.SavedCents)

	fetched, err := repo.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), fetched.SavedCents)
}

func TestSavingsGoalRepo_Allocate_ExceedsAvailableBalance(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 5000)

	repo := NewSavingsGoalRepo(db)

	goal1, err := repo.Create(child.ID, "Goal A", 10000, nil)
	require.NoError(t, err)
	_, err = repo.Allocate(goal1.ID, child.ID, 3000)
	require.NoError(t, err)

	goal2, err := repo.Create(child.ID, "Goal B", 10000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal2.ID, child.ID, 2500)
	require.Error(t, err)
}

func TestSavingsGoalRepo_Allocate_ZeroAmount(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Goal", 5000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 0)
	require.Error(t, err)
}

func TestSavingsGoalRepo_Allocate_GoalNotFound(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	_, err := repo.Allocate(99999, child.ID, 1000)
	require.Error(t, err)
}

func TestSavingsGoalRepo_Allocate_Deallocate_Success(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Bike Fund", 50000, nil)
	require.NoError(t, err)

	// Allocate $20
	updated, err := repo.Allocate(goal.ID, child.ID, 2000)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), updated.SavedCents)

	// De-allocate $10
	updated, err = repo.Allocate(goal.ID, child.ID, -1000)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), updated.SavedCents)

	fetched, err := repo.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), fetched.SavedCents)
}

func TestSavingsGoalRepo_Allocate_Deallocate_ExceedsSavedCents(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Bike Fund", 50000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	// Try to de-allocate $15 when only $10 saved
	_, err = repo.Allocate(goal.ID, child.ID, -1500)
	require.Error(t, err)
}

// --- Allocate auto-completion tests ---

func TestSavingsGoalRepo_Allocate_CompletesGoal(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Small Goal", 5000, nil)
	require.NoError(t, err)

	updated, err := repo.Allocate(goal.ID, child.ID, 5000)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, int64(5000), updated.SavedCents)
	assert.Equal(t, "completed", updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestSavingsGoalRepo_Allocate_OverAllocationCompletesGoal(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Small Goal", 3000, nil)
	require.NoError(t, err)

	updated, err := repo.Allocate(goal.ID, child.ID, 5000)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, int64(5000), updated.SavedCents)
	assert.Equal(t, "completed", updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestSavingsGoalRepo_Allocate_ToCompletedGoal_Fails(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Done Goal", 1000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 500)
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

// --- TestSavingsGoalRepo_GetAvailableBalance ---

func TestSavingsGoalRepo_GetAvailableBalance_WithActiveGoals(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)

	goal1, err := repo.Create(child.ID, "Goal A", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ?, updated_at = NOW() WHERE id = ?`, 2000, goal1.ID).Error)

	goal2, err := repo.Create(child.ID, "Goal B", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ?, updated_at = NOW() WHERE id = ?`, 3000, goal2.ID).Error)

	available, err := repo.GetAvailableBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5000), available)
}

func TestSavingsGoalRepo_GetAvailableBalance_NoGoals(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)

	available, err := repo.GetAvailableBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000), available)
}

// --- TestSavingsGoalRepo_Update ---

func TestSavingsGoalRepo_Update_NameOnly(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	emoji := "🎯"
	goal, err := repo.Create(child.ID, "Original Name", 10000, &emoji)
	require.NoError(t, err)

	updated, err := repo.Update(goal.ID, child.ID, &UpdateGoalParams{
		Name: strPtr("New Name"),
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, int64(10000), updated.TargetCents)
	require.NotNil(t, updated.Emoji)
	assert.Equal(t, "🎯", *updated.Emoji)
	assert.Equal(t, "active", updated.Status)
	assert.Nil(t, updated.CompletedAt)
}

func TestSavingsGoalRepo_Update_AllFields(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	goal, err := repo.Create(child.ID, "Old Goal", 5000, nil)
	require.NoError(t, err)

	updated, err := repo.Update(goal.ID, child.ID, &UpdateGoalParams{
		Name:        strPtr("Updated Goal"),
		TargetCents: int64Ptr(20000),
		Emoji:       strPtr("🚀"),
		EmojiSet:    true,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, "Updated Goal", updated.Name)
	assert.Equal(t, int64(20000), updated.TargetCents)
	require.NotNil(t, updated.Emoji)
	assert.Equal(t, "🚀", *updated.Emoji)
	assert.Equal(t, "active", updated.Status)
}

func TestSavingsGoalRepo_Update_AutoCompleteOnTargetReduction(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)

	goal, err := repo.Create(child.ID, "Big Goal", 10000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 5000)
	require.NoError(t, err)

	// Update target down to 3000; saved (5000) >= new target (3000) => auto-complete
	updated, err := repo.Update(goal.ID, child.ID, &UpdateGoalParams{
		TargetCents: int64Ptr(3000),
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, int64(3000), updated.TargetCents)
	assert.Equal(t, int64(5000), updated.SavedCents)
	assert.Equal(t, "completed", updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestSavingsGoalRepo_Update_NotFound(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	_, err := repo.Update(99999, child.ID, &UpdateGoalParams{
		Name: strPtr("Ghost Goal"),
	})
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

func TestSavingsGoalRepo_Update_CompletedGoal(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Done Goal", 1000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	_, err = repo.Update(goal.ID, child.ID, &UpdateGoalParams{
		Name: strPtr("Renamed Done Goal"),
	})
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

// --- TestSavingsGoalRepo_Delete ---

func TestSavingsGoalRepo_Delete_Success(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Doomed Goal", 50000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 3000)
	require.NoError(t, err)

	released, err := repo.Delete(goal.ID, child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3000), released)

	fetched, err := repo.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestSavingsGoalRepo_Delete_NotFound(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	_, err := repo.Delete(99999, child.ID)
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

func TestSavingsGoalRepo_Delete_CompletedGoal(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)
	goal, err := repo.Create(child.ID, "Completed Goal", 1000, nil)
	require.NoError(t, err)

	_, err = repo.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	completed, err := repo.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)
	assert.Equal(t, int64(1000), completed.SavedCents)

	released, err := repo.Delete(goal.ID, child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), released)

	fetched, err := repo.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

// --- TestSavingsGoalRepo_GetTotalSavedByChild ---

func TestSavingsGoalRepo_GetTotalSavedByChild(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)

	goal1, err := repo.Create(child.ID, "Goal A", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ? WHERE id = ?`, 2000, goal1.ID).Error)

	goal2, err := repo.Create(child.ID, "Goal B", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ? WHERE id = ?`, 3000, goal2.ID).Error)

	total, err := repo.GetTotalSavedByChild(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5000), total)
}

func TestSavingsGoalRepo_GetTotalSavedByChild_NoGoals(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	total, err := repo.GetTotalSavedByChild(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

// --- TestSavingsGoalRepo_ReduceGoalsProportionally ---

func TestSavingsGoalRepo_ReduceGoalsProportionally(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)

	goal1, err := repo.Create(child.ID, "Goal A", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ? WHERE id = ?`, 4000, goal1.ID).Error)

	goal2, err := repo.Create(child.ID, "Goal B", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ? WHERE id = ?`, 6000, goal2.ID).Error)

	// Release 5000 from total 10000 saved proportionally
	err = repo.ReduceGoalsProportionally(child.ID, 5000)
	require.NoError(t, err)

	g1, err := repo.GetByID(goal1.ID)
	require.NoError(t, err)
	g2, err := repo.GetByID(goal2.ID)
	require.NoError(t, err)

	// goal1: 4000 * 5000 / 10000 = 2000 reduction => 2000 remaining
	assert.Equal(t, int64(2000), g1.SavedCents)
	// goal2: gets remainder 5000 - 2000 = 3000 reduction => 3000 remaining
	assert.Equal(t, int64(3000), g2.SavedCents)
}

func TestSavingsGoalRepo_ReduceGoalsProportionally_ZeroRelease(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	repo := NewSavingsGoalRepo(db)

	err := repo.ReduceGoalsProportionally(child.ID, 0)
	require.NoError(t, err)
}

// --- TestSavingsGoalRepo_GetAffectedGoals ---

func TestSavingsGoalRepo_GetAffectedGoals(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	repo := NewSavingsGoalRepo(db)

	goal1, err := repo.Create(child.ID, "Goal A", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ? WHERE id = ?`, 4000, goal1.ID).Error)

	goal2, err := repo.Create(child.ID, "Goal B", 10000, nil)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE savings_goals SET saved_cents = ? WHERE id = ?`, 6000, goal2.ID).Error)

	affected, err := repo.GetAffectedGoals(child.ID, 5000)
	require.NoError(t, err)
	require.Len(t, affected, 2)

	assert.Equal(t, goal1.ID, affected[0].ID)
	assert.Equal(t, int64(4000), affected[0].CurrentSavedCents)
	assert.Equal(t, int64(2000), affected[0].NewSavedCents)

	assert.Equal(t, goal2.ID, affected[1].ID)
	assert.Equal(t, int64(6000), affected[1].CurrentSavedCents)
	assert.Equal(t, int64(3000), affected[1].NewSavedCents)
}
