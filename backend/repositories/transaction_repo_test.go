package repositories

import (
	"testing"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTransactionTest(t *testing.T) (*gorm.DB, *models.Family, *models.Parent, *models.Child, *TransactionRepo) {
	t.Helper()
	db := testDB(t)
	tr := NewTransactionRepo(db)

	fam := &models.Family{Slug: "test-family"}
	require.NoError(t, db.Create(fam).Error)

	parent := &models.Parent{
		GoogleID:    "google-test-123",
		Email:       "test@example.com",
		DisplayName: "Test Parent",
		FamilyID:    fam.ID,
	}
	require.NoError(t, db.Create(parent).Error)

	child := &models.Child{
		FamilyID:     fam.ID,
		FirstName:    "TestChild",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, db.Create(child).Error)

	return db, fam, parent, child, tr
}

// T005: Tests for TransactionRepo.Deposit()
func TestDeposit(t *testing.T) {
	db, _, parent, child, tr := setupTransactionTest(t)

	// Verify initial balance is 0
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(0), c.BalanceCents)

	// Test deposit
	tx, newBalance, err := tr.Deposit(child.ID, parent.ID, 1000, "Weekly allowance")
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, int64(1000), tx.AmountCents)
	assert.Equal(t, models.TransactionTypeDeposit, tx.TransactionType)
	assert.NotNil(t, tx.Note)
	assert.Equal(t, "Weekly allowance", *tx.Note)
	assert.Equal(t, child.ID, tx.ChildID)
	assert.Equal(t, parent.ID, tx.ParentID)
	assert.Equal(t, int64(1000), newBalance)

	// Verify balance updated
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(1000), c.BalanceCents)
}

func TestDeposit_EmptyNote(t *testing.T) {
	_, _, parent, child, tr := setupTransactionTest(t)

	tx, _, err := tr.Deposit(child.ID, parent.ID, 500, "")
	require.NoError(t, err)
	assert.Nil(t, tx.Note) // Empty note should be nil
}

func TestDeposit_InvalidAmount(t *testing.T) {
	_, _, parent, child, tr := setupTransactionTest(t)

	// Zero amount
	_, _, err := tr.Deposit(child.ID, parent.ID, 0, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")

	// Negative amount
	_, _, err = tr.Deposit(child.ID, parent.ID, -100, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

func TestDeposit_MultipleDeposits(t *testing.T) {
	db, _, parent, child, tr := setupTransactionTest(t)

	// First deposit
	_, newBalance1, err := tr.Deposit(child.ID, parent.ID, 1000, "First")
	require.NoError(t, err)
	assert.Equal(t, int64(1000), newBalance1)

	// Second deposit
	_, newBalance2, err := tr.Deposit(child.ID, parent.ID, 500, "Second")
	require.NoError(t, err)
	assert.Equal(t, int64(1500), newBalance2)

	// Verify final balance
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(1500), c.BalanceCents)
}

// T006: Tests for TransactionRepo.Withdraw()
func TestWithdraw(t *testing.T) {
	db, _, parent, child, tr := setupTransactionTest(t)

	// Deposit first
	_, _, err := tr.Deposit(child.ID, parent.ID, 5000, "Initial deposit")
	require.NoError(t, err)

	// Withdraw
	tx, newBalance, err := tr.Withdraw(child.ID, parent.ID, 1500, "Bought a book")
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, int64(1500), tx.AmountCents)
	assert.Equal(t, models.TransactionTypeWithdrawal, tx.TransactionType)
	assert.NotNil(t, tx.Note)
	assert.Equal(t, "Bought a book", *tx.Note)
	assert.Equal(t, int64(3500), newBalance)

	// Verify balance updated
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(3500), c.BalanceCents)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	db, _, parent, child, tr := setupTransactionTest(t)

	// Deposit $20
	_, _, err := tr.Deposit(child.ID, parent.ID, 2000, "")
	require.NoError(t, err)

	// Try to withdraw $25
	_, currentBalance, err := tr.Withdraw(child.ID, parent.ID, 2500, "")
	assert.ErrorIs(t, err, models.ErrInsufficientFunds)
	assert.Equal(t, int64(2000), currentBalance) // Returns current balance for error message

	// Verify balance unchanged
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(2000), c.BalanceCents)
}

func TestWithdraw_ExactBalance(t *testing.T) {
	db, _, parent, child, tr := setupTransactionTest(t)

	// Deposit $50
	_, _, err := tr.Deposit(child.ID, parent.ID, 5000, "")
	require.NoError(t, err)

	// Withdraw exact balance (should succeed)
	_, newBalance, err := tr.Withdraw(child.ID, parent.ID, 5000, "Withdraw all")
	require.NoError(t, err)
	assert.Equal(t, int64(0), newBalance)

	// Verify balance is 0
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(0), c.BalanceCents)
}

func TestWithdraw_InvalidAmount(t *testing.T) {
	_, _, parent, child, tr := setupTransactionTest(t)

	// Zero amount
	_, _, err := tr.Withdraw(child.ID, parent.ID, 0, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")

	// Negative amount
	_, _, err = tr.Withdraw(child.ID, parent.ID, -100, "")
	assert.Error(t, err)
}

// T007: Tests for TransactionRepo.ListByChild()
func TestListByChild(t *testing.T) {
	_, _, parent, child, tr := setupTransactionTest(t)

	// Create several transactions
	_, _, _ = tr.Deposit(child.ID, parent.ID, 1000, "First")
	_, _, _ = tr.Deposit(child.ID, parent.ID, 500, "Second")
	_, _, _ = tr.Withdraw(child.ID, parent.ID, 200, "Third")

	// List transactions
	transactions, err := tr.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, transactions, 3)

	// Check that all transactions are present
	amounts := make(map[int64]bool)
	for _, tx := range transactions {
		amounts[tx.AmountCents] = true
	}
	assert.True(t, amounts[1000], "Should contain 1000 cent transaction")
	assert.True(t, amounts[500], "Should contain 500 cent transaction")
	assert.True(t, amounts[200], "Should contain 200 cent transaction")
}

func TestListByChild_Empty(t *testing.T) {
	_, _, _, child, tr := setupTransactionTest(t)

	transactions, err := tr.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Empty(t, transactions)
}

func TestListByChild_OnlyOwnTransactions(t *testing.T) {
	db := testDB(t)
	tr := NewTransactionRepo(db)

	fam := &models.Family{Slug: "test-family"}
	require.NoError(t, db.Create(fam).Error)

	parent := &models.Parent{
		GoogleID:    "google-test-123",
		Email:       "test@example.com",
		DisplayName: "Test Parent",
		FamilyID:    fam.ID,
	}
	require.NoError(t, db.Create(parent).Error)

	child1 := &models.Child{
		FamilyID:     fam.ID,
		FirstName:    "Child1",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, db.Create(child1).Error)

	child2 := &models.Child{
		FamilyID:     fam.ID,
		FirstName:    "Child2",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, db.Create(child2).Error)

	// Deposit to both
	_, _, _ = tr.Deposit(child1.ID, parent.ID, 1000, "For Child1")
	_, _, _ = tr.Deposit(child2.ID, parent.ID, 2000, "For Child2")

	// Each child should only see their own transactions
	tx1, err := tr.ListByChild(child1.ID)
	require.NoError(t, err)
	assert.Len(t, tx1, 1)
	assert.Equal(t, int64(1000), tx1[0].AmountCents)

	tx2, err := tr.ListByChild(child2.ID)
	require.NoError(t, err)
	assert.Len(t, tx2, 1)
	assert.Equal(t, int64(2000), tx2[0].AmountCents)
}

func TestGetByID_Transaction(t *testing.T) {
	_, _, parent, child, tr := setupTransactionTest(t)

	// Create a transaction
	created, _, err := tr.Deposit(child.ID, parent.ID, 1000, "Test note")
	require.NoError(t, err)

	// Fetch by ID
	fetched, err := tr.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, int64(1000), fetched.AmountCents)
	assert.Equal(t, models.TransactionTypeDeposit, fetched.TransactionType)
	assert.NotNil(t, fetched.Note)
	assert.Equal(t, "Test note", *fetched.Note)
}

func TestGetByID_Transaction_NotFound(t *testing.T) {
	db := testDB(t)
	tr := NewTransactionRepo(db)

	fetched, err := tr.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestDepositAllowance(t *testing.T) {
	db, _, parent, child, tr := setupTransactionTest(t)

	// Create an allowance schedule first (needed for FK)
	dow := 1
	schedule := &models.AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 500,
		Frequency:   models.FrequencyWeekly,
		DayOfWeek:   &dow,
	}
	require.NoError(t, db.Create(schedule).Error)

	tx, newBalance, err := tr.DepositAllowance(child.ID, parent.ID, 500, schedule.ID, "Weekly allowance")
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, int64(500), tx.AmountCents)
	assert.Equal(t, models.TransactionTypeAllowance, tx.TransactionType)
	assert.NotNil(t, tx.ScheduleID)
	assert.Equal(t, schedule.ID, *tx.ScheduleID)
	assert.NotNil(t, tx.Note)
	assert.Equal(t, "Weekly allowance", *tx.Note)
	assert.Equal(t, int64(500), newBalance)

	// Verify balance updated
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(500), c.BalanceCents)
}

func TestDepositAllowance_InvalidAmount(t *testing.T) {
	_, _, parent, child, tr := setupTransactionTest(t)

	_, _, err := tr.DepositAllowance(child.ID, parent.ID, 0, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

func TestListByChildPaginated(t *testing.T) {
	_, _, parent, child, tr := setupTransactionTest(t)

	// Create 5 transactions
	for i := 1; i <= 5; i++ {
		_, _, _ = tr.Deposit(child.ID, parent.ID, int64(i*100), "")
	}

	// Get first page (limit 2, offset 0)
	page1, err := tr.ListByChildPaginated(child.ID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page (limit 2, offset 2)
	page2, err := tr.ListByChildPaginated(child.ID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)

	// Get third page (limit 2, offset 4)
	page3, err := tr.ListByChildPaginated(child.ID, 2, 4)
	require.NoError(t, err)
	assert.Len(t, page3, 1)
}
