# Quickstart: Account Balances

**Feature**: 002-account-balances
**Date**: 2026-02-03

## Overview

This guide provides implementation context for the account balances feature. Follow Test-Driven Development (TDD) as required by the project constitution.

## Getting Started

### Prerequisites

1. Docker and Docker Compose running
2. Go 1.24+ installed
3. Node.js 18+ installed
4. Existing codebase from 001-user-auth feature

### Development Environment

```bash
# Start the full stack
docker compose up -d

# Backend development (with hot reload)
cd backend
go run main.go

# Frontend development
cd frontend
npm run dev
```

## Implementation Order

Follow this order to build the feature incrementally with tests first:

### Phase 1: Database Layer

1. **Add migration for balance_cents column**
   - Modify `backend/internal/store/db.go` migration
   - Add `balance_cents INTEGER NOT NULL DEFAULT 0` to children table

2. **Create transactions table**
   - Add CREATE TABLE for transactions in migration
   - Add index for child_id + created_at

### Phase 2: Store Layer (Test First)

1. **Create TransactionStore**
   - Write tests in `transaction_test.go` first
   - Implement `Create()`, `ListByChild()` methods

2. **Extend ChildStore for balance operations**
   - Write tests for balance operations first
   - Add `GetBalance()` method
   - Add `UpdateBalance()` method (internal, used by transaction)

3. **Create atomic deposit/withdraw operations**
   - Test transaction + balance update atomicity
   - Implement `Deposit()` and `Withdraw()` in TransactionStore

### Phase 3: API Handlers (Test First)

1. **Create BalanceHandler**
   - Test `HandleGetBalance` first
   - Test authorization (parent sees own children, child sees self)

2. **Create TransactionHandler**
   - Test `HandleDeposit` first
   - Test `HandleWithdraw` with insufficient funds case
   - Test `HandleGetTransactions`

3. **Register routes in main.go**

### Phase 4: Frontend Components

1. **Add TypeScript types** (`types.ts`)
2. **Add API functions** (`api.ts`)
3. **Create BalanceDisplay component**
4. **Create TransactionHistory component**
5. **Create DepositForm and WithdrawForm**
6. **Update ParentDashboard** to show balances
7. **Update ChildDashboard** to show balance and history

## Key Code Patterns

### Store Pattern (following existing code)

```go
// backend/internal/store/transaction.go
type TransactionStore struct {
    db *DB
}

func NewTransactionStore(db *DB) *TransactionStore {
    return &TransactionStore{db: db}
}

// Deposit creates a transaction and updates balance atomically
func (s *TransactionStore) Deposit(childID, parentID, amountCents int64, note string) (*Transaction, error) {
    tx, err := s.db.writeDB.Begin()
    if err != nil {
        return nil, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Insert transaction record
    result, err := tx.Exec(`
        INSERT INTO transactions (child_id, parent_id, amount_cents, transaction_type, note)
        VALUES (?, ?, ?, 'deposit', ?)
    `, childID, parentID, amountCents, nullableString(note))
    if err != nil {
        return nil, fmt.Errorf("insert transaction: %w", err)
    }

    // Update balance
    _, err = tx.Exec(`
        UPDATE children SET balance_cents = balance_cents + ?, updated_at = datetime('now')
        WHERE id = ?
    `, amountCents, childID)
    if err != nil {
        return nil, fmt.Errorf("update balance: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit: %w", err)
    }

    // Return created transaction
    txID, _ := result.LastInsertId()
    return s.GetByID(txID)
}
```

### Handler Pattern (following existing code)

```go
// backend/internal/balance/handler.go
type Handler struct {
    txStore      *store.TransactionStore
    childStore   *store.ChildStore
    sessionStore *store.SessionStore
}

func (h *Handler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
    // Get child ID from URL
    childIDStr := r.PathValue("id")
    childID, err := strconv.ParseInt(childIDStr, 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid_child_id", "Invalid child ID")
        return
    }

    // Get user from context (set by auth middleware)
    userType := r.Context().Value("user_type").(string)
    userID := r.Context().Value("user_id").(int64)
    familyID := r.Context().Value("family_id").(int64)

    // Verify authorization
    child, err := h.childStore.GetByID(childID)
    if err != nil {
        writeError(w, http.StatusNotFound, "not_found", "Child not found")
        return
    }

    if child.FamilyID != familyID {
        writeError(w, http.StatusForbidden, "forbidden", "Access denied")
        return
    }

    if userType == "child" && userID != childID {
        writeError(w, http.StatusForbidden, "forbidden", "Access denied")
        return
    }

    // Return balance
    writeJSON(w, http.StatusOK, BalanceResponse{
        ChildID:      child.ID,
        BalanceCents: child.BalanceCents,
    })
}
```

### Test Pattern (following existing code)

```go
// backend/internal/store/transaction_test.go
func TestDeposit(t *testing.T) {
    db := openTestDB(t)
    txStore := NewTransactionStore(db)
    childStore := NewChildStore(db)

    // Setup: create family, parent, child
    family := createTestFamily(t, db)
    parent := createTestParent(t, db, family.ID)
    child := createTestChild(t, db, family.ID)

    // Verify initial balance is 0
    balance, err := childStore.GetBalance(child.ID)
    require.NoError(t, err)
    assert.Equal(t, int64(0), balance)

    // Test deposit
    tx, err := txStore.Deposit(child.ID, parent.ID, 1000, "Weekly allowance")
    require.NoError(t, err)
    assert.Equal(t, int64(1000), tx.AmountCents)
    assert.Equal(t, TransactionTypeDeposit, tx.TransactionType)
    assert.Equal(t, "Weekly allowance", *tx.Note)

    // Verify balance updated
    balance, err = childStore.GetBalance(child.ID)
    require.NoError(t, err)
    assert.Equal(t, int64(1000), balance)
}

func TestWithdrawInsufficientFunds(t *testing.T) {
    db := openTestDB(t)
    txStore := NewTransactionStore(db)

    family := createTestFamily(t, db)
    parent := createTestParent(t, db, family.ID)
    child := createTestChild(t, db, family.ID)

    // Deposit $10
    _, err := txStore.Deposit(child.ID, parent.ID, 1000, "")
    require.NoError(t, err)

    // Try to withdraw $20 - should fail
    _, err = txStore.Withdraw(child.ID, parent.ID, 2000, "")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient funds")
}
```

### Frontend Component Pattern

```tsx
// frontend/src/components/BalanceDisplay.tsx
interface BalanceDisplayProps {
  balanceCents: number;
}

export function BalanceDisplay({ balanceCents }: BalanceDisplayProps) {
  const formatted = (balanceCents / 100).toFixed(2);

  return (
    <div className="balance-display">
      <span className="balance-label">Balance:</span>
      <span className="balance-amount">${formatted}</span>
    </div>
  );
}
```

```tsx
// frontend/src/components/TransactionHistory.tsx
interface TransactionHistoryProps {
  transactions: Transaction[];
}

export function TransactionHistory({ transactions }: TransactionHistoryProps) {
  if (transactions.length === 0) {
    return <p>No transactions yet.</p>;
  }

  return (
    <ul className="transaction-list">
      {transactions.map((tx) => (
        <li key={tx.id} className={`transaction ${tx.type}`}>
          <span className="date">
            {new Date(tx.created_at).toLocaleDateString()}
          </span>
          <span className="type">{tx.type}</span>
          <span className="amount">
            {tx.type === 'deposit' ? '+' : '-'}${(tx.amount_cents / 100).toFixed(2)}
          </span>
          {tx.note && <span className="note">{tx.note}</span>}
        </li>
      ))}
    </ul>
  );
}
```

## API Endpoints Summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/children` | Parent | List children with balances |
| GET | `/api/children/{id}/balance` | Parent/Child | Get balance |
| POST | `/api/children/{id}/deposit` | Parent | Add money |
| POST | `/api/children/{id}/withdraw` | Parent | Remove money |
| GET | `/api/children/{id}/transactions` | Parent/Child | Transaction history |

## Testing Commands

```bash
# Run backend tests
cd backend
go test ./...

# Run specific store tests
go test ./internal/store -v -run TestDeposit

# Run frontend tests
cd frontend
npm test
```

## Common Pitfalls

1. **Floating point for money**: Use integers (cents), never float64
2. **Missing atomicity**: Always wrap balance update + transaction insert in SQL transaction
3. **Authorization bypass**: Always verify family_id matches before allowing access
4. **Negative balance**: Check balance before withdrawal, not after
5. **Missing tests**: Write the test first, watch it fail, then implement
