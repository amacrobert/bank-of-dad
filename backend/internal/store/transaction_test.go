package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestParent(t *testing.T, db *DB, familyID int64) *Parent {
	t.Helper()
	ps := NewParentStore(db)
	p, err := ps.Create("google-test-123", "test@example.com", "Test Parent")
	require.NoError(t, err)
	// Associate with family
	err = ps.SetFamilyID(p.ID, familyID)
	require.NoError(t, err)
	p.FamilyID = familyID
	return p
}

func createTestChild(t *testing.T, db *DB, familyID int64) *Child {
	t.Helper()
	cs := NewChildStore(db)
	c, err := cs.Create(familyID, "TestChild", "password123", nil)
	require.NoError(t, err)
	return c
}

// T005: Tests for TransactionStore.Deposit()
func TestDeposit(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Verify initial balance is 0
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), balance)

	// Test deposit
	tx, newBalance, err := ts.Deposit(child.ID, parent.ID, 1000, "Weekly allowance")
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, int64(1000), tx.AmountCents)
	assert.Equal(t, TransactionTypeDeposit, tx.TransactionType)
	assert.NotNil(t, tx.Note)
	assert.Equal(t, "Weekly allowance", *tx.Note)
	assert.Equal(t, child.ID, tx.ChildID)
	assert.Equal(t, parent.ID, tx.ParentID)
	assert.Equal(t, int64(1000), newBalance)

	// Verify balance updated
	balance, err = cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), balance)
}

func TestDeposit_EmptyNote(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	tx, _, err := ts.Deposit(child.ID, parent.ID, 500, "")
	require.NoError(t, err)
	assert.Nil(t, tx.Note) // Empty note should be nil
}

func TestDeposit_InvalidAmount(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Zero amount
	_, _, err := ts.Deposit(child.ID, parent.ID, 0, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")

	// Negative amount
	_, _, err = ts.Deposit(child.ID, parent.ID, -100, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

func TestDeposit_MultipleDeposits(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// First deposit
	_, newBalance1, err := ts.Deposit(child.ID, parent.ID, 1000, "First")
	require.NoError(t, err)
	assert.Equal(t, int64(1000), newBalance1)

	// Second deposit
	_, newBalance2, err := ts.Deposit(child.ID, parent.ID, 500, "Second")
	require.NoError(t, err)
	assert.Equal(t, int64(1500), newBalance2)

	// Verify final balance
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1500), balance)
}

// T006: Tests for TransactionStore.Withdraw()
func TestWithdraw(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Deposit first
	_, _, err := ts.Deposit(child.ID, parent.ID, 5000, "Initial deposit")
	require.NoError(t, err)

	// Withdraw
	tx, newBalance, err := ts.Withdraw(child.ID, parent.ID, 1500, "Bought a book")
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, int64(1500), tx.AmountCents)
	assert.Equal(t, TransactionTypeWithdrawal, tx.TransactionType)
	assert.NotNil(t, tx.Note)
	assert.Equal(t, "Bought a book", *tx.Note)
	assert.Equal(t, int64(3500), newBalance)

	// Verify balance updated
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3500), balance)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Deposit $20
	_, _, err := ts.Deposit(child.ID, parent.ID, 2000, "")
	require.NoError(t, err)

	// Try to withdraw $25
	_, currentBalance, err := ts.Withdraw(child.ID, parent.ID, 2500, "")
	assert.ErrorIs(t, err, ErrInsufficientFunds)
	assert.Equal(t, int64(2000), currentBalance) // Returns current balance for error message

	// Verify balance unchanged
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), balance)
}

func TestWithdraw_ExactBalance(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Deposit $50
	_, _, err := ts.Deposit(child.ID, parent.ID, 5000, "")
	require.NoError(t, err)

	// Withdraw exact balance (should succeed)
	_, newBalance, err := ts.Withdraw(child.ID, parent.ID, 5000, "Withdraw all")
	require.NoError(t, err)
	assert.Equal(t, int64(0), newBalance)

	// Verify balance is 0
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), balance)
}

func TestWithdraw_InvalidAmount(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Zero amount
	_, _, err := ts.Withdraw(child.ID, parent.ID, 0, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")

	// Negative amount
	_, _, err = ts.Withdraw(child.ID, parent.ID, -100, "")
	assert.Error(t, err)
}

// T007: Tests for TransactionStore.ListByChild()
func TestListByChild(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Create several transactions
	ts.Deposit(child.ID, parent.ID, 1000, "First")
	ts.Deposit(child.ID, parent.ID, 500, "Second")
	ts.Withdraw(child.ID, parent.ID, 200, "Third")

	// List transactions
	transactions, err := ts.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, transactions, 3)

	// Check that all transactions are present (order may vary within same timestamp)
	amounts := make(map[int64]bool)
	for _, tx := range transactions {
		amounts[tx.AmountCents] = true
	}
	assert.True(t, amounts[1000], "Should contain 1000 cent transaction")
	assert.True(t, amounts[500], "Should contain 500 cent transaction")
	assert.True(t, amounts[200], "Should contain 200 cent transaction")
}

func TestListByChild_Empty(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	child := createTestChild(t, db, fam.ID)

	transactions, err := ts.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Empty(t, transactions)
}

func TestListByChild_OnlyOwnTransactions(t *testing.T) {
	db := testDB(t)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)

	// Create two children
	cs := NewChildStore(db)
	child1, _ := cs.Create(fam.ID, "Child1", "pass123", nil)
	child2, _ := cs.Create(fam.ID, "Child2", "pass123", nil)

	// Deposit to both
	ts.Deposit(child1.ID, parent.ID, 1000, "For Child1")
	ts.Deposit(child2.ID, parent.ID, 2000, "For Child2")

	// Each child should only see their own transactions
	tx1, err := ts.ListByChild(child1.ID)
	require.NoError(t, err)
	assert.Len(t, tx1, 1)
	assert.Equal(t, int64(1000), tx1[0].AmountCents)

	tx2, err := ts.ListByChild(child2.ID)
	require.NoError(t, err)
	assert.Len(t, tx2, 1)
	assert.Equal(t, int64(2000), tx2[0].AmountCents)
}
