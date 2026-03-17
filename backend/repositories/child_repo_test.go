package repositories

import (
	"testing"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// createChildRepoFamily creates a test family directly via GORM.
func createChildRepoFamily(t *testing.T, repo *ChildRepo) *models.Family {
	t.Helper()
	fam := models.Family{Slug: "test-family-" + t.Name()}
	require.NoError(t, repo.db.Create(&fam).Error)
	return &fam
}

func TestChildRepo_Create(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.NotZero(t, child.ID)
	assert.Equal(t, fam.ID, child.FamilyID)
	assert.Equal(t, "Tommy", child.FirstName)
	assert.False(t, child.IsLocked)
	assert.Equal(t, 0, child.FailedLoginAttempts)
}

func TestChildRepo_Create_PasswordHashed(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	assert.NotEqual(t, "secret123", child.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(child.PasswordHash), []byte("secret123")))
}

func TestChildRepo_Create_DuplicateName(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	_, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	_, err = repo.Create(fam.ID, "Tommy", "different456", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestChildRepo_GetByID(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	created, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	found, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Tommy", found.FirstName)
}

func TestChildRepo_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)

	found, err := repo.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestChildRepo_GetByFamilyAndName(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	created, err := repo.Create(fam.ID, "Sarah", "pass123456", nil)
	require.NoError(t, err)

	found, err := repo.GetByFamilyAndName(fam.ID, "Sarah")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Sarah", found.FirstName)
}

func TestChildRepo_GetByFamilyAndName_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	found, err := repo.GetByFamilyAndName(fam.ID, "Nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestChildRepo_ListByFamily(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	_, err := repo.Create(fam.ID, "Zara", "pass123456", nil)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)

	children, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
	assert.Equal(t, "Zara", children[0].FirstName)
	assert.Equal(t, "Alice", children[1].FirstName)
}

func TestChildRepo_CheckPassword(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	assert.True(t, repo.CheckPassword(child, "secret123"))
	assert.False(t, repo.CheckPassword(child, "wrongpassword"))
}

func TestChildRepo_IncrementFailedAttempts(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	attempts, err := repo.IncrementFailedAttempts(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)

	attempts, err = repo.IncrementFailedAttempts(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestChildRepo_LockAccount(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.False(t, child.IsLocked)

	err = repo.LockAccount(child.ID)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.True(t, updated.IsLocked)
}

func TestChildRepo_ResetFailedAttempts(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	repo.IncrementFailedAttempts(child.ID)
	repo.IncrementFailedAttempts(child.ID)

	err = repo.ResetFailedAttempts(child.ID)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, updated.FailedLoginAttempts)
}

func TestChildRepo_UpdatePassword(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "oldpass123", nil)
	require.NoError(t, err)

	repo.LockAccount(child.ID)
	repo.IncrementFailedAttempts(child.ID)

	err = repo.UpdatePassword(child.ID, "newpass456")
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.False(t, updated.IsLocked)
	assert.Equal(t, 0, updated.FailedLoginAttempts)
	assert.True(t, repo.CheckPassword(updated, "newpass456"))
	assert.False(t, repo.CheckPassword(updated, "oldpass123"))
}

func TestChildRepo_UpdateName(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	err = repo.UpdateNameAndAvatar(child.ID, fam.ID, "Thomas", nil, false)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.Equal(t, "Thomas", updated.FirstName)
}

func TestChildRepo_UpdateName_DuplicateRejected(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	_, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	child2, err := repo.Create(fam.ID, "Sarah", "secret123", nil)
	require.NoError(t, err)

	err = repo.UpdateNameAndAvatar(child2.ID, fam.ID, "Tommy", nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestChildRepo_GetBalance(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	balance, err := repo.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), balance)
}

func TestChildRepo_GetBalance_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)

	_, err := repo.GetBalance(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestChildRepo_Delete(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	err = repo.Delete(child.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestChildRepo_Delete_DoesNotAffectOtherChildren(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child1, err := repo.Create(fam.ID, "Alice", "secret123", nil)
	require.NoError(t, err)
	child2, err := repo.Create(fam.ID, "Bob", "secret123", nil)
	require.NoError(t, err)

	err = repo.Delete(child1.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(child2.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Bob", found.FirstName)
}

func TestChildRepo_Create_WithAvatar(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	avatar := "🌻"
	child, err := repo.Create(fam.ID, "Tommy", "secret123", &avatar)
	require.NoError(t, err)
	assert.NotNil(t, child.Avatar)
	assert.Equal(t, "🌻", *child.Avatar)

	fetched, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetched.Avatar)
	assert.Equal(t, "🌻", *fetched.Avatar)
}

func TestChildRepo_Create_WithoutAvatar(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.Nil(t, child.Avatar)
}

func TestChildRepo_ListByFamily_ReturnsAvatar(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	avatar := "🦋"
	_, err := repo.Create(fam.ID, "Alice", "pass123456", &avatar)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)

	children, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
	assert.NotNil(t, children[0].Avatar)
	assert.Equal(t, "🦋", *children[0].Avatar)
	assert.Nil(t, children[1].Avatar)
}

func TestChildRepo_UpdateNameAndAvatar_SetAvatar(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.Nil(t, child.Avatar)

	avatar := "🐸"
	err = repo.UpdateNameAndAvatar(child.ID, fam.ID, "Tommy", &avatar, true)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.NotNil(t, updated.Avatar)
	assert.Equal(t, "🐸", *updated.Avatar)
	assert.Equal(t, "Tommy", updated.FirstName)
}

func TestChildRepo_UpdateNameAndAvatar_ClearAvatar(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	avatar := "🐸"
	child, err := repo.Create(fam.ID, "Tommy", "secret123", &avatar)
	require.NoError(t, err)
	assert.NotNil(t, child.Avatar)

	err = repo.UpdateNameAndAvatar(child.ID, fam.ID, "Tommy", nil, true)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.Avatar)
}

func TestChildRepo_UpdateNameAndAvatar_PreservesAvatarWhenNotSet(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	avatar := "🐸"
	child, err := repo.Create(fam.ID, "Tommy", "secret123", &avatar)
	require.NoError(t, err)

	err = repo.UpdateNameAndAvatar(child.ID, fam.ID, "Thomas", nil, false)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.Equal(t, "Thomas", updated.FirstName)
	assert.NotNil(t, updated.Avatar)
	assert.Equal(t, "🐸", *updated.Avatar)
}

func TestChildRepo_UpdateTheme(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.Nil(t, child.Theme)

	err = repo.UpdateTheme(child.ID, "piggybank")
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Theme)
	assert.Equal(t, "piggybank", *updated.Theme)

	err = repo.UpdateTheme(child.ID, "sparkle")
	require.NoError(t, err)

	updated, err = repo.GetByID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Theme)
	assert.Equal(t, "sparkle", *updated.Theme)
}

func TestChildRepo_UpdateTheme_NonexistentChild(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)

	err := repo.UpdateTheme(99999, "sparkle")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestChildRepo_UpdateAvatar(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)

	avatar := "🐸"
	err = repo.UpdateAvatar(child.ID, &avatar)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.NotNil(t, updated.Avatar)
	assert.Equal(t, "🐸", *updated.Avatar)

	// Clear avatar
	err = repo.UpdateAvatar(child.ID, nil)
	require.NoError(t, err)

	updated, err = repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.Avatar)
}

func TestChildRepo_UpdateAvatar_NonexistentChild(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)

	avatar := "🐸"
	err := repo.UpdateAvatar(99999, &avatar)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestChildRepo_GetByFamilyAndName_ReturnsTheme(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Sarah", "pass123456", nil)
	require.NoError(t, err)

	err = repo.UpdateTheme(child.ID, "sparkle")
	require.NoError(t, err)

	found, err := repo.GetByFamilyAndName(fam.ID, "Sarah")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.NotNil(t, found.Theme)
	assert.Equal(t, "sparkle", *found.Theme)
}

func TestChildRepo_ListByFamily_ReturnsTheme(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	_, err := repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	child2, err := repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)

	err = repo.UpdateTheme(child2.ID, "piggybank")
	require.NoError(t, err)

	children, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
	assert.Nil(t, children[0].Theme)
	require.NotNil(t, children[1].Theme)
	assert.Equal(t, "piggybank", *children[1].Theme)
}

func TestChildRepo_CountByFamily(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	count, err := repo.CountByFamily(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	_, err = repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)

	count, err = repo.CountByFamily(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestChildRepo_SetDisabled(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child, err := repo.Create(fam.ID, "Tommy", "secret123", nil)
	require.NoError(t, err)
	assert.False(t, child.IsDisabled)

	err = repo.SetDisabled(child.ID, true)
	require.NoError(t, err)

	updated, err := repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.True(t, updated.IsDisabled)

	err = repo.SetDisabled(child.ID, false)
	require.NoError(t, err)

	updated, err = repo.GetByID(child.ID)
	require.NoError(t, err)
	assert.False(t, updated.IsDisabled)
}

func TestChildRepo_CountEnabledByFamily(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	_, err := repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	child2, err := repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Charlie", "pass123456", nil)
	require.NoError(t, err)

	count, err := repo.CountEnabledByFamily(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	err = repo.SetDisabled(child2.ID, true)
	require.NoError(t, err)

	count, err = repo.CountEnabledByFamily(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestChildRepo_EnableAllChildren(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child1, err := repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	child2, err := repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)

	repo.SetDisabled(child1.ID, true)
	repo.SetDisabled(child2.ID, true)

	count, err := repo.CountEnabledByFamily(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	err = repo.EnableAllChildren(fam.ID)
	require.NoError(t, err)

	count, err = repo.CountEnabledByFamily(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestChildRepo_DisableExcessChildren(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	_, err := repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Charlie", "pass123456", nil)
	require.NoError(t, err)

	err = repo.DisableExcessChildren(fam.ID, 1)
	require.NoError(t, err)

	children, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.False(t, children[0].IsDisabled)
	assert.True(t, children[1].IsDisabled)
	assert.True(t, children[2].IsDisabled)
}

func TestChildRepo_ReconcileChildLimits(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	child1, err := repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	child2, err := repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)
	child3, err := repo.Create(fam.ID, "Charlie", "pass123456", nil)
	require.NoError(t, err)

	repo.SetDisabled(child1.ID, true)
	repo.SetDisabled(child2.ID, true)
	repo.SetDisabled(child3.ID, true)

	err = repo.ReconcileChildLimits(fam.ID, 2)
	require.NoError(t, err)

	children, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, children, 3)
	assert.False(t, children[0].IsDisabled)
	assert.False(t, children[1].IsDisabled)
	assert.True(t, children[2].IsDisabled)
}

func TestChildRepo_ReconcileChildLimits_DisablesExcess(t *testing.T) {
	db := testDB(t)
	repo := NewChildRepo(db)
	fam := createChildRepoFamily(t, repo)

	_, err := repo.Create(fam.ID, "Alice", "pass123456", nil)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Bob", "pass123456", nil)
	require.NoError(t, err)
	_, err = repo.Create(fam.ID, "Charlie", "pass123456", nil)
	require.NoError(t, err)

	err = repo.ReconcileChildLimits(fam.ID, 1)
	require.NoError(t, err)

	children, err := repo.ListByFamily(fam.ID)
	require.NoError(t, err)
	assert.False(t, children[0].IsDisabled)
	assert.True(t, children[1].IsDisabled)
	assert.True(t, children[2].IsDisabled)
}
