package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T003: Tests for SetInterestRate

func TestSetInterestRate(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)

	fam := createTestFamily(t, db)
	child := createTestChild(t, db, fam.ID)

	// Set interest rate
	err := is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Verify rate was stored
	var rateBps int
	err = db.Read.QueryRow("SELECT interest_rate_bps FROM children WHERE id = ?", child.ID).Scan(&rateBps)
	require.NoError(t, err)
	assert.Equal(t, 500, rateBps)
}

func TestSetInterestRate_Update(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)

	fam := createTestFamily(t, db)
	child := createTestChild(t, db, fam.ID)

	// Set initial rate
	err := is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Update to new rate
	err = is.SetInterestRate(child.ID, 1000)
	require.NoError(t, err)

	var rateBps int
	err = db.Read.QueryRow("SELECT interest_rate_bps FROM children WHERE id = ?", child.ID).Scan(&rateBps)
	require.NoError(t, err)
	assert.Equal(t, 1000, rateBps)
}

func TestSetInterestRate_SetToZero(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)

	fam := createTestFamily(t, db)
	child := createTestChild(t, db, fam.ID)

	// Set rate then disable
	err := is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	err = is.SetInterestRate(child.ID, 0)
	require.NoError(t, err)

	var rateBps int
	err = db.Read.QueryRow("SELECT interest_rate_bps FROM children WHERE id = ?", child.ID).Scan(&rateBps)
	require.NoError(t, err)
	assert.Equal(t, 0, rateBps)
}

func TestSetInterestRate_ValidationBounds(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)

	fam := createTestFamily(t, db)
	child := createTestChild(t, db, fam.ID)

	// Negative rate
	err := is.SetInterestRate(child.ID, -1)
	assert.Error(t, err)

	// Rate above 10000 (100%)
	err = is.SetInterestRate(child.ID, 10001)
	assert.Error(t, err)

	// Boundary values should succeed
	err = is.SetInterestRate(child.ID, 0)
	assert.NoError(t, err)

	err = is.SetInterestRate(child.ID, 10000)
	assert.NoError(t, err)
}

func TestGetInterestRate(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)

	fam := createTestFamily(t, db)
	child := createTestChild(t, db, fam.ID)

	// Default rate is 0
	rate, err := is.GetInterestRate(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, rate)

	// Set and read back
	err = is.SetInterestRate(child.ID, 750)
	require.NoError(t, err)

	rate, err = is.GetInterestRate(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 750, rate)
}

// T004: Tests for ApplyInterest

func TestApplyInterest(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)
	ts := NewTransactionStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	// Set up: deposit $200.00 and set 10% interest rate
	_, _, err := ts.Deposit(child.ID, parent.ID, 20000, "Initial deposit")
	require.NoError(t, err)
	err = is.SetInterestRate(child.ID, 1000) // 10%
	require.NoError(t, err)

	// Apply interest with monthly proration (12): 20000 * 1000 / 12 / 10000 = 166.67 → 167 cents
	err = is.ApplyInterest(child.ID, parent.ID, 1000, 12)
	require.NoError(t, err)

	// Verify balance increased
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(20167), balance) // 20000 + 167

	// Verify interest transaction was created
	txns, err := ts.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, txns, 2) // deposit + interest

	interestTx := txns[0] // most recent first
	assert.Equal(t, TransactionTypeInterest, interestTx.TransactionType)
	assert.Equal(t, int64(167), interestTx.AmountCents)
	assert.Equal(t, parent.ID, interestTx.ParentID)
	require.NotNil(t, interestTx.Note)
	assert.Contains(t, *interestTx.Note, "10.00%")
}

func TestApplyInterest_UpdatesLastInterestAt(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	_, _, err := ts.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Before applying, last_interest_at should be null
	var lastInterest *string
	err = db.Read.QueryRow("SELECT last_interest_at FROM children WHERE id = ?", child.ID).Scan(&lastInterest)
	require.NoError(t, err)
	assert.Nil(t, lastInterest)

	// Apply interest
	err = is.ApplyInterest(child.ID, parent.ID, 500, 12)
	require.NoError(t, err)

	// After applying, last_interest_at should be set
	var lastInterestStr string
	err = db.Read.QueryRow("SELECT last_interest_at FROM children WHERE id = ?", child.ID).Scan(&lastInterestStr)
	require.NoError(t, err)
	assert.NotEmpty(t, lastInterestStr)
}

func TestApplyInterest_CalculationExamples(t *testing.T) {
	tests := []struct {
		name           string
		balanceCents   int64
		rateBps        int
		periodsPerYear int
		wantInterest   int64
	}{
		// Monthly (12 periods): balance * rate / 12 / 10000
		{"$100 at 5% monthly", 10000, 500, 12, 42},           // 10000 * 500 / 12 / 10000 = 41.67 → 42
		{"$100 at 10% monthly", 10000, 1000, 12, 83},          // 10000 * 1000 / 12 / 10000 = 83.33 → 83
		{"$1000 at 5% monthly", 100000, 500, 12, 417},         // 100000 * 500 / 12 / 10000 = 416.67 → 417
		{"$50 at 12% monthly", 5000, 1200, 12, 50},            // 5000 * 1200 / 12 / 10000 = 50
		{"$1 at 5% monthly", 100, 500, 12, 0},                 // 100 * 500 / 12 / 10000 = 0.42 → 0 (skip)
		// Weekly (52 periods): balance * rate / 52 / 10000
		{"$1000 at 5% weekly", 100000, 500, 52, 96},           // 100000 * 500 / 52 / 10000 = 96.15 → 96
		{"$200 at 10% weekly", 20000, 1000, 52, 38},           // 20000 * 1000 / 52 / 10000 = 38.46 → 38
		// Biweekly (26 periods): balance * rate / 26 / 10000
		{"$1000 at 5% biweekly", 100000, 500, 26, 192},        // 100000 * 500 / 26 / 10000 = 192.31 → 192
		{"$200 at 10% biweekly", 20000, 1000, 26, 77},         // 20000 * 1000 / 26 / 10000 = 76.92 → 77
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testDB(t)
			is := NewInterestStore(db)
			ts := NewTransactionStore(db)
			cs := NewChildStore(db)

			fam := createTestFamily(t, db)
			parent := createTestParent(t, db, fam.ID)
			child := createTestChild(t, db, fam.ID)

			if tt.balanceCents > 0 {
				_, _, err := ts.Deposit(child.ID, parent.ID, tt.balanceCents, "")
				require.NoError(t, err)
			}
			err := is.SetInterestRate(child.ID, tt.rateBps)
			require.NoError(t, err)

			err = is.ApplyInterest(child.ID, parent.ID, tt.rateBps, tt.periodsPerYear)

			if tt.wantInterest == 0 {
				assert.Error(t, err, "should return error for zero interest")
				// Balance unchanged
				balance, err := cs.GetBalance(child.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.balanceCents, balance)
			} else {
				require.NoError(t, err)
				balance, err := cs.GetBalance(child.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.balanceCents+tt.wantInterest, balance)
			}
		})
	}
}

// T005: Edge case tests

func TestApplyInterest_ZeroBalance(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	err := is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Should fail — zero balance means zero interest
	err = is.ApplyInterest(child.ID, parent.ID, 500, 12)
	assert.Error(t, err)
}

func TestApplyInterest_ZeroRate(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	_, _, err := ts.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)

	// Should fail — zero rate means zero interest
	err = is.ApplyInterest(child.ID, parent.ID, 0, 12)
	assert.Error(t, err)
}

func TestApplyInterest_DuplicatePrevention(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)
	ts := NewTransactionStore(db)
	cs := NewChildStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	_, _, err := ts.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// First accrual should succeed
	err = is.ApplyInterest(child.ID, parent.ID, 500, 12)
	require.NoError(t, err)

	balanceAfterFirst, err := cs.GetBalance(child.ID)
	require.NoError(t, err)

	// Second accrual in the same month should be skipped by ListDueForInterest
	// But if called directly, ApplyInterest updates last_interest_at
	// so ListDueForInterest won't return it again
	dues, err := is.ListDueForInterest()
	require.NoError(t, err)
	// The child should NOT be in the due list after accrual this month
	for _, d := range dues {
		assert.NotEqual(t, child.ID, d.ChildID, "child should not be due for interest again this month")
	}

	// Balance should not have changed from the second attempt's perspective
	balance, err := cs.GetBalance(child.ID)
	require.NoError(t, err)
	assert.Equal(t, balanceAfterFirst, balance)
}

// T006: Tests for ListDueForInterest

func TestListDueForInterest(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)

	cs := NewChildStore(db)
	child1, err := cs.Create(fam.ID, "Child1", "pass123")
	require.NoError(t, err)
	child2, err := cs.Create(fam.ID, "Child2", "pass123")
	require.NoError(t, err)
	child3, err := cs.Create(fam.ID, "Child3", "pass123")
	require.NoError(t, err)

	// Child1: has rate and balance → should be due
	_, _, err = ts.Deposit(child1.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = is.SetInterestRate(child1.ID, 500)
	require.NoError(t, err)

	// Child2: has rate but no balance → should NOT be due
	err = is.SetInterestRate(child2.ID, 500)
	require.NoError(t, err)

	// Child3: has balance but no rate → should NOT be due
	_, _, err = ts.Deposit(child3.ID, parent.ID, 10000, "")
	require.NoError(t, err)

	dues, err := is.ListDueForInterest()
	require.NoError(t, err)
	assert.Len(t, dues, 1)
	assert.Equal(t, child1.ID, dues[0].ChildID)
	assert.Equal(t, int64(10000), dues[0].BalanceCents)
	assert.Equal(t, 500, dues[0].InterestRateBps)
}

func TestListDueForInterest_AlreadyAccruedThisMonth(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	_, _, err := ts.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Apply interest (sets last_interest_at to now)
	err = is.ApplyInterest(child.ID, parent.ID, 500, 12)
	require.NoError(t, err)

	// Should not be due anymore
	dues, err := is.ListDueForInterest()
	require.NoError(t, err)
	assert.Len(t, dues, 0)
}

func TestListDueForInterest_LastAccruedPreviousMonth(t *testing.T) {
	db := testDB(t)
	is := NewInterestStore(db)
	ts := NewTransactionStore(db)

	fam := createTestFamily(t, db)
	parent := createTestParent(t, db, fam.ID)
	child := createTestChild(t, db, fam.ID)

	_, _, err := ts.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = is.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Manually set last_interest_at to previous month
	prevMonth := time.Now().AddDate(0, -1, 0).Format(time.RFC3339)
	_, err = db.Write.Exec("UPDATE children SET last_interest_at = ? WHERE id = ?", prevMonth, child.ID)
	require.NoError(t, err)

	// Should be due
	dues, err := is.ListDueForInterest()
	require.NoError(t, err)
	assert.Len(t, dues, 1)
	assert.Equal(t, child.ID, dues[0].ChildID)
}
