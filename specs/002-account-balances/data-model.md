# Data Model: Account Balances

**Feature**: 002-account-balances
**Date**: 2026-02-03

## Entity Overview

```
┌─────────────┐       ┌─────────────┐       ┌──────────────────┐
│   Family    │──1:N──│   Parent    │       │                  │
└─────────────┘       └─────────────┘       │                  │
      │                     │               │                  │
      │                     │ creates       │                  │
      │                     ▼               │                  │
      │              ┌──────────────┐       │                  │
      └──────1:N─────│    Child     │───────│  has balance     │
                     │              │       │                  │
                     │ balance_cents│       │                  │
                     └──────────────┘       │                  │
                           │                │                  │
                           │ 1:N            │                  │
                           ▼                │                  │
                     ┌──────────────┐       │                  │
                     │ Transaction  │◀──────┘                  │
                     │              │  created by parent       │
                     └──────────────┘                          │
```

## Schema Changes

### Modify: `children` table

Add balance field to existing children table:

```sql
-- Migration: Add balance_cents column
ALTER TABLE children ADD COLUMN balance_cents INTEGER NOT NULL DEFAULT 0;
```

**Updated children table structure:**
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Child ID |
| family_id | INTEGER | NOT NULL, FK(families) | Family membership |
| first_name | TEXT | NOT NULL | Child's name |
| password_hash | TEXT | NOT NULL | Bcrypt hash |
| is_locked | INTEGER | DEFAULT 0 | Account lockout flag |
| failed_login_attempts | INTEGER | DEFAULT 0 | Login attempt counter |
| **balance_cents** | **INTEGER** | **NOT NULL DEFAULT 0** | **Current balance in cents (NEW)** |
| created_at | TEXT | NOT NULL | Creation timestamp |
| updated_at | TEXT | NOT NULL | Last update timestamp |

### New: `transactions` table

```sql
CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id INTEGER NOT NULL,
    parent_id INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    transaction_type TEXT NOT NULL CHECK(transaction_type IN ('deposit', 'withdrawal')),
    note TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),

    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT
);

-- Index for efficient child transaction history queries
CREATE INDEX idx_transactions_child_created ON transactions(child_id, created_at DESC);
```

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Transaction ID |
| child_id | INTEGER | NOT NULL, FK(children) | Child account affected |
| parent_id | INTEGER | NOT NULL, FK(parents) | Parent who made the change |
| amount_cents | INTEGER | NOT NULL | Amount in cents (always positive) |
| transaction_type | TEXT | NOT NULL, CHECK IN | 'deposit' or 'withdrawal' |
| note | TEXT | NULLABLE | Optional description |
| created_at | TEXT | NOT NULL | Transaction timestamp |

## Go Model Definitions

### Transaction Model

```go
// backend/internal/store/transaction.go

type TransactionType string

const (
    TransactionTypeDeposit    TransactionType = "deposit"
    TransactionTypeWithdrawal TransactionType = "withdrawal"
)

type Transaction struct {
    ID              int64           `json:"id"`
    ChildID         int64           `json:"child_id"`
    ParentID        int64           `json:"parent_id"`
    AmountCents     int64           `json:"amount_cents"`
    TransactionType TransactionType `json:"type"`
    Note            *string         `json:"note,omitempty"`
    CreatedAt       time.Time       `json:"created_at"`
}
```

### Balance Response Model

```go
// Used for API responses
type BalanceResponse struct {
    ChildID      int64 `json:"child_id"`
    BalanceCents int64 `json:"balance_cents"`
}

type ChildWithBalance struct {
    ID           int64  `json:"id"`
    FirstName    string `json:"first_name"`
    BalanceCents int64  `json:"balance_cents"`
}
```

## TypeScript Model Definitions

### Frontend Types

```typescript
// frontend/src/types.ts

export type TransactionType = 'deposit' | 'withdrawal';

export interface Transaction {
  id: number;
  child_id: number;
  parent_id: number;
  amount_cents: number;
  type: TransactionType;
  note?: string;
  created_at: string;
}

export interface BalanceResponse {
  child_id: number;
  balance_cents: number;
}

export interface ChildWithBalance {
  id: number;
  first_name: string;
  balance_cents: number;
}

export interface TransactionListResponse {
  transactions: Transaction[];
}

export interface DepositRequest {
  amount_cents: number;
  note?: string;
}

export interface WithdrawRequest {
  amount_cents: number;
  note?: string;
}
```

## Validation Rules

### Amount Validation
- Must be a positive integer > 0
- Maximum: 99999999 cents ($999,999.99) - practical limit
- No decimal values accepted (API uses cents)

### Note Validation
- Optional field
- Maximum length: 500 characters
- Trimmed of leading/trailing whitespace
- Empty string treated as NULL

### Withdrawal Validation
- Cannot exceed current balance
- Results in exactly $0.00 is allowed
- Negative resulting balance rejected with error

## State Transitions

### Balance State Machine

```
                    ┌─────────────┐
                    │   $0.00     │ (initial state)
                    └──────┬──────┘
                           │
                   deposit │
                           ▼
                    ┌─────────────┐
         ┌─────────│   $X.XX     │─────────┐
         │         └─────────────┘         │
         │               │                 │
  deposit│               │withdrawal       │withdrawal
  (+$Y)  │               │(-$Y where      │(full balance)
         │               │ Y < balance)   │
         │               │                 │
         ▼               ▼                 ▼
    ┌─────────┐    ┌─────────┐       ┌─────────┐
    │ $X+Y.XX │    │ $X-Y.XX │       │  $0.00  │
    └─────────┘    └─────────┘       └─────────┘
         │               │
         └───────┬───────┘
                 │
         (continues...)
```

**Invariants:**
- Balance is never negative
- Every balance change creates a transaction record
- Transaction + balance update are atomic

## Relationships

| From | To | Cardinality | Constraint |
|------|-----|-------------|------------|
| Child | Transaction | 1:N | Child has many transactions |
| Parent | Transaction | 1:N | Parent creates transactions |
| Family | Child | 1:N | Family has many children (existing) |
| Parent | Family | N:1 | Parent belongs to family (existing) |

## Indexes

| Table | Index | Columns | Purpose |
|-------|-------|---------|---------|
| transactions | idx_transactions_child_created | (child_id, created_at DESC) | Transaction history queries |

## Migration Strategy

1. Add `balance_cents` column to children table (default 0)
2. Create transactions table with foreign keys
3. Create index on transactions
4. No data migration needed (new feature, all balances start at 0)

```sql
-- Full migration script
ALTER TABLE children ADD COLUMN balance_cents INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id INTEGER NOT NULL,
    parent_id INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    transaction_type TEXT NOT NULL CHECK(transaction_type IN ('deposit', 'withdrawal')),
    note TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT
);

CREATE INDEX idx_transactions_child_created ON transactions(child_id, created_at DESC);
```
