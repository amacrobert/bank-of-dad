package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TransactionType represents the type of a balance transaction.
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeAllowance  TransactionType = "allowance"
)

// Transaction represents a record of money added or removed from a child's account.
type Transaction struct {
	ID              int64           `json:"id"`
	ChildID         int64           `json:"child_id"`
	ParentID        int64           `json:"parent_id"`
	AmountCents     int64           `json:"amount_cents"`
	TransactionType TransactionType `json:"type"`
	Note            *string         `json:"note,omitempty"`
	ScheduleID      *int64          `json:"schedule_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ErrInsufficientFunds is returned when a withdrawal exceeds the available balance.
var ErrInsufficientFunds = errors.New("insufficient funds")

// TransactionStore handles database operations for transactions.
type TransactionStore struct {
	db *DB
}

// NewTransactionStore creates a new TransactionStore.
func NewTransactionStore(db *DB) *TransactionStore {
	return &TransactionStore{db: db}
}

// Deposit adds money to a child's account and records the transaction.
// The operation is atomic - both the transaction record and balance update happen together.
func (s *TransactionStore) Deposit(childID, parentID, amountCents int64, note string) (*Transaction, int64, error) {
	if amountCents <= 0 {
		return nil, 0, fmt.Errorf("amount must be positive")
	}

	tx, err := s.db.Write.Begin()
	if err != nil {
		return nil, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert transaction record
	notePtr := nullableString(note)
	result, err := tx.Exec(
		`INSERT INTO transactions (child_id, parent_id, amount_cents, transaction_type, note)
		 VALUES (?, ?, ?, ?, ?)`,
		childID, parentID, amountCents, TransactionTypeDeposit, notePtr,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("insert transaction: %w", err)
	}

	txID, _ := result.LastInsertId()

	// Update balance
	_, err = tx.Exec(
		`UPDATE children SET balance_cents = balance_cents + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		amountCents, childID,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("update balance: %w", err)
	}

	// Get new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT balance_cents FROM children WHERE id = ?`, childID).Scan(&newBalance)
	if err != nil {
		return nil, 0, fmt.Errorf("get new balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit: %w", err)
	}

	// Fetch the created transaction
	transaction, err := s.GetByID(txID)
	if err != nil {
		return nil, 0, fmt.Errorf("get created transaction: %w", err)
	}

	return transaction, newBalance, nil
}

// Withdraw removes money from a child's account and records the transaction.
// Returns ErrInsufficientFunds if the withdrawal would result in a negative balance.
func (s *TransactionStore) Withdraw(childID, parentID, amountCents int64, note string) (*Transaction, int64, error) {
	if amountCents <= 0 {
		return nil, 0, fmt.Errorf("amount must be positive")
	}

	tx, err := s.db.Write.Begin()
	if err != nil {
		return nil, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check current balance
	var currentBalance int64
	err = tx.QueryRow(`SELECT balance_cents FROM children WHERE id = ?`, childID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return nil, 0, fmt.Errorf("child not found")
	}
	if err != nil {
		return nil, 0, fmt.Errorf("get current balance: %w", err)
	}

	if currentBalance < amountCents {
		return nil, currentBalance, ErrInsufficientFunds
	}

	// Insert transaction record
	notePtr := nullableString(note)
	result, err := tx.Exec(
		`INSERT INTO transactions (child_id, parent_id, amount_cents, transaction_type, note)
		 VALUES (?, ?, ?, ?, ?)`,
		childID, parentID, amountCents, TransactionTypeWithdrawal, notePtr,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("insert transaction: %w", err)
	}

	txID, _ := result.LastInsertId()

	// Update balance
	_, err = tx.Exec(
		`UPDATE children SET balance_cents = balance_cents - ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		amountCents, childID,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("update balance: %w", err)
	}

	// Get new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT balance_cents FROM children WHERE id = ?`, childID).Scan(&newBalance)
	if err != nil {
		return nil, 0, fmt.Errorf("get new balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit: %w", err)
	}

	// Fetch the created transaction
	transaction, err := s.GetByID(txID)
	if err != nil {
		return nil, 0, fmt.Errorf("get created transaction: %w", err)
	}

	return transaction, newBalance, nil
}

// GetByID retrieves a transaction by its ID.
func (s *TransactionStore) GetByID(id int64) (*Transaction, error) {
	var t Transaction
	var createdAt string
	var note sql.NullString
	var scheduleID sql.NullInt64

	err := s.db.Read.QueryRow(
		`SELECT id, child_id, parent_id, amount_cents, transaction_type, note, schedule_id, created_at
		 FROM transactions WHERE id = ?`, id,
	).Scan(&t.ID, &t.ChildID, &t.ParentID, &t.AmountCents, &t.TransactionType, &note, &scheduleID, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	t.CreatedAt, _ = parseTime(createdAt)
	if note.Valid {
		t.Note = &note.String
	}
	if scheduleID.Valid {
		t.ScheduleID = &scheduleID.Int64
	}

	return &t, nil
}

// ListByChild retrieves all transactions for a child, ordered by most recent first.
func (s *TransactionStore) ListByChild(childID int64) ([]Transaction, error) {
	rows, err := s.db.Read.Query(
		`SELECT id, child_id, parent_id, amount_cents, transaction_type, note, schedule_id, created_at
		 FROM transactions WHERE child_id = ? ORDER BY created_at DESC, id DESC`, childID,
	)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var createdAt string
		var note sql.NullString
		var scheduleID sql.NullInt64

		if err := rows.Scan(&t.ID, &t.ChildID, &t.ParentID, &t.AmountCents, &t.TransactionType, &note, &scheduleID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}

		t.CreatedAt, _ = parseTime(createdAt)
		if note.Valid {
			t.Note = &note.String
		}
		if scheduleID.Valid {
			t.ScheduleID = &scheduleID.Int64
		}

		transactions = append(transactions, t)
	}

	return transactions, rows.Err()
}

// DepositAllowance adds money to a child's account as a scheduled allowance transaction.
// Similar to Deposit but includes a schedule_id and uses "allowance" transaction type.
func (s *TransactionStore) DepositAllowance(childID, parentID, amountCents, scheduleID int64, note string) (*Transaction, int64, error) {
	if amountCents <= 0 {
		return nil, 0, fmt.Errorf("amount must be positive")
	}

	tx, err := s.db.Write.Begin()
	if err != nil {
		return nil, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	notePtr := nullableString(note)
	result, err := tx.Exec(
		`INSERT INTO transactions (child_id, parent_id, amount_cents, transaction_type, note, schedule_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		childID, parentID, amountCents, TransactionTypeAllowance, notePtr, scheduleID,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("insert transaction: %w", err)
	}

	txID, _ := result.LastInsertId()

	_, err = tx.Exec(
		`UPDATE children SET balance_cents = balance_cents + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		amountCents, childID,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("update balance: %w", err)
	}

	var newBalance int64
	err = tx.QueryRow(`SELECT balance_cents FROM children WHERE id = ?`, childID).Scan(&newBalance)
	if err != nil {
		return nil, 0, fmt.Errorf("get new balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit: %w", err)
	}

	transaction, err := s.GetByID(txID)
	if err != nil {
		return nil, 0, fmt.Errorf("get created transaction: %w", err)
	}

	return transaction, newBalance, nil
}

// nullableString returns nil for empty strings, otherwise a pointer to the trimmed string.
func nullableString(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
