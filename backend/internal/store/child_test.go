package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func createTestFamily(t *testing.T, db *DB) *Family {
	t.Helper()
	fs := NewFamilyStore(db)
	f, err := fs.Create("test-family")
	require.NoError(t, err)
	return f
}

func TestCreateChild(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)
	assert.NotZero(t, child.ID)
	assert.Equal(t, fam.ID, child.FamilyID)
	assert.Equal(t, "Tommy", child.FirstName)
	assert.False(t, child.IsLocked)
	assert.Equal(t, 0, child.FailedLoginAttempts)
}

func TestCreateChild_PasswordHashed(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)

	// Password hash should not be plaintext
	assert.NotEqual(t, "secret123", child.PasswordHash)
	// Should be a valid bcrypt hash
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(child.PasswordHash), []byte("secret123")))
}

func TestCreateChild_DuplicateName(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	_, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)

	_, err = cs.Create(fam.ID, "Tommy", "different456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetByFamilyAndName(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	created, err := cs.Create(fam.ID, "Sarah", "pass123456")
	require.NoError(t, err)

	found, err := cs.GetByFamilyAndName(fam.ID, "Sarah")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Sarah", found.FirstName)
}

func TestGetByFamilyAndName_NotFound(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	found, err := cs.GetByFamilyAndName(fam.ID, "Nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestListByFamily(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	_, err := cs.Create(fam.ID, "Alice", "pass123456")
	require.NoError(t, err)
	_, err = cs.Create(fam.ID, "Bob", "pass123456")
	require.NoError(t, err)

	children, err := cs.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
	// Should be sorted alphabetically
	assert.Equal(t, "Alice", children[0].FirstName)
	assert.Equal(t, "Bob", children[1].FirstName)
}

func TestCheckPassword(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)

	assert.True(t, cs.CheckPassword(child, "secret123"))
	assert.False(t, cs.CheckPassword(child, "wrongpassword"))
}

func TestIncrementFailedAttempts(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)

	attempts, err := cs.IncrementFailedAttempts(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)

	attempts, err = cs.IncrementFailedAttempts(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestLockAccount(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)
	assert.False(t, child.IsLocked)

	err = cs.LockAccount(child.ID)
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.True(t, updated.IsLocked)
}

func TestResetFailedAttempts(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)

	cs.IncrementFailedAttempts(child.ID)
	cs.IncrementFailedAttempts(child.ID)

	err = cs.ResetFailedAttempts(child.ID)
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, updated.FailedLoginAttempts)
}

func TestUpdatePassword(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "oldpass123")
	require.NoError(t, err)

	// Lock the account first
	cs.LockAccount(child.ID)
	cs.IncrementFailedAttempts(child.ID)

	err = cs.UpdatePassword(child.ID, "newpass456")
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	// Password reset should unlock and reset attempts
	assert.False(t, updated.IsLocked)
	assert.Equal(t, 0, updated.FailedLoginAttempts)
	// New password should work
	assert.True(t, cs.CheckPassword(updated, "newpass456"))
	// Old password should not
	assert.False(t, cs.CheckPassword(updated, "oldpass123"))
}

func TestUpdateName(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)

	err = cs.UpdateName(child.ID, fam.ID, "Thomas")
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Equal(t, "Thomas", updated.FirstName)
}

func TestUpdateName_DuplicateRejected(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	_, err := cs.Create(fam.ID, "Tommy", "secret123")
	require.NoError(t, err)
	child2, err := cs.Create(fam.ID, "Sarah", "secret123")
	require.NoError(t, err)

	err = cs.UpdateName(child2.ID, fam.ID, "Tommy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
