package repositories

import (
	"errors"
	"fmt"
	"strings"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// TransactionRepo handles database operations for transactions using GORM.
type TransactionRepo struct {
	db *gorm.DB
}

// NewTransactionRepo creates a new TransactionRepo.
func NewTransactionRepo(db *gorm.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

// Deposit adds money to a child's account and records the transaction.
// The operation is atomic - both the transaction record and balance update happen together.
func (r *TransactionRepo) Deposit(childID, parentID, amountCents int64, note string) (*models.Transaction, int64, error) {
	if amountCents <= 0 {
		return nil, 0, fmt.Errorf("amount must be positive")
	}

	var transaction models.Transaction
	var newBalance int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Insert transaction record
		notePtr := nullableString(note)
		transaction = models.Transaction{
			ChildID:         childID,
			ParentID:        parentID,
			AmountCents:     amountCents,
			TransactionType: models.TransactionTypeDeposit,
			Note:            notePtr,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}

		// Update balance
		if err := tx.Exec(
			`UPDATE children SET balance_cents = balance_cents + ?, updated_at = NOW() WHERE id = ?`,
			amountCents, childID,
		).Error; err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		// Get new balance
		var child models.Child
		if err := tx.Select("balance_cents").First(&child, childID).Error; err != nil {
			return fmt.Errorf("get new balance: %w", err)
		}
		newBalance = child.BalanceCents

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return &transaction, newBalance, nil
}

// Withdraw removes money from a child's account and records the transaction.
// Returns ErrInsufficientFunds if the withdrawal would result in a negative balance.
func (r *TransactionRepo) Withdraw(childID, parentID, amountCents int64, note string) (*models.Transaction, int64, error) {
	if amountCents <= 0 {
		return nil, 0, fmt.Errorf("amount must be positive")
	}

	var transaction models.Transaction
	var newBalance int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Check current balance
		var child models.Child
		if err := tx.Select("balance_cents").First(&child, childID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("child not found")
			}
			return fmt.Errorf("get current balance: %w", err)
		}

		if child.BalanceCents < amountCents {
			newBalance = child.BalanceCents
			return models.ErrInsufficientFunds
		}

		// Insert transaction record
		notePtr := nullableString(note)
		transaction = models.Transaction{
			ChildID:         childID,
			ParentID:        parentID,
			AmountCents:     amountCents,
			TransactionType: models.TransactionTypeWithdrawal,
			Note:            notePtr,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}

		// Update balance
		if err := tx.Exec(
			`UPDATE children SET balance_cents = balance_cents - ?, updated_at = NOW() WHERE id = ?`,
			amountCents, childID,
		).Error; err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		// Get new balance
		var updated models.Child
		if err := tx.Select("balance_cents").First(&updated, childID).Error; err != nil {
			return fmt.Errorf("get new balance: %w", err)
		}
		newBalance = updated.BalanceCents

		return nil
	})
	if err != nil {
		if err == models.ErrInsufficientFunds {
			return nil, newBalance, err
		}
		return nil, 0, err
	}

	return &transaction, newBalance, nil
}

// GetByID retrieves a transaction by its ID.
func (r *TransactionRepo) GetByID(id int64) (*models.Transaction, error) {
	var t models.Transaction
	err := r.db.First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}
	return &t, nil
}

// ListByChild retrieves all transactions for a child, ordered by most recent first.
func (r *TransactionRepo) ListByChild(childID int64) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("child_id = ?", childID).
		Order("created_at DESC, id DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	return transactions, nil
}

// ListByChildPaginated retrieves transactions for a child with limit/offset pagination.
func (r *TransactionRepo) ListByChildPaginated(childID int64, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("child_id = ?", childID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("list transactions paginated: %w", err)
	}
	return transactions, nil
}

// DepositAllowance adds money to a child's account as a scheduled allowance transaction.
// Similar to Deposit but includes a schedule_id and uses "allowance" transaction type.
func (r *TransactionRepo) DepositAllowance(childID, parentID, amountCents, scheduleID int64, note string) (*models.Transaction, int64, error) {
	if amountCents <= 0 {
		return nil, 0, fmt.Errorf("amount must be positive")
	}

	var transaction models.Transaction
	var newBalance int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		notePtr := nullableString(note)
		transaction = models.Transaction{
			ChildID:         childID,
			ParentID:        parentID,
			AmountCents:     amountCents,
			TransactionType: models.TransactionTypeAllowance,
			Note:            notePtr,
			ScheduleID:      &scheduleID,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}

		// Update balance
		if err := tx.Exec(
			`UPDATE children SET balance_cents = balance_cents + ?, updated_at = NOW() WHERE id = ?`,
			amountCents, childID,
		).Error; err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		// Get new balance
		var child models.Child
		if err := tx.Select("balance_cents").First(&child, childID).Error; err != nil {
			return fmt.Errorf("get new balance: %w", err)
		}
		newBalance = child.BalanceCents

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return &transaction, newBalance, nil
}

// DepositChore adds money to a child's account as a chore reward transaction.
// Similar to Deposit but uses "chore" transaction type.
func (r *TransactionRepo) DepositChore(childID, parentID int64, amountCents int64, note string) (*models.Transaction, int64, error) {
	if amountCents < 0 {
		return nil, 0, fmt.Errorf("amount must not be negative")
	}

	var transaction models.Transaction
	var newBalance int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		notePtr := nullableString(note)
		transaction = models.Transaction{
			ChildID:         childID,
			ParentID:        parentID,
			AmountCents:     amountCents,
			TransactionType: models.TransactionTypeChore,
			Note:            notePtr,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}

		if amountCents > 0 {
			// Update balance
			if err := tx.Exec(
				`UPDATE children SET balance_cents = balance_cents + ?, updated_at = NOW() WHERE id = ?`,
				amountCents, childID,
			).Error; err != nil {
				return fmt.Errorf("update balance: %w", err)
			}
		}

		// Get current balance
		var child models.Child
		if err := tx.Select("balance_cents").First(&child, childID).Error; err != nil {
			return fmt.Errorf("get new balance: %w", err)
		}
		newBalance = child.BalanceCents

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return &transaction, newBalance, nil
}

// nullableString returns nil for empty strings, otherwise a pointer to the trimmed string.
func nullableString(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
