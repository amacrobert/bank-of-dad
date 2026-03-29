package repositories

import (
	"testing"
	"time"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupInterestTest(t *testing.T) (*gorm.DB, *models.Parent, *models.Child, *InterestRepo, *TransactionRepo) {
	t.Helper()
	db := testDB(t)
	ir := NewInterestRepo(db)
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

	return db, parent, child, ir, tr
}

// T003: Tests for SetInterestRate

func TestSetInterestRate(t *testing.T) {
	db, _, child, ir, _ := setupInterestTest(t)

	// Set interest rate
	err := ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Verify rate was stored
	var rateBps int
	require.NoError(t, db.Raw("SELECT interest_rate_bps FROM children WHERE id = ?", child.ID).Scan(&rateBps).Error)
	assert.Equal(t, 500, rateBps)
}

func TestSetInterestRate_Update(t *testing.T) {
	db, _, child, ir, _ := setupInterestTest(t)

	// Set initial rate
	err := ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Update to new rate
	err = ir.SetInterestRate(child.ID, 1000)
	require.NoError(t, err)

	var rateBps int
	require.NoError(t, db.Raw("SELECT interest_rate_bps FROM children WHERE id = ?", child.ID).Scan(&rateBps).Error)
	assert.Equal(t, 1000, rateBps)
}

func TestSetInterestRate_SetToZero(t *testing.T) {
	db, _, child, ir, _ := setupInterestTest(t)

	// Set rate then disable
	err := ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	err = ir.SetInterestRate(child.ID, 0)
	require.NoError(t, err)

	var rateBps int
	require.NoError(t, db.Raw("SELECT interest_rate_bps FROM children WHERE id = ?", child.ID).Scan(&rateBps).Error)
	assert.Equal(t, 0, rateBps)
}

func TestSetInterestRate_ValidationBounds(t *testing.T) {
	_, _, child, ir, _ := setupInterestTest(t)

	// Negative rate
	err := ir.SetInterestRate(child.ID, -1)
	assert.Error(t, err)

	// Rate above 10000 (100%)
	err = ir.SetInterestRate(child.ID, 10001)
	assert.Error(t, err)

	// Boundary values should succeed
	err = ir.SetInterestRate(child.ID, 0)
	assert.NoError(t, err)

	err = ir.SetInterestRate(child.ID, 10000)
	assert.NoError(t, err)
}

func TestGetInterestRate(t *testing.T) {
	_, _, child, ir, _ := setupInterestTest(t)

	// Default rate is 0
	rate, err := ir.GetInterestRate(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, rate)

	// Set and read back
	err = ir.SetInterestRate(child.ID, 750)
	require.NoError(t, err)

	rate, err = ir.GetInterestRate(child.ID)
	require.NoError(t, err)
	assert.Equal(t, 750, rate)
}

// T004: Tests for ApplyInterest

func TestApplyInterest(t *testing.T) {
	db, parent, child, ir, tr := setupInterestTest(t)

	// Set up: deposit $200.00 and set 10% interest rate
	_, _, err := tr.Deposit(child.ID, parent.ID, 20000, "Initial deposit")
	require.NoError(t, err)
	err = ir.SetInterestRate(child.ID, 1000) // 10%
	require.NoError(t, err)

	// Apply interest with monthly proration (12): 20000 * 1000 / 12 / 10000 = 166.67 → 167 cents
	err = ir.ApplyInterest(child.ID, parent.ID, 1000, models.FrequencyMonthly)
	require.NoError(t, err)

	// Verify balance increased
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, int64(20167), c.BalanceCents) // 20000 + 167

	// Verify interest transaction was created
	txns, err := tr.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, txns, 2) // deposit + interest

	interestTx := txns[0] // most recent first
	assert.Equal(t, models.TransactionTypeInterest, interestTx.TransactionType)
	assert.Equal(t, int64(167), interestTx.AmountCents)
	assert.Equal(t, parent.ID, interestTx.ParentID)
	require.NotNil(t, interestTx.Note)
	assert.Contains(t, *interestTx.Note, "10%")
	assert.Contains(t, *interestTx.Note, "compounded monthly")
}

func TestApplyInterest_UpdatesLastInterestAt(t *testing.T) {
	db, parent, child, ir, tr := setupInterestTest(t)

	_, _, err := tr.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Before applying, last_interest_at should be null
	var lastInterest *string
	require.NoError(t, db.Raw("SELECT last_interest_at FROM children WHERE id = ?", child.ID).Scan(&lastInterest).Error)
	assert.Nil(t, lastInterest)

	// Apply interest
	err = ir.ApplyInterest(child.ID, parent.ID, 500, models.FrequencyMonthly)
	require.NoError(t, err)

	// After applying, last_interest_at should be set
	var lastInterestStr string
	require.NoError(t, db.Raw("SELECT last_interest_at FROM children WHERE id = ?", child.ID).Scan(&lastInterestStr).Error)
	assert.NotEmpty(t, lastInterestStr)
}

func TestApplyInterest_CalculationExamples(t *testing.T) {
	tests := []struct {
		name         string
		balanceCents int64
		rateBps      int
		frequency    models.Frequency
		wantInterest int64
	}{
		// Monthly (12 periods): balance * rate / 12 / 10000
		{"$100 at 5% monthly", 10000, 500, models.FrequencyMonthly, 42},            // 10000 * 500 / 12 / 10000 = 41.67 → 42
		{"$100 at 10% monthly", 10000, 1000, models.FrequencyMonthly, 83},           // 10000 * 1000 / 12 / 10000 = 83.33 → 83
		{"$1000 at 5% monthly", 100000, 500, models.FrequencyMonthly, 417},          // 100000 * 500 / 12 / 10000 = 416.67 → 417
		{"$50 at 12% monthly", 5000, 1200, models.FrequencyMonthly, 50},             // 5000 * 1200 / 12 / 10000 = 50
		{"$1 at 5% monthly", 100, 500, models.FrequencyMonthly, 0},                  // 100 * 500 / 12 / 10000 = 0.42 → 0 (skip)
		// Weekly (52 periods): balance * rate / 52 / 10000
		{"$1000 at 5% weekly", 100000, 500, models.FrequencyWeekly, 96},             // 100000 * 500 / 52 / 10000 = 96.15 → 96
		{"$200 at 10% weekly", 20000, 1000, models.FrequencyWeekly, 38},             // 20000 * 1000 / 52 / 10000 = 38.46 → 38
		// Biweekly (26 periods): balance * rate / 26 / 10000
		{"$1000 at 5% biweekly", 100000, 500, models.FrequencyBiweekly, 192},       // 100000 * 500 / 26 / 10000 = 192.31 → 192
		{"$200 at 10% biweekly", 20000, 1000, models.FrequencyBiweekly, 77},        // 20000 * 1000 / 26 / 10000 = 76.92 → 77
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, parent, child, ir, tr := setupInterestTest(t)

			if tt.balanceCents > 0 {
				_, _, err := tr.Deposit(child.ID, parent.ID, tt.balanceCents, "")
				require.NoError(t, err)
			}
			err := ir.SetInterestRate(child.ID, tt.rateBps)
			require.NoError(t, err)

			err = ir.ApplyInterest(child.ID, parent.ID, tt.rateBps, tt.frequency)

			if tt.wantInterest == 0 {
				assert.Error(t, err, "should return error for zero interest")
				// Balance unchanged
				var c models.Child
				require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
				assert.Equal(t, tt.balanceCents, c.BalanceCents)
			} else {
				require.NoError(t, err)
				var c models.Child
				require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
				assert.Equal(t, tt.balanceCents+tt.wantInterest, c.BalanceCents)
			}
		})
	}
}

// T005: Edge case tests

func TestApplyInterest_ZeroBalance(t *testing.T) {
	_, parent, child, ir, _ := setupInterestTest(t)

	err := ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Should fail — zero balance means zero interest
	err = ir.ApplyInterest(child.ID, parent.ID, 500, models.FrequencyMonthly)
	assert.Error(t, err)
}

func TestApplyInterest_ZeroRate(t *testing.T) {
	_, parent, child, ir, tr := setupInterestTest(t)

	_, _, err := tr.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)

	// Should fail — zero rate means zero interest
	err = ir.ApplyInterest(child.ID, parent.ID, 0, models.FrequencyMonthly)
	assert.Error(t, err)
}

func TestApplyInterest_DuplicatePrevention(t *testing.T) {
	db, parent, child, ir, tr := setupInterestTest(t)

	_, _, err := tr.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// First accrual should succeed
	err = ir.ApplyInterest(child.ID, parent.ID, 500, models.FrequencyMonthly)
	require.NoError(t, err)

	var balanceAfterFirst int64
	var c models.Child
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	balanceAfterFirst = c.BalanceCents

	// Second accrual in the same month should be skipped by ListDueForInterest
	dues, err := ir.ListDueForInterest()
	require.NoError(t, err)
	// The child should NOT be in the due list after accrual this month
	for _, d := range dues {
		assert.NotEqual(t, child.ID, d.ChildID, "child should not be due for interest again this month")
	}

	// Balance should not have changed
	require.NoError(t, db.Select("balance_cents").First(&c, child.ID).Error)
	assert.Equal(t, balanceAfterFirst, c.BalanceCents)
}

// T006: Tests for ListDueForInterest

func TestListDueForInterest(t *testing.T) {
	db := testDB(t)
	ir := NewInterestRepo(db)
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

	child1 := &models.Child{FamilyID: fam.ID, FirstName: "Child1", PasswordHash: "hashed_password"}
	require.NoError(t, db.Create(child1).Error)
	child2 := &models.Child{FamilyID: fam.ID, FirstName: "Child2", PasswordHash: "hashed_password"}
	require.NoError(t, db.Create(child2).Error)
	child3 := &models.Child{FamilyID: fam.ID, FirstName: "Child3", PasswordHash: "hashed_password"}
	require.NoError(t, db.Create(child3).Error)

	// Child1: has rate and balance → should be due
	_, _, err := tr.Deposit(child1.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = ir.SetInterestRate(child1.ID, 500)
	require.NoError(t, err)

	// Child2: has rate but no balance → should NOT be due
	err = ir.SetInterestRate(child2.ID, 500)
	require.NoError(t, err)

	// Child3: has balance but no rate → should NOT be due
	_, _, err = tr.Deposit(child3.ID, parent.ID, 10000, "")
	require.NoError(t, err)

	dues, err := ir.ListDueForInterest()
	require.NoError(t, err)
	assert.Len(t, dues, 1)
	assert.Equal(t, child1.ID, dues[0].ChildID)
	assert.Equal(t, int64(10000), dues[0].BalanceCents)
	assert.Equal(t, 500, dues[0].InterestRateBps)
}

func TestListDueForInterest_AlreadyAccruedThisMonth(t *testing.T) {
	_, parent, child, ir, tr := setupInterestTest(t)

	_, _, err := tr.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Apply interest (sets last_interest_at to now)
	err = ir.ApplyInterest(child.ID, parent.ID, 500, models.FrequencyMonthly)
	require.NoError(t, err)

	// Should not be due anymore
	dues, err := ir.ListDueForInterest()
	require.NoError(t, err)
	assert.Len(t, dues, 0)
}

func TestListDueForInterest_LastAccruedPreviousMonth(t *testing.T) {
	db, parent, child, ir, tr := setupInterestTest(t)

	_, _, err := tr.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = ir.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	// Manually set last_interest_at to previous month
	// Use first day of current month minus 1 day to guarantee landing in previous month
	// (AddDate(0, -1, 0) can overflow back into the current month on the 29th-31st)
	now := time.Now()
	prevMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1).Format(time.RFC3339)
	require.NoError(t, db.Exec("UPDATE children SET last_interest_at = ? WHERE id = ?", prevMonth, child.ID).Error)

	// Should be due
	dues, err := ir.ListDueForInterest()
	require.NoError(t, err)
	assert.Len(t, dues, 1)
	assert.Equal(t, child.ID, dues[0].ChildID)
}
