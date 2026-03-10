package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- TestSavingsGoalStore_Create ---

func TestSavingsGoalStore_Create_AllFields(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	emoji := "🎮"
	goal, err := gs.Create(child.ID, "New Bike", 50000, &emoji)
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

func TestSavingsGoalStore_Create_RequiredFieldsOnly(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	goal, err := gs.Create(child.ID, "Piggy Bank", 1000, nil)
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

// --- TestSavingsGoalStore_GetByID ---

func TestSavingsGoalStore_GetByID_Found(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	emoji := "🚀"
	created, err := gs.Create(child.ID, "Rocket Ship", 99999, &emoji)
	require.NoError(t, err)

	found, err := gs.GetByID(created.ID)
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

func TestSavingsGoalStore_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	gs := NewSavingsGoalStore(db)

	found, err := gs.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, found)
}

// --- TestSavingsGoalStore_ListByChild ---

func TestSavingsGoalStore_ListByChild_Ordering(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	// Create three goals; they'll get created_at in ascending order
	goal1, err := gs.Create(child.ID, "Goal A", 1000, nil)
	require.NoError(t, err)
	goal2, err := gs.Create(child.ID, "Goal B", 2000, nil)
	require.NoError(t, err)
	goal3, err := gs.Create(child.ID, "Goal C", 3000, nil)
	require.NoError(t, err)

	// Mark goal2 as completed via direct SQL
	_, err = db.Exec(`UPDATE savings_goals SET status = 'completed', completed_at = NOW() WHERE id = $1`, goal2.ID)
	require.NoError(t, err)

	goals, err := gs.ListByChild(child.ID)
	require.NoError(t, err)
	require.Len(t, goals, 3)

	// Active goals first, ordered by created_at DESC
	// goal3 (active, newest) should come second
	assert.Equal(t, goal3.ID, goals[1].ID)
	assert.Equal(t, "active", goals[0].Status)

	// goal1 (active, oldest) should come first
	assert.Equal(t, goal1.ID, goals[0].ID)
	assert.Equal(t, "active", goals[1].Status)

	// goal2 (completed) should come last
	assert.Equal(t, goal2.ID, goals[2].ID)
	assert.Equal(t, "completed", goals[2].Status)
}

func TestSavingsGoalStore_ListByChild_Empty(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	goals, err := gs.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, goals, 0)
}

// --- TestSavingsGoalStore_CountActiveByChild ---

func TestSavingsGoalStore_CountActiveByChild(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	// Create two active goals
	_, err = gs.Create(child.ID, "Goal A", 1000, nil)
	require.NoError(t, err)
	goal2, err := gs.Create(child.ID, "Goal B", 2000, nil)
	require.NoError(t, err)
	_, err = gs.Create(child.ID, "Goal C", 3000, nil)
	require.NoError(t, err)

	// Mark one as completed
	_, err = db.Exec(`UPDATE savings_goals SET status = 'completed', completed_at = NOW() WHERE id = $1`, goal2.ID)
	require.NoError(t, err)

	count, err := gs.CountActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSavingsGoalStore_CountActiveByChild_NoGoals(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	count, err := gs.CountActiveByChild(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// --- TestSavingsGoalStore_Allocate ---

func TestSavingsGoalStore_Allocate_Success(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	// Create parent and associate with family
	ps := NewParentStore(db)
	parent, err := ps.Create("google-alloc-1", "parent-alloc@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	// Create child and deposit $100
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	// Create a goal
	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "New Bike", 50000, nil)
	require.NoError(t, err)

	// Allocate $20 to the goal
	updated, err := gs.Allocate(goal.ID, child.ID, 2000)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, int64(2000), updated.SavedCents)

	// Verify the goal was updated in the DB
	fetched, err := gs.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), fetched.SavedCents)
}

func TestSavingsGoalStore_Allocate_ExceedsAvailableBalance(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-alloc-2", "parent-alloc2@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	// Deposit $50
	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 5000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	// Create a goal and allocate $30 to it
	goal1, err := gs.Create(child.ID, "Goal A", 10000, nil)
	require.NoError(t, err)
	_, err = gs.Allocate(goal1.ID, child.ID, 3000)
	require.NoError(t, err)

	// Create a second goal and try to allocate $25 — only $20 is available
	goal2, err := gs.Create(child.ID, "Goal B", 10000, nil)
	require.NoError(t, err)

	_, err = gs.Allocate(goal2.ID, child.ID, 2500)
	require.Error(t, err)
}

func TestSavingsGoalStore_Allocate_ZeroAmount(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Goal", 5000, nil)
	require.NoError(t, err)

	_, err = gs.Allocate(goal.ID, child.ID, 0)
	require.Error(t, err)
}

func TestSavingsGoalStore_Allocate_GoalNotFound(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	_, err = gs.Allocate(99999, child.ID, 1000)
	require.Error(t, err)
}

func TestSavingsGoalStore_Allocate_Deallocate_Success(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-alloc-3", "parent-alloc3@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Bike Fund", 50000, nil)
	require.NoError(t, err)

	// Allocate $20
	updated, err := gs.Allocate(goal.ID, child.ID, 2000)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), updated.SavedCents)

	// De-allocate $10 (negative amount)
	updated, err = gs.Allocate(goal.ID, child.ID, -1000)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), updated.SavedCents)

	// Verify in DB
	fetched, err := gs.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), fetched.SavedCents)
}

func TestSavingsGoalStore_Allocate_Deallocate_ExceedsSavedCents(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-alloc-4", "parent-alloc4@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Bike Fund", 50000, nil)
	require.NoError(t, err)

	// Allocate $10
	_, err = gs.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	// Try to de-allocate $15 — goal only has $10
	_, err = gs.Allocate(goal.ID, child.ID, -1500)
	require.Error(t, err)
}

// --- TestSavingsGoalStore_GetAvailableBalance ---

func TestSavingsGoalStore_GetAvailableBalance_WithActiveGoals(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-avail-1", "parent-avail@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	// Deposit $100
	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	// Create two active goals with saved amounts via direct SQL
	goal1, err := gs.Create(child.ID, "Goal A", 10000, nil)
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE savings_goals SET saved_cents = $1, updated_at = NOW() WHERE id = $2`, 2000, goal1.ID)
	require.NoError(t, err)

	goal2, err := gs.Create(child.ID, "Goal B", 10000, nil)
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE savings_goals SET saved_cents = $1, updated_at = NOW() WHERE id = $2`, 3000, goal2.ID)
	require.NoError(t, err)

	// Available should be $100 - $20 - $30 = $50
	available, err := gs.GetAvailableBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5000), available)
}

func TestSavingsGoalStore_GetAvailableBalance_NoGoals(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-avail-2", "parent-avail2@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	// Deposit $100
	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	// Available should be the full $100
	available, err := gs.GetAvailableBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000), available)
}

// --- TestSavingsGoalStore_ListAllocationsByGoal ---

func TestSavingsGoalStore_ListAllocationsByGoal_Order(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-list-alloc-1", "parent-listalloc@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Bike Fund", 50000, nil)
	require.NoError(t, err)

	// Make multiple allocations
	_, err = gs.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)
	_, err = gs.Allocate(goal.ID, child.ID, 2000)
	require.NoError(t, err)
	_, err = gs.Allocate(goal.ID, child.ID, -500)
	require.NoError(t, err)

	allocs, err := gs.ListAllocationsByGoal(goal.ID)
	require.NoError(t, err)
	require.Len(t, allocs, 3)

	// Reverse chronological order: newest first
	assert.Equal(t, int64(-500), allocs[0].AmountCents)
	assert.Equal(t, int64(2000), allocs[1].AmountCents)
	assert.Equal(t, int64(1000), allocs[2].AmountCents)

	// Verify all allocations have correct goal and child IDs
	for _, a := range allocs {
		assert.Equal(t, goal.ID, a.GoalID)
		assert.Equal(t, child.ID, a.ChildID)
		assert.False(t, a.CreatedAt.IsZero())
		assert.NotZero(t, a.ID)
	}
}

func TestSavingsGoalStore_ListAllocationsByGoal_Empty(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Empty Goal", 5000, nil)
	require.NoError(t, err)

	allocs, err := gs.ListAllocationsByGoal(goal.ID)
	require.NoError(t, err)
	assert.Len(t, allocs, 0)
}

// --- T029: Allocate auto-completion tests ---

func TestSavingsGoalStore_Allocate_CompletesGoal(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-complete-1", "parent-complete1@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Small Goal", 5000, nil)
	require.NoError(t, err)

	// Allocate exactly the target amount — should auto-complete
	updated, err := gs.Allocate(goal.ID, child.ID, 5000)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, int64(5000), updated.SavedCents)
	assert.Equal(t, "completed", updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestSavingsGoalStore_Allocate_OverAllocationCompletesGoal(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-complete-2", "parent-complete2@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Small Goal", 3000, nil)
	require.NoError(t, err)

	// Allocate more than target — should still complete
	updated, err := gs.Allocate(goal.ID, child.ID, 5000)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, int64(5000), updated.SavedCents)
	assert.Equal(t, "completed", updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestSavingsGoalStore_Allocate_ToCompletedGoal_Fails(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-complete-3", "parent-complete3@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Done Goal", 1000, nil)
	require.NoError(t, err)

	// Complete the goal
	_, err = gs.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	// Try to allocate to completed goal — should fail
	_, err = gs.Allocate(goal.ID, child.ID, 500)
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

// --- Helpers for pointer construction ---

func strPtr(s string) *string { return &s }
func int64Ptr(i int64) *int64 { return &i }

// --- TestSavingsGoalStore_Update ---

func TestSavingsGoalStore_Update_NameOnly(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	emoji := "🎯"
	goal, err := gs.Create(child.ID, "Original Name", 10000, &emoji)
	require.NoError(t, err)

	// Update only the name
	updated, err := gs.Update(goal.ID, child.ID, &UpdateGoalParams{
		Name: strPtr("New Name"),
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, "New Name", updated.Name)
	// Other fields should remain unchanged
	assert.Equal(t, int64(10000), updated.TargetCents)
	require.NotNil(t, updated.Emoji)
	assert.Equal(t, "🎯", *updated.Emoji)
	assert.Equal(t, "active", updated.Status)
	assert.Nil(t, updated.CompletedAt)
}

func TestSavingsGoalStore_Update_AllFields(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	goal, err := gs.Create(child.ID, "Old Goal", 5000, nil)
	require.NoError(t, err)

	updated, err := gs.Update(goal.ID, child.ID, &UpdateGoalParams{
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

func TestSavingsGoalStore_Update_AutoCompleteOnTargetReduction(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-update-1", "parent-update1@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	// Deposit $100 so child has balance for allocation
	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	// Create a goal with target $100 (10000 cents)
	goal, err := gs.Create(child.ID, "Big Goal", 10000, nil)
	require.NoError(t, err)

	// Allocate $50 (5000 cents)
	_, err = gs.Allocate(goal.ID, child.ID, 5000)
	require.NoError(t, err)

	// Update target down to $30 (3000 cents). Since saved (5000) >= new target (3000),
	// the goal should auto-complete.
	updated, err := gs.Update(goal.ID, child.ID, &UpdateGoalParams{
		TargetCents: int64Ptr(3000),
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, int64(3000), updated.TargetCents)
	assert.Equal(t, int64(5000), updated.SavedCents)
	assert.Equal(t, "completed", updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestSavingsGoalStore_Update_NotFound(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	_, err = gs.Update(99999, child.ID, &UpdateGoalParams{
		Name: strPtr("Ghost Goal"),
	})
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

func TestSavingsGoalStore_Update_CompletedGoal(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-update-2", "parent-update2@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Done Goal", 1000, nil)
	require.NoError(t, err)

	// Complete the goal by allocating the full target
	_, err = gs.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	// Try to update the completed goal — should fail
	_, err = gs.Update(goal.ID, child.ID, &UpdateGoalParams{
		Name: strPtr("Renamed Done Goal"),
	})
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

// --- TestSavingsGoalStore_Delete ---

func TestSavingsGoalStore_Delete_Success(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-delete-1", "parent-delete1@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Doomed Goal", 50000, nil)
	require.NoError(t, err)

	// Allocate $30 to the goal
	_, err = gs.Allocate(goal.ID, child.ID, 3000)
	require.NoError(t, err)

	// Delete the goal — should return the released saved_cents (3000)
	released, err := gs.Delete(goal.ID, child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3000), released)

	// Verify the goal no longer exists (or is not retrievable as active)
	fetched, err := gs.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestSavingsGoalStore_Delete_NotFound(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)
	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)

	_, err = gs.Delete(99999, child.ID)
	require.Error(t, err)
	assert.Equal(t, ErrGoalNotFound, err)
}

func TestSavingsGoalStore_Delete_CompletedGoal(t *testing.T) {
	db := testDB(t)
	fam := createTestFamily(t, db)

	ps := NewParentStore(db)
	parent, err := ps.Create("google-delete-2", "parent-delete2@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	cs := NewChildStore(db)
	child, err := cs.Create(fam.ID, "Saver", "password123", nil)
	require.NoError(t, err)

	txStore := NewTransactionStore(db)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	gs := NewSavingsGoalStore(db)
	goal, err := gs.Create(child.ID, "Completed Goal", 1000, nil)
	require.NoError(t, err)

	// Complete the goal
	_, err = gs.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)

	// Verify the goal is completed with saved_cents still set
	completed, err := gs.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)
	assert.Equal(t, int64(1000), completed.SavedCents)

	// Delete the completed goal — should succeed
	released, err := gs.Delete(goal.ID, child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), released)

	// Verify the goal no longer exists
	fetched, err := gs.GetByID(goal.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}
