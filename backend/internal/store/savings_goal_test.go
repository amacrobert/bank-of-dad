package store

import (
	"testing"
	"time"

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
	targetDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	goal, err := gs.Create(child.ID, "New Bike", 50000, &emoji, &targetDate)
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.NotZero(t, goal.ID)
	assert.Equal(t, child.ID, goal.ChildID)
	assert.Equal(t, "New Bike", goal.Name)
	assert.Equal(t, int64(50000), goal.TargetCents)
	assert.Equal(t, int64(0), goal.SavedCents)
	require.NotNil(t, goal.Emoji)
	assert.Equal(t, "🎮", *goal.Emoji)
	require.NotNil(t, goal.TargetDate)
	assert.Equal(t, 2026, goal.TargetDate.Year())
	assert.Equal(t, time.Month(6), goal.TargetDate.Month())
	assert.Equal(t, 15, goal.TargetDate.Day())
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

	goal, err := gs.Create(child.ID, "Piggy Bank", 1000, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.NotZero(t, goal.ID)
	assert.Equal(t, child.ID, goal.ChildID)
	assert.Equal(t, "Piggy Bank", goal.Name)
	assert.Equal(t, int64(1000), goal.TargetCents)
	assert.Equal(t, int64(0), goal.SavedCents)
	assert.Nil(t, goal.Emoji)
	assert.Nil(t, goal.TargetDate)
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
	targetDate := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	created, err := gs.Create(child.ID, "Rocket Ship", 99999, &emoji, &targetDate)
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
	require.NotNil(t, found.TargetDate)
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
	goal1, err := gs.Create(child.ID, "Goal A", 1000, nil, nil)
	require.NoError(t, err)
	goal2, err := gs.Create(child.ID, "Goal B", 2000, nil, nil)
	require.NoError(t, err)
	goal3, err := gs.Create(child.ID, "Goal C", 3000, nil, nil)
	require.NoError(t, err)

	// Mark goal2 as completed via direct SQL
	_, err = db.Exec(`UPDATE savings_goals SET status = 'completed', completed_at = NOW() WHERE id = $1`, goal2.ID)
	require.NoError(t, err)

	goals, err := gs.ListByChild(child.ID)
	require.NoError(t, err)
	require.Len(t, goals, 3)

	// Active goals first, ordered by created_at DESC
	// goal3 (active, newest) should come first
	assert.Equal(t, goal3.ID, goals[0].ID)
	assert.Equal(t, "active", goals[0].Status)

	// goal1 (active, oldest) should come second
	assert.Equal(t, goal1.ID, goals[1].ID)
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
	_, err = gs.Create(child.ID, "Goal A", 1000, nil, nil)
	require.NoError(t, err)
	goal2, err := gs.Create(child.ID, "Goal B", 2000, nil, nil)
	require.NoError(t, err)
	_, err = gs.Create(child.ID, "Goal C", 3000, nil, nil)
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
