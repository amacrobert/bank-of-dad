package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- TestGoalAllocationRepo_ListByGoal ---

func TestGoalAllocationRepo_ListByGoal_Order(t *testing.T) {
	db := testDB(t)
	_, parent, child := createTestFamilyWithParentAndChild(t, db)
	depositForChild(t, db, child.ID, parent.ID, 10000)

	goalRepo := NewSavingsGoalRepo(db)
	goal, err := goalRepo.Create(child.ID, "Bike Fund", 50000, nil)
	require.NoError(t, err)

	// Make multiple allocations
	_, err = goalRepo.Allocate(goal.ID, child.ID, 1000)
	require.NoError(t, err)
	_, err = goalRepo.Allocate(goal.ID, child.ID, 2000)
	require.NoError(t, err)
	_, err = goalRepo.Allocate(goal.ID, child.ID, -500)
	require.NoError(t, err)

	allocRepo := NewGoalAllocationRepo(db)
	allocs, err := allocRepo.ListByGoal(goal.ID)
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

func TestGoalAllocationRepo_ListByGoal_Empty(t *testing.T) {
	db := testDB(t)
	_, _, child := createTestFamilyWithParentAndChild(t, db)

	goalRepo := NewSavingsGoalRepo(db)
	goal, err := goalRepo.Create(child.ID, "Empty Goal", 5000, nil)
	require.NoError(t, err)

	allocRepo := NewGoalAllocationRepo(db)
	allocs, err := allocRepo.ListByGoal(goal.ID)
	require.NoError(t, err)
	assert.Len(t, allocs, 0)
}
