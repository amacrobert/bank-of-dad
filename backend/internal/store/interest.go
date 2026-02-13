package store

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// InterestDue represents a child eligible for interest accrual.
type InterestDue struct {
	ChildID        int64
	ParentID       int64
	BalanceCents   int64
	InterestRateBps int
}

// InterestStore handles database operations for interest accrual.
type InterestStore struct {
	db *sql.DB
}

// NewInterestStore creates a new InterestStore.
func NewInterestStore(db *sql.DB) *InterestStore {
	return &InterestStore{db: db}
}

// SetInterestRate sets the annual interest rate in basis points for a child.
// Rate must be between 0 and 10000 (0% to 100%).
func (s *InterestStore) SetInterestRate(childID int64, rateBps int) error {
	if rateBps < 0 || rateBps > 10000 {
		return fmt.Errorf("interest rate must be between 0 and 10000 basis points")
	}

	_, err := s.db.Exec(
		`UPDATE children SET interest_rate_bps = $1, updated_at = NOW() WHERE id = $2`,
		rateBps, childID,
	)
	if err != nil {
		return fmt.Errorf("set interest rate: %w", err)
	}
	return nil
}

// GetInterestRate returns the current interest rate in basis points for a child.
func (s *InterestStore) GetInterestRate(childID int64) (int, error) {
	var rateBps int
	err := s.db.QueryRow(
		`SELECT interest_rate_bps FROM children WHERE id = $1`, childID,
	).Scan(&rateBps)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("child not found")
	}
	if err != nil {
		return 0, fmt.Errorf("get interest rate: %w", err)
	}
	return rateBps, nil
}

// ListDueForInterest returns children eligible for interest accrual:
// - interest_rate_bps > 0
// - balance_cents > 0
// - last_interest_at is NULL or not in the current calendar month
func (s *InterestStore) ListDueForInterest() ([]InterestDue, error) {
	now := time.Now()
	currentMonth := now.Format("2006-01")

	rows, err := s.db.Query(`
		SELECT c.id, c.balance_cents, c.interest_rate_bps, p.id
		FROM children c
		JOIN parents p ON p.family_id = c.family_id
		WHERE c.interest_rate_bps > 0
		  AND c.balance_cents > 0
		  AND (c.last_interest_at IS NULL OR to_char(c.last_interest_at, 'YYYY-MM') != $1)
	`, currentMonth)
	if err != nil {
		return nil, fmt.Errorf("list due for interest: %w", err)
	}
	defer rows.Close()

	var dues []InterestDue
	for rows.Next() {
		var d InterestDue
		if err := rows.Scan(&d.ChildID, &d.BalanceCents, &d.InterestRateBps, &d.ParentID); err != nil {
			return nil, fmt.Errorf("scan interest due: %w", err)
		}
		dues = append(dues, d)
	}
	return dues, rows.Err()
}

// ApplyInterest atomically calculates interest, creates a transaction, updates the balance,
// and sets last_interest_at. periodsPerYear controls proration: 12 for monthly, 26 for biweekly, 52 for weekly.
// Returns an error if the calculated interest rounds to zero.
func (s *InterestStore) ApplyInterest(childID, parentID int64, rateBps int, periodsPerYear int) error {
	// Get current balance
	var balanceCents int64
	err := s.db.QueryRow(`SELECT balance_cents FROM children WHERE id = $1`, childID).Scan(&balanceCents)
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	if balanceCents <= 0 {
		return fmt.Errorf("no interest on zero or negative balance")
	}
	if rateBps <= 0 {
		return fmt.Errorf("no interest with zero rate")
	}

	if periodsPerYear <= 0 {
		return fmt.Errorf("periodsPerYear must be positive")
	}

	// Calculate interest: balance_cents * rate_bps / periodsPerYear / 10000
	// Use float64 for the intermediate calculation to get proper rounding
	interestFloat := float64(balanceCents) * float64(rateBps) / float64(periodsPerYear) / 10000.0
	interestCents := int64(math.Round(interestFloat))

	if interestCents <= 0 {
		return fmt.Errorf("calculated interest rounds to zero")
	}

	// Format rate for note
	ratePercent := fmt.Sprintf("%.2f%%", float64(rateBps)/100.0)
	note := ratePercent + " annual rate"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert interest transaction
	_, err = tx.Exec(
		`INSERT INTO transactions (child_id, parent_id, amount_cents, transaction_type, note)
		 VALUES ($1, $2, $3, $4, $5)`,
		childID, parentID, interestCents, TransactionTypeInterest, note,
	)
	if err != nil {
		return fmt.Errorf("insert interest transaction: %w", err)
	}

	// Update balance
	_, err = tx.Exec(
		`UPDATE children SET balance_cents = balance_cents + $1, last_interest_at = $2, updated_at = NOW() WHERE id = $3`,
		interestCents, time.Now().UTC(), childID,
	)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
