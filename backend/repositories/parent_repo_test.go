package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateParent(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)

	p, err := pr.Create("google-123", "jane@example.com", "Jane Smith")
	require.NoError(t, err)
	assert.NotZero(t, p.ID)
	assert.Equal(t, "google-123", p.GoogleID)
	assert.Equal(t, "jane@example.com", p.Email)
	assert.Equal(t, "Jane Smith", p.DisplayName)
	assert.Equal(t, int64(0), p.FamilyID)
}

func TestGetByGoogleID(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)

	created, err := pr.Create("google-456", "john@example.com", "John Doe")
	require.NoError(t, err)

	found, err := pr.GetByGoogleID("google-456")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "John Doe", found.DisplayName)
}

func TestGetByGoogleID_NotFound(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)

	found, err := pr.GetByGoogleID("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestCreateParent_DuplicateGoogleID(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)

	_, err := pr.Create("google-dup", "a@example.com", "User A")
	require.NoError(t, err)

	_, err = pr.Create("google-dup", "b@example.com", "User B")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestSetFamilyID(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)
	fr := NewFamilyRepo(db)

	p, err := pr.Create("google-789", "test@example.com", "Test User")
	require.NoError(t, err)
	assert.Equal(t, int64(0), p.FamilyID)

	fam, err := fr.Create("test-fam")
	require.NoError(t, err)

	err = pr.SetFamilyID(p.ID, fam.ID)
	require.NoError(t, err)

	updated, err := pr.GetByID(p.ID)
	require.NoError(t, err)
	assert.Equal(t, fam.ID, updated.FamilyID)
}

func TestGetByID_Parent(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)

	created, err := pr.Create("google-getbyid", "getbyid@example.com", "Get By ID")
	require.NoError(t, err)

	found, err := pr.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "google-getbyid", found.GoogleID)
}

func TestGetByID_Parent_NotFound(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)

	found, err := pr.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestGetByFamilyID(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("notify-family")
	require.NoError(t, err)

	p1, err := pr.Create("google-fam-1", "parent1@example.com", "Parent One")
	require.NoError(t, err)
	require.NoError(t, pr.SetFamilyID(p1.ID, fam.ID))

	p2, err := pr.Create("google-fam-2", "parent2@example.com", "Parent Two")
	require.NoError(t, err)
	require.NoError(t, pr.SetFamilyID(p2.ID, fam.ID))

	parents, err := pr.GetByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Len(t, parents, 2)

	// Family with no parents
	fam2, err := fr.Create("empty-family")
	require.NoError(t, err)
	empty, err := pr.GetByFamilyID(fam2.ID)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestUpdateNotificationPrefs(t *testing.T) {
	db := testDB(t)
	pr := NewParentRepo(db)

	p, err := pr.Create("google-prefs", "prefs@example.com", "Prefs User")
	require.NoError(t, err)

	// Defaults should be true
	found, err := pr.GetByID(p.ID)
	require.NoError(t, err)
	assert.True(t, found.NotifyWithdrawalRequests)
	assert.True(t, found.NotifyChoreCompletions)
	assert.True(t, found.NotifyDecisions)

	// Partial update — only change one field
	err = pr.UpdateNotificationPrefs(p.ID, map[string]bool{
		"notify_chore_completions": false,
	})
	require.NoError(t, err)

	found, err = pr.GetByID(p.ID)
	require.NoError(t, err)
	assert.True(t, found.NotifyWithdrawalRequests)
	assert.False(t, found.NotifyChoreCompletions)
	assert.True(t, found.NotifyDecisions)
}
