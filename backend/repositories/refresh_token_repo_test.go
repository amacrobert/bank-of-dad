package repositories

import (
	"testing"
	"time"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func refreshTokenTestRepo(t *testing.T) *RefreshTokenRepo {
	t.Helper()
	db := testDB(t)
	return NewRefreshTokenRepo(db)
}

func seedFamilyAndParent(t *testing.T, repo *RefreshTokenRepo) (int64, int64) {
	t.Helper()
	var family models.Family
	family.Slug = "test-family"
	require.NoError(t, repo.db.Create(&family).Error)

	var parent models.Parent
	parent.FamilyID = family.ID
	parent.GoogleID = "google-123"
	parent.Email = "test@example.com"
	parent.DisplayName = "Test Parent"
	require.NoError(t, repo.db.Create(&parent).Error)

	return family.ID, parent.ID
}

func TestRefreshTokenRepo_Create(t *testing.T) {
	repo := refreshTokenTestRepo(t)
	familyID, _ := seedFamilyAndParent(t, repo)

	token1, err := repo.Create("parent", 1, familyID, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := repo.Create("parent", 1, familyID, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)
}

func TestRefreshTokenRepo_Validate(t *testing.T) {
	repo := refreshTokenTestRepo(t)
	familyID, _ := seedFamilyAndParent(t, repo)

	rawToken, err := repo.Create("child", 2, familyID, time.Hour)
	require.NoError(t, err)

	rt, err := repo.Validate(rawToken)
	require.NoError(t, err)
	require.NotNil(t, rt)

	assert.Equal(t, "child", rt.UserType)
	assert.Equal(t, int64(2), rt.UserID)
	assert.Equal(t, familyID, rt.FamilyID)
	assert.Equal(t, HashToken(rawToken), rt.TokenHash)
	assert.False(t, rt.CreatedAt.IsZero())
	assert.True(t, rt.ExpiresAt.After(rt.CreatedAt))
}

func TestRefreshTokenRepo_Validate_NotFound(t *testing.T) {
	repo := refreshTokenTestRepo(t)

	rt, err := repo.Validate("nonexistent-token")
	require.NoError(t, err)
	assert.Nil(t, rt)
}

func TestRefreshTokenRepo_Validate_Expired(t *testing.T) {
	repo := refreshTokenTestRepo(t)
	familyID, _ := seedFamilyAndParent(t, repo)

	rawToken, err := repo.Create("parent", 1, familyID, -time.Hour)
	require.NoError(t, err)

	rt, err := repo.Validate(rawToken)
	require.NoError(t, err)
	assert.Nil(t, rt, "expired refresh token should not be returned")
}

func TestRefreshTokenRepo_GetByHash(t *testing.T) {
	repo := refreshTokenTestRepo(t)
	familyID, _ := seedFamilyAndParent(t, repo)

	rawToken, err := repo.Create("parent", 1, familyID, time.Hour)
	require.NoError(t, err)

	tokenHash := HashToken(rawToken)
	rt, err := repo.GetByHash(tokenHash)
	require.NoError(t, err)
	require.NotNil(t, rt)
	assert.Equal(t, tokenHash, rt.TokenHash)
	assert.Equal(t, "parent", rt.UserType)
}

func TestRefreshTokenRepo_GetByHash_NotFound(t *testing.T) {
	repo := refreshTokenTestRepo(t)

	rt, err := repo.GetByHash("nonexistent-hash")
	require.NoError(t, err)
	assert.Nil(t, rt)
}

func TestRefreshTokenRepo_DeleteByHash(t *testing.T) {
	repo := refreshTokenTestRepo(t)
	familyID, _ := seedFamilyAndParent(t, repo)

	rawToken, err := repo.Create("parent", 1, familyID, time.Hour)
	require.NoError(t, err)

	tokenHash := HashToken(rawToken)
	err = repo.DeleteByHash(tokenHash)
	require.NoError(t, err)

	rt, err := repo.Validate(rawToken)
	require.NoError(t, err)
	assert.Nil(t, rt, "deleted refresh token should not validate")
}

func TestRefreshTokenRepo_DeleteExpired(t *testing.T) {
	repo := refreshTokenTestRepo(t)
	familyID, _ := seedFamilyAndParent(t, repo)

	// Create two expired tokens
	_, err := repo.Create("parent", 1, familyID, -time.Hour)
	require.NoError(t, err)
	_, err = repo.Create("child", 2, familyID, -2*time.Hour)
	require.NoError(t, err)

	// Create one valid token
	validToken, err := repo.Create("parent", 3, familyID, time.Hour)
	require.NoError(t, err)

	count, err := repo.DeleteExpired()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Valid token should still work
	rt, err := repo.Validate(validToken)
	require.NoError(t, err)
	assert.NotNil(t, rt)
}

func TestRefreshTokenRepo_DeleteAllForUser(t *testing.T) {
	repo := refreshTokenTestRepo(t)
	familyID, _ := seedFamilyAndParent(t, repo)

	// Create tokens for two different users
	_, err := repo.Create("parent", 1, familyID, time.Hour)
	require.NoError(t, err)
	_, err = repo.Create("parent", 1, familyID, time.Hour)
	require.NoError(t, err)
	otherToken, err := repo.Create("child", 2, familyID, time.Hour)
	require.NoError(t, err)

	err = repo.DeleteAllForUser("parent", 1)
	require.NoError(t, err)

	// Other user's token should still work
	rt, err := repo.Validate(otherToken)
	require.NoError(t, err)
	assert.NotNil(t, rt)
}
