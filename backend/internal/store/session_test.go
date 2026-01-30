package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		os.Remove(path)
	})
	return db
}

func TestCreateSession(t *testing.T) {
	db := testDB(t)
	ss := NewSessionStore(db)

	token, err := ss.Create("parent", 1, 10, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	// base64-url encoded 32 bytes = 44 characters
	assert.Len(t, token, 44)
}

func TestGetByToken(t *testing.T) {
	db := testDB(t)
	ss := NewSessionStore(db)

	token, err := ss.Create("child", 2, 10, time.Hour)
	require.NoError(t, err)

	sess, err := ss.GetByToken(token)
	require.NoError(t, err)
	require.NotNil(t, sess)

	assert.Equal(t, token, sess.Token)
	assert.Equal(t, "child", sess.UserType)
	assert.Equal(t, int64(2), sess.UserID)
	assert.Equal(t, int64(10), sess.FamilyID)
	assert.False(t, sess.CreatedAt.IsZero())
	assert.False(t, sess.ExpiresAt.IsZero())
	assert.True(t, sess.ExpiresAt.After(sess.CreatedAt))
}

func TestGetByToken_NotFound(t *testing.T) {
	db := testDB(t)
	ss := NewSessionStore(db)

	sess, err := ss.GetByToken("nonexistent-token")
	require.NoError(t, err)
	assert.Nil(t, sess)
}

func TestGetByToken_Expired(t *testing.T) {
	db := testDB(t)
	ss := NewSessionStore(db)

	token, err := ss.Create("parent", 1, 10, -time.Hour)
	require.NoError(t, err)

	sess, err := ss.GetByToken(token)
	require.NoError(t, err)
	assert.Nil(t, sess, "expired session should not be returned")
}

func TestDeleteByToken(t *testing.T) {
	db := testDB(t)
	ss := NewSessionStore(db)

	token, err := ss.Create("parent", 1, 10, time.Hour)
	require.NoError(t, err)

	err = ss.DeleteByToken(token)
	require.NoError(t, err)

	sess, err := ss.GetByToken(token)
	require.NoError(t, err)
	assert.Nil(t, sess, "session should be gone after delete")
}

func TestDeleteExpired(t *testing.T) {
	db := testDB(t)
	ss := NewSessionStore(db)

	// Create two expired sessions
	_, err := ss.Create("parent", 1, 10, -time.Hour)
	require.NoError(t, err)
	_, err = ss.Create("child", 2, 10, -2*time.Hour)
	require.NoError(t, err)

	// Create one valid session
	validToken, err := ss.Create("parent", 3, 10, time.Hour)
	require.NoError(t, err)

	count, err := ss.DeleteExpired()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Valid session should still exist
	sess, err := ss.GetByToken(validToken)
	require.NoError(t, err)
	assert.NotNil(t, sess)
}
