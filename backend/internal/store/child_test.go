package store

import (
	"testing"
	"time"

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

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
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

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
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

	_, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	_, err = cs.Create(fam.ID, "Tommy", "different456", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetByFamilyAndName(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	created, err := cs.Create(fam.ID, "Sarah", "pass123456", nil)
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

	_, err := cs.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	_, err = cs.Create(fam.ID, "Bob", "pass123456", nil)
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

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	assert.True(t, cs.CheckPassword(child, "secret123"))
	assert.False(t, cs.CheckPassword(child, "wrongpassword"))
}

func TestIncrementFailedAttempts(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
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

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
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

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
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

	child, err := cs.Create(fam.ID, "Tommy", "oldpass123", nil)
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

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	err = cs.UpdateNameAndAvatar(child.ID, fam.ID, "Thomas", nil, false)
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Equal(t, "Thomas", updated.FirstName)
}

func TestUpdateName_DuplicateRejected(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	_, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	child2, err := cs.Create(fam.ID, "Sarah", "secret123", nil)
	require.NoError(t, err)

	err = cs.UpdateNameAndAvatar(child2.ID, fam.ID, "Tommy", nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// T008: Tests for ChildStore.GetBalance()
func TestGetBalance(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	// Initial balance should be 0
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), balance)
}

func TestGetBalance_NotFound(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)

	_, err := cs.GetBalance(99999) // Non-existent ID
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteChild(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	err = cs.Delete(child.ID)
	require.NoError(t, err)

	// Child should no longer exist
	found, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestDeleteChild_CascadesTransactions(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)
	ps := NewParentStore(db)
	parent, err := ps.Create("g-123", "p@test.com", "Parent")
	require.NoError(t, err)
	require.NoError(t, ps.SetFamilyID(parent.ID, fam.ID))

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	ts := NewTransactionStore(db)
	_, _, err = ts.Deposit(child.ID, parent.ID, 1000, "test deposit")
	require.NoError(t, err)

	err = cs.Delete(child.ID)
	require.NoError(t, err)

	// Transactions should be gone (cascade)
	var count int
	err = db.Read.QueryRow(`SELECT COUNT(*) FROM transactions WHERE child_id = ?`, child.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteChild_CleansUpSessions(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	ss := NewSessionStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	token, err := ss.Create("child", child.ID, fam.ID, time.Hour)
	require.NoError(t, err)

	err = cs.Delete(child.ID)
	require.NoError(t, err)

	// Session should be gone
	sess, err := ss.GetByToken(token)
	require.NoError(t, err)
	assert.Nil(t, sess)
}

func TestDeleteChild_CleansUpAuthEvents(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	es := NewAuthEventStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	err = es.LogEvent(AuthEvent{
		EventType: "login_success",
		UserType:  "child",
		UserID:    child.ID,
		FamilyID:  fam.ID,
		IPAddress: "127.0.0.1",
		CreatedAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	err = cs.Delete(child.ID)
	require.NoError(t, err)

	// Auth events for this child should be gone
	var count int
	err = db.Read.QueryRow(`SELECT COUNT(*) FROM auth_events WHERE user_type = 'child' AND user_id = ?`, child.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteChild_DoesNotAffectOtherChildren(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child1, err := cs.Create(fam.ID, "Alice", "secret123", nil)
	require.NoError(t, err)
	child2, err := cs.Create(fam.ID, "Bob", "secret123", nil)
	require.NoError(t, err)

	err = cs.Delete(child1.ID)
	require.NoError(t, err)

	// Other child should still exist
	found, err := cs.GetByID(child2.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Bob", found.FirstName)
}

// T005: Avatar tests for Create
func TestCreateChild_WithAvatar(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	avatar := "🌻"
	child, err := cs.Create(fam.ID, "Tommy", "secret123", &avatar)
	require.NoError(t, err)
	assert.NotNil(t, child.Avatar)
	assert.Equal(t, "🌻", *child.Avatar)

	// Verify via GetByID
	fetched, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetched.Avatar)
	assert.Equal(t, "🌻", *fetched.Avatar)
}

func TestCreateChild_WithoutAvatar(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.Nil(t, child.Avatar)
}

func TestListByFamily_ReturnsAvatar(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	avatar := "🦋"
	_, err := cs.Create(fam.ID, "Alice", "pass123456", &avatar)
	require.NoError(t, err)
	_, err = cs.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)

	children, err := cs.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
	// Alice has avatar
	assert.NotNil(t, children[0].Avatar)
	assert.Equal(t, "🦋", *children[0].Avatar)
	// Bob has no avatar
	assert.Nil(t, children[1].Avatar)
}

// T011: Avatar tests for UpdateNameAndAvatar
func TestUpdateNameAndAvatar_SetAvatar(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	child, err := cs.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.Nil(t, child.Avatar)

	avatar := "🐸"
	err = cs.UpdateNameAndAvatar(child.ID, fam.ID, "Tommy", &avatar, true)
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.NotNil(t, updated.Avatar)
	assert.Equal(t, "🐸", *updated.Avatar)
	assert.Equal(t, "Tommy", updated.FirstName)
}

func TestUpdateNameAndAvatar_ClearAvatar(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	avatar := "🐸"
	child, err := cs.Create(fam.ID, "Tommy", "secret123", &avatar)
	require.NoError(t, err)
	assert.NotNil(t, child.Avatar)

	err = cs.UpdateNameAndAvatar(child.ID, fam.ID, "Tommy", nil, true)
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.Avatar)
}

func TestUpdateNameAndAvatar_PreservesAvatarWhenNotSet(t *testing.T) {
	db := testDB(t)
	cs := NewChildStore(db)
	fam := createTestFamily(t, db)

	avatar := "🐸"
	child, err := cs.Create(fam.ID, "Tommy", "secret123", &avatar)
	require.NoError(t, err)

	// Update name only, avatarSet=false should preserve avatar
	err = cs.UpdateNameAndAvatar(child.ID, fam.ID, "Thomas", nil, false)
	require.NoError(t, err)

	updated, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Equal(t, "Thomas", updated.FirstName)
	assert.NotNil(t, updated.Avatar)
	assert.Equal(t, "🐸", *updated.Avatar)
}
