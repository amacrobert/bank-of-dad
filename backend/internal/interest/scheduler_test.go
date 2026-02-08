package interest

import (
	"testing"
	"time"

	"bank-of-dad/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T017: Tests for ProcessDue

func TestProcessDue_AppliesInterest(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	interestStore := store.NewInterestStore(db)
	txStore := store.NewTransactionStore(db)
	cs := store.NewChildStore(db)

	// Set up: deposit $100 and set 10% rate
	_, _, err := txStore.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = interestStore.SetInterestRate(child.ID, 1000)
	require.NoError(t, err)

	scheduler := NewScheduler(interestStore)
	scheduler.ProcessDue()

	// Verify balance increased: 10000 * 1000 / 12 / 10000 = 83.33 → 83
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10083), balance)

	// Verify interest transaction was created
	txns, err := txStore.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, txns, 2) // deposit + interest
	assert.Equal(t, store.TransactionTypeInterest, txns[0].TransactionType)
}

func TestProcessDue_SkipsAlreadyAccruedThisMonth(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	interestStore := store.NewInterestStore(db)
	txStore := store.NewTransactionStore(db)
	cs := store.NewChildStore(db)

	_, _, err := txStore.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = interestStore.SetInterestRate(child.ID, 1000)
	require.NoError(t, err)

	scheduler := NewScheduler(interestStore)

	// First accrual
	scheduler.ProcessDue()
	balanceAfterFirst, err := cs.GetBalance(child.ID)
	require.NoError(t, err)

	// Second accrual — should be skipped (same month)
	scheduler.ProcessDue()
	balanceAfterSecond, err := cs.GetBalance(child.ID)
	require.NoError(t, err)

	assert.Equal(t, balanceAfterFirst, balanceAfterSecond, "balance should not change on duplicate accrual")
}

func TestProcessDue_SkipsZeroBalance(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	interestStore := store.NewInterestStore(db)
	cs := store.NewChildStore(db)

	// Set rate but no balance
	err := interestStore.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	scheduler := NewScheduler(interestStore)
	scheduler.ProcessDue()

	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), balance, "balance should remain zero")
}

func TestProcessDue_SkipsZeroRate(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	interestStore := store.NewInterestStore(db)
	txStore := store.NewTransactionStore(db)
	cs := store.NewChildStore(db)

	// Deposit but no rate set (default 0)
	_, _, err := txStore.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)

	scheduler := NewScheduler(interestStore)
	scheduler.ProcessDue()

	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000), balance, "balance should not change with zero rate")
}

func TestProcessDue_MultipleChildren(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)

	cs := store.NewChildStore(db)
	child1, err := cs.Create(family.ID, "Child1", "pass123")
	require.NoError(t, err)
	child2, err := cs.Create(family.ID, "Child2", "pass123")
	require.NoError(t, err)

	interestStore := store.NewInterestStore(db)
	txStore := store.NewTransactionStore(db)

	// Both children have balance and rate
	_, _, err = txStore.Deposit(child1.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	_, _, err = txStore.Deposit(child2.ID, parent.ID, 20000, "")
	require.NoError(t, err)
	err = interestStore.SetInterestRate(child1.ID, 500) // 5%
	require.NoError(t, err)
	err = interestStore.SetInterestRate(child2.ID, 1000) // 10%
	require.NoError(t, err)

	scheduler := NewScheduler(interestStore)
	scheduler.ProcessDue()

	// Child1: 10000 * 500 / 12 / 10000 = 41.67 → 42
	balance1, err := cs.GetBalance(child1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10042), balance1)

	// Child2: 20000 * 1000 / 12 / 10000 = 166.67 → 167
	balance2, err := cs.GetBalance(child2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(20167), balance2)
}

// T018: Test for partial failure

func TestProcessDue_PartialFailure(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)

	cs := store.NewChildStore(db)
	child1, err := cs.Create(family.ID, "Child1", "pass123")
	require.NoError(t, err)
	child2, err := cs.Create(family.ID, "Child2", "pass123")
	require.NoError(t, err)

	interestStore := store.NewInterestStore(db)
	txStore := store.NewTransactionStore(db)

	// Child1: $1 at 5% → interest rounds to 0, will be skipped (error from ApplyInterest)
	_, _, err = txStore.Deposit(child1.ID, parent.ID, 100, "")
	require.NoError(t, err)
	err = interestStore.SetInterestRate(child1.ID, 500)
	require.NoError(t, err)

	// Child2: $100 at 5% → 42 cents interest, should succeed
	_, _, err = txStore.Deposit(child2.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = interestStore.SetInterestRate(child2.ID, 500)
	require.NoError(t, err)

	scheduler := NewScheduler(interestStore)
	scheduler.ProcessDue()

	// Child1 should be unchanged (interest rounds to 0)
	balance1, err := cs.GetBalance(child1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100), balance1)

	// Child2 should have interest applied
	balance2, err := cs.GetBalance(child2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10042), balance2)
}

func TestScheduler_StartAndStop(t *testing.T) {
	db := setupTestDB(t)
	interestStore := store.NewInterestStore(db)

	scheduler := NewScheduler(interestStore)
	stop := make(chan struct{})

	scheduler.Start(100*time.Millisecond, stop)

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop should not hang
	close(stop)
	time.Sleep(50 * time.Millisecond) // Give goroutine time to exit
}
