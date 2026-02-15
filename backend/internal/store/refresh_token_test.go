package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func refreshTokenTestDB(t *testing.T) *RefreshTokenStore {
	t.Helper()
	db := testDB(t)
	return NewRefreshTokenStore(db)
}

func TestRefreshToken_Create(t *testing.T) {
	rts := refreshTokenTestDB(t)

	token1, err := rts.Create("parent", 1, 10, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := rts.Create("parent", 1, 10, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)
}

func TestRefreshToken_Validate(t *testing.T) {
	rts := refreshTokenTestDB(t)

	rawToken, err := rts.Create("child", 2, 10, time.Hour)
	require.NoError(t, err)

	rt, err := rts.Validate(rawToken)
	require.NoError(t, err)
	require.NotNil(t, rt)

	assert.Equal(t, "child", rt.UserType)
	assert.Equal(t, int64(2), rt.UserID)
	assert.Equal(t, int64(10), rt.FamilyID)
	assert.Equal(t, HashToken(rawToken), rt.TokenHash)
	assert.False(t, rt.CreatedAt.IsZero())
	assert.True(t, rt.ExpiresAt.After(rt.CreatedAt))
}

func TestRefreshToken_Validate_NotFound(t *testing.T) {
	rts := refreshTokenTestDB(t)

	rt, err := rts.Validate("nonexistent-token")
	require.NoError(t, err)
	assert.Nil(t, rt)
}

func TestRefreshToken_Validate_Expired(t *testing.T) {
	rts := refreshTokenTestDB(t)

	rawToken, err := rts.Create("parent", 1, 10, -time.Hour)
	require.NoError(t, err)

	rt, err := rts.Validate(rawToken)
	require.NoError(t, err)
	assert.Nil(t, rt, "expired refresh token should not be returned")
}

func TestRefreshToken_DeleteByHash(t *testing.T) {
	rts := refreshTokenTestDB(t)

	rawToken, err := rts.Create("parent", 1, 10, time.Hour)
	require.NoError(t, err)

	tokenHash := HashToken(rawToken)
	err = rts.DeleteByHash(tokenHash)
	require.NoError(t, err)

	rt, err := rts.Validate(rawToken)
	require.NoError(t, err)
	assert.Nil(t, rt, "deleted refresh token should not validate")
}

func TestRefreshToken_DeleteExpired(t *testing.T) {
	rts := refreshTokenTestDB(t)

	// Create two expired tokens
	_, err := rts.Create("parent", 1, 10, -time.Hour)
	require.NoError(t, err)
	_, err = rts.Create("child", 2, 10, -2*time.Hour)
	require.NoError(t, err)

	// Create one valid token
	validToken, err := rts.Create("parent", 3, 10, time.Hour)
	require.NoError(t, err)

	count, err := rts.DeleteExpired()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Valid token should still work
	rt, err := rts.Validate(validToken)
	require.NoError(t, err)
	assert.NotNil(t, rt)
}

func TestRefreshToken_DeleteByUser(t *testing.T) {
	rts := refreshTokenTestDB(t)

	// Create tokens for two different users
	_, err := rts.Create("parent", 1, 10, time.Hour)
	require.NoError(t, err)
	_, err = rts.Create("parent", 1, 10, time.Hour)
	require.NoError(t, err)
	otherToken, err := rts.Create("child", 2, 10, time.Hour)
	require.NoError(t, err)

	err = rts.DeleteByUser("parent", 1)
	require.NoError(t, err)

	// Other user's token should still work
	rt, err := rts.Validate(otherToken)
	require.NoError(t, err)
	assert.NotNil(t, rt)
}
