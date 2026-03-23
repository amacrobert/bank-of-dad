package repositories

import (
	"testing"
	"time"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChoreInstanceRepo_CreateInstance(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "InstanceKid")

	choreRepo := NewChoreRepo(db)
	desc := "Sweep the porch"
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Sweep Porch",
		Description:       &desc,
		RewardCents:       300,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	})
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(24 * time.Hour)
	end := now.Add(7 * 24 * time.Hour)
	instance := &models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 300, // snapshotted from chore
		Status:      models.ChoreInstanceStatusAvailable,
		PeriodStart: &now,
		PeriodEnd:   &end,
	}

	created, err := repo.CreateInstance(instance)
	require.NoError(t, err)
	assert.True(t, created.ID > 0)
	assert.Equal(t, chore.ID, created.ChoreID)
	assert.Equal(t, child.ID, created.ChildID)
	assert.Equal(t, 300, created.RewardCents)
	assert.Equal(t, models.ChoreInstanceStatusAvailable, created.Status)
	assert.NotNil(t, created.PeriodStart)
	assert.NotNil(t, created.PeriodEnd)
	assert.Nil(t, created.CompletedAt)
	assert.Nil(t, created.ReviewedAt)
	assert.Nil(t, created.TransactionID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())
}

func TestChoreInstanceRepo_GetByID(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "GetKid")

	choreRepo := NewChoreRepo(db)
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Wash Car",
		RewardCents:       500,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	})
	require.NoError(t, err)

	created, err := repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 500,
		Status:      models.ChoreInstanceStatusAvailable,
	})
	require.NoError(t, err)

	// Found
	fetched, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, chore.ID, fetched.ChoreID)
	assert.Equal(t, 500, fetched.RewardCents)

	// Not found returns (nil, nil)
	notFound, err := repo.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestChoreInstanceRepo_ListByChild(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "ListKid")

	choreRepo := NewChoreRepo(db)
	desc := "Clean the kitchen"
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Clean Kitchen",
		Description:       &desc,
		RewardCents:       200,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	// Create instances in different statuses
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 200,
		Status:      models.ChoreInstanceStatusAvailable,
	})
	require.NoError(t, err)

	completedAt := time.Now().UTC()
	reviewedAt := time.Now().UTC()
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 200,
		Status:      models.ChoreInstanceStatusPendingApproval,
		CompletedAt: &completedAt,
	})
	require.NoError(t, err)

	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 200,
		Status:      models.ChoreInstanceStatusApproved,
		CompletedAt: &completedAt,
		ReviewedAt:  &reviewedAt,
	})
	require.NoError(t, err)

	// Create an expired instance (should NOT appear)
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 200,
		Status:      models.ChoreInstanceStatusExpired,
	})
	require.NoError(t, err)

	available, pending, completed, err := repo.ListByChild(child.ID)
	require.NoError(t, err)

	assert.Len(t, available, 1)
	assert.Len(t, pending, 1)
	assert.Len(t, completed, 1)

	// Verify join fields are populated
	assert.Equal(t, "Clean Kitchen", available[0].ChoreName)
	assert.NotNil(t, available[0].ChoreDescription)
	assert.Equal(t, "Clean the kitchen", *available[0].ChoreDescription)

	assert.Equal(t, "Clean Kitchen", pending[0].ChoreName)
	assert.Equal(t, "Clean Kitchen", completed[0].ChoreName)
}

func TestChoreInstanceRepo_ListPendingByFamily(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child1 := createChoreTestChild(t, db, fam.ID, "PendingAlice")
	child2 := createChoreTestChild(t, db, fam.ID, "PendingBob")

	choreRepo := NewChoreRepo(db)
	chore1, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Dust Shelves",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	chore2, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Fold Laundry",
		RewardCents:       150,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	})
	require.NoError(t, err)

	completedAt1 := time.Now().UTC().Add(-2 * time.Hour)
	completedAt2 := time.Now().UTC().Add(-1 * time.Hour)

	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore1.ID,
		ChildID:     child1.ID,
		RewardCents: 100,
		Status:      models.ChoreInstanceStatusPendingApproval,
		CompletedAt: &completedAt1,
	})
	require.NoError(t, err)

	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore2.ID,
		ChildID:     child2.ID,
		RewardCents: 150,
		Status:      models.ChoreInstanceStatusPendingApproval,
		CompletedAt: &completedAt2,
	})
	require.NoError(t, err)

	// An available instance should NOT appear
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore1.ID,
		ChildID:     child1.ID,
		RewardCents: 100,
		Status:      models.ChoreInstanceStatusAvailable,
	})
	require.NoError(t, err)

	results, err := repo.ListPendingByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Ordered by completed_at ASC (oldest first)
	assert.Equal(t, "Dust Shelves", results[0].ChoreName)
	assert.Equal(t, "PendingAlice", results[0].ChildName)
	assert.Equal(t, "Fold Laundry", results[1].ChoreName)
	assert.Equal(t, "PendingBob", results[1].ChildName)

	// Verify family scoping: create another family with a pending instance
	fr := NewFamilyRepo(db)
	fam2, err := fr.Create("chore-test-family-2")
	require.NoError(t, err)
	pr := NewParentRepo(db)
	p2, err := pr.Create("google-chore-test-other-123", "choreother@example.com", "Other Parent")
	require.NoError(t, err)
	err = pr.SetFamilyID(p2.ID, fam2.ID)
	require.NoError(t, err)
	parent2 := p2
	child3 := createChoreTestChild(t, db, fam2.ID, "OtherFamilyKid")
	chore3, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam2.ID,
		CreatedByParentID: parent2.ID,
		Name:              "Other Chore",
		RewardCents:       50,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)
	completedAt3 := time.Now().UTC()
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore3.ID,
		ChildID:     child3.ID,
		RewardCents: 50,
		Status:      models.ChoreInstanceStatusPendingApproval,
		CompletedAt: &completedAt3,
	})
	require.NoError(t, err)

	// Original family should still only see 2
	results, err = repo.ListPendingByFamily(fam.ID)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestChoreInstanceRepo_MarkComplete(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "CompleteKid")
	otherChild := createChoreTestChild(t, db, fam.ID, "OtherKid")

	choreRepo := NewChoreRepo(db)
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Make Bed",
		RewardCents:       50,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	instance, err := repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 50,
		Status:      models.ChoreInstanceStatusAvailable,
	})
	require.NoError(t, err)

	// Happy path: available -> pending_approval
	err = repo.MarkComplete(instance.ID, child.ID)
	require.NoError(t, err)

	fetched, err := repo.GetByID(instance.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusPendingApproval, fetched.Status)
	assert.NotNil(t, fetched.CompletedAt)
	assert.Nil(t, fetched.RejectionReason)

	// Wrong child should fail
	instance2, err := repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 50,
		Status:      models.ChoreInstanceStatusAvailable,
	})
	require.NoError(t, err)

	err = repo.MarkComplete(instance2.ID, otherChild.ID)
	assert.Error(t, err)

	// Already pending should fail
	err = repo.MarkComplete(instance.ID, child.ID)
	assert.Error(t, err)
}

func TestChoreInstanceRepo_Approve(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "ApproveKid")

	choreRepo := NewChoreRepo(db)
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Wash Windows",
		RewardCents:       400,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	})
	require.NoError(t, err)

	instance, err := repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 400,
		Status:      models.ChoreInstanceStatusPendingApproval,
	})
	require.NoError(t, err)

	// Create a real transaction for the FK constraint
	txn := &models.Transaction{
		ChildID:         child.ID,
		ParentID:        parent.ID,
		AmountCents:     400,
		TransactionType: models.TransactionTypeChore,
	}
	require.NoError(t, db.Create(txn).Error)
	txnID := txn.ID
	err = repo.Approve(instance.ID, parent.ID, &txnID)
	require.NoError(t, err)

	fetched, err := repo.GetByID(instance.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusApproved, fetched.Status)
	assert.NotNil(t, fetched.ReviewedAt)
	assert.Equal(t, parent.ID, *fetched.ReviewedByParentID)
	assert.Equal(t, txnID, *fetched.TransactionID)

	// Already approved should fail
	err = repo.Approve(instance.ID, parent.ID, &txnID)
	assert.Error(t, err)
}

func TestChoreInstanceRepo_Reject(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "RejectKid")

	choreRepo := NewChoreRepo(db)
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Organize Closet",
		RewardCents:       250,
		Recurrence:        models.ChoreRecurrenceWeekly,
		IsActive:          true,
	})
	require.NoError(t, err)

	completedAt := time.Now().UTC()
	instance, err := repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 250,
		Status:      models.ChoreInstanceStatusPendingApproval,
		CompletedAt: &completedAt,
	})
	require.NoError(t, err)

	err = repo.Reject(instance.ID, parent.ID, "Not done properly")
	require.NoError(t, err)

	fetched, err := repo.GetByID(instance.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusAvailable, fetched.Status)
	assert.NotNil(t, fetched.ReviewedAt)
	assert.Equal(t, parent.ID, *fetched.ReviewedByParentID)
	assert.Equal(t, "Not done properly", *fetched.RejectionReason)
	assert.Nil(t, fetched.CompletedAt) // completed_at should be cleared

	// Reject on available should fail (not pending_approval)
	err = repo.Reject(instance.ID, parent.ID, "Another reason")
	assert.Error(t, err)
}

func TestChoreInstanceRepo_ExpireByPeriod(t *testing.T) {
	db := testDB(t)
	repo := NewChoreInstanceRepo(db)

	fam := createChoreTestFamily(t, db)
	parent := createChoreTestParent(t, db, fam.ID)
	child := createChoreTestChild(t, db, fam.ID, "ExpireKid")

	choreRepo := NewChoreRepo(db)
	chore, err := choreRepo.Create(&models.Chore{
		FamilyID:          fam.ID,
		CreatedByParentID: parent.ID,
		Name:              "Water Garden",
		RewardCents:       75,
		Recurrence:        models.ChoreRecurrenceDaily,
		IsActive:          true,
	})
	require.NoError(t, err)

	pastEnd := time.Now().UTC().Add(-48 * time.Hour).Truncate(24 * time.Hour)
	futureEnd := time.Now().UTC().Add(48 * time.Hour).Truncate(24 * time.Hour)

	// Instance with past period_end (should expire)
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 75,
		Status:      models.ChoreInstanceStatusAvailable,
		PeriodEnd:   &pastEnd,
	})
	require.NoError(t, err)

	// Another with past period_end (should expire)
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 75,
		Status:      models.ChoreInstanceStatusAvailable,
		PeriodEnd:   &pastEnd,
	})
	require.NoError(t, err)

	// Instance with future period_end (should NOT expire)
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 75,
		Status:      models.ChoreInstanceStatusAvailable,
		PeriodEnd:   &futureEnd,
	})
	require.NoError(t, err)

	// Pending instance with past period_end (should NOT expire - only 'available' expires)
	_, err = repo.CreateInstance(&models.ChoreInstance{
		ChoreID:     chore.ID,
		ChildID:     child.ID,
		RewardCents: 75,
		Status:      models.ChoreInstanceStatusPendingApproval,
		PeriodEnd:   &pastEnd,
	})
	require.NoError(t, err)

	count, err := repo.ExpireByPeriod(time.Now().UTC())
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Verify the expired instances
	var expiredCount int64
	db.Model(&models.ChoreInstance{}).Where("status = ?", models.ChoreInstanceStatusExpired).Count(&expiredCount)
	assert.Equal(t, int64(2), expiredCount)

	// Running again should expire 0
	count, err = repo.ExpireByPeriod(time.Now().UTC())
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
