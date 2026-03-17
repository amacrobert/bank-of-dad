package repositories

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// InterestDue represents a child eligible for interest accrual.
type InterestDue struct {
	ChildID         int64
	ParentID        int64
	BalanceCents    int64
	InterestRateBps int
}

// InterestRepo handles database operations for interest accrual using GORM.
type InterestRepo struct {
	db *gorm.DB
}

// NewInterestRepo creates a new InterestRepo.
func NewInterestRepo(db *gorm.DB) *InterestRepo {
	return &InterestRepo{db: db}
}

// SetInterestRate sets the annual interest rate in basis points for a child.
// Rate must be between 0 and 10000 (0% to 100%).
func (r *InterestRepo) SetInterestRate(childID int64, rateBps int) error {
	if rateBps < 0 || rateBps > 10000 {
		return fmt.Errorf("interest rate must be between 0 and 10000 basis points")
	}

	result := r.db.Model(&models.Child{}).Where("id = ?", childID).Updates(map[string]interface{}{
		"interest_rate_bps": rateBps,
		"updated_at":        time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("set interest rate: %w", result.Error)
	}
	return nil
}

// GetInterestRate returns the current interest rate in basis points for a child.
func (r *InterestRepo) GetInterestRate(childID int64) (int, error) {
	var child models.Child
	err := r.db.Select("interest_rate_bps").First(&child, childID).Error
	if err == gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("child not found")
	}
	if err != nil {
		return 0, fmt.Errorf("get interest rate: %w", err)
	}
	return child.InterestRateBps, nil
}

// ListDueForInterest returns children eligible for interest accrual:
// - interest_rate_bps > 0
// - balance_cents > 0
// - is_disabled = false
// - last_interest_at is NULL or not in the current calendar month
func (r *InterestRepo) ListDueForInterest() ([]InterestDue, error) {
	now := time.Now()
	currentMonth := now.Format("2006-01")

	var dues []InterestDue
	err := r.db.Raw(`
		SELECT c.id AS child_id, c.balance_cents, c.interest_rate_bps, p.id AS parent_id
		FROM children c
		JOIN parents p ON p.family_id = c.family_id
		WHERE c.interest_rate_bps > 0
		  AND c.balance_cents > 0
		  AND c.is_disabled = FALSE
		  AND (c.last_interest_at IS NULL OR to_char(c.last_interest_at, 'YYYY-MM') != ?)
	`, currentMonth).Scan(&dues).Error
	if err != nil {
		return nil, fmt.Errorf("list due for interest: %w", err)
	}
	return dues, nil
}

// ApplyInterest atomically calculates interest, creates a transaction, updates the balance,
// and sets last_interest_at. frequency controls proration: monthly=12, biweekly=26, weekly=52 periods per year.
// Returns an error if the calculated interest rounds to zero.
func (r *InterestRepo) ApplyInterest(childID, parentID int64, rateBps int, frequency models.Frequency) error {
	// Get current balance
	var child models.Child
	err := r.db.Select("balance_cents").First(&child, childID).Error
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	balanceCents := child.BalanceCents

	if balanceCents <= 0 {
		return fmt.Errorf("no interest on zero or negative balance")
	}
	if rateBps <= 0 {
		return fmt.Errorf("no interest with zero rate")
	}

	var periodsPerYear int
	switch frequency {
	case models.FrequencyWeekly:
		periodsPerYear = 52
	case models.FrequencyBiweekly:
		periodsPerYear = 26
	case models.FrequencyMonthly:
		periodsPerYear = 12
	default:
		periodsPerYear = 12
	}

	// Calculate interest: balance_cents * rate_bps / periodsPerYear / 10000
	interestFloat := float64(balanceCents) * float64(rateBps) / float64(periodsPerYear) / 10000.0
	interestCents := int64(math.Round(interestFloat))

	if interestCents <= 0 {
		return fmt.Errorf("calculated interest rounds to zero")
	}

	// Format rate for note (no trailing zeros: 500bps->"5", 525bps->"5.25")
	ratePercent := strconv.FormatFloat(float64(rateBps)/100.0, 'f', -1, 64)
	note := ratePercent + "% annual interest compounded " + string(frequency)

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Insert interest transaction
		transaction := models.Transaction{
			ChildID:         childID,
			ParentID:        parentID,
			AmountCents:     interestCents,
			TransactionType: models.TransactionTypeInterest,
			Note:            &note,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("insert interest transaction: %w", err)
		}

		// Update balance and last_interest_at
		now := time.Now().UTC()
		if err := tx.Exec(
			`UPDATE children SET balance_cents = balance_cents + ?, last_interest_at = ?, updated_at = ? WHERE id = ?`,
			interestCents, now, now, childID,
		).Error; err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		return nil
	})
}
