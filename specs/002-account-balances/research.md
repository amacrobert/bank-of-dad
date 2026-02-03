# Research: Account Balances

**Feature**: 002-account-balances
**Date**: 2026-02-03

## Research Summary

No NEEDS CLARIFICATION items in Technical Context. All technical decisions align with existing codebase patterns discovered during exploration.

---

## Decision 1: Money Storage Format

**Decision**: Store monetary amounts as INTEGER (cents) in SQLite

**Rationale**:
- Avoids floating-point precision issues inherent in decimal calculations
- SQLite's DECIMAL type is actually stored as TEXT or REAL, not true decimal
- Integer arithmetic is precise and fast
- Common pattern in financial applications (Stripe, Square, etc.)
- Frontend displays by dividing by 100: `$12.50` stored as `1250`

**Alternatives Considered**:
- TEXT with manual parsing: More complex, no benefit over integers
- REAL (floating point): Precision loss risk (e.g., 0.1 + 0.2 ≠ 0.3)
- Third-party decimal library: Adds dependency, violates Simplicity principle

**Implementation Notes**:
- All amounts in API requests/responses use cents (integers)
- Frontend formats for display: `(cents / 100).toFixed(2)`
- Validation: amounts must be positive integers > 0

---

## Decision 2: Balance Storage Strategy

**Decision**: Store balance as a field on the `children` table, with transactions in a separate `transactions` table

**Rationale**:
- Simple, performant reads for dashboard display (single column query)
- Transactions table provides complete audit trail
- Balance is updated atomically within transaction insert
- Follows existing pattern: children table has all child-specific data

**Alternatives Considered**:
- Calculate balance from transactions each time: Slow for large transaction histories
- Separate balances table: Extra join for no benefit, children already have 1:1 relationship
- Event sourcing: Overengineered for this scale, violates Simplicity

**Implementation Notes**:
- Add `balance_cents INTEGER NOT NULL DEFAULT 0` to children table
- Transactions table stores: amount, type, note, timestamps, actor (parent)
- Balance update and transaction insert happen in same SQL transaction

---

## Decision 3: Transaction Types

**Decision**: Two transaction types: `deposit` and `withdrawal`

**Rationale**:
- Maps directly to spec requirements (FR-003, FR-004)
- Simple, clear semantics
- Deposit = positive change to balance
- Withdrawal = negative change to balance (parent removes money for child purchase)

**Alternatives Considered**:
- Single `amount` with sign: Confusing semantics, harder to query
- Multiple types (allowance, reward, purchase, etc.): Out of scope per spec, adds complexity

**Implementation Notes**:
- Type stored as TEXT enum: 'deposit' | 'withdrawal'
- Go const for type safety: `TransactionTypeDeposit`, `TransactionTypeWithdrawal`
- Frontend uses TypeScript string literal types

---

## Decision 4: Authorization Model

**Decision**: Use existing middleware pattern with additional child-specific checks

**Rationale**:
- `RequireAuth` middleware already validates sessions
- `RequireParent` middleware already checks user_type = 'parent'
- Add route-level checks: parent can only access own children (existing pattern in child handlers)
- Children can only view their own balance/transactions

**Alternatives Considered**:
- Custom balance middleware: Unnecessary duplication
- Role-based access control library: Overkill for two roles

**Implementation Notes**:
- Deposit/withdraw endpoints: `RequireParent` + verify child belongs to parent's family
- Balance view endpoint: `RequireAuth` + verify (child viewing self OR parent viewing own child)
- Transaction history: Same as balance view

---

## Decision 5: API Response Format

**Decision**: Follow existing JSON response patterns with snake_case keys

**Rationale**:
- Consistency with existing API (e.g., `first_name`, `family_id`)
- Existing `writeJSON` helper in auth package
- Frontend already handles snake_case from API

**Alternatives Considered**:
- camelCase: Would break consistency with existing endpoints

**Implementation Notes**:
- Balance response: `{ "child_id": 1, "balance_cents": 1250 }`
- Transaction response: `{ "id": 1, "amount_cents": 500, "type": "deposit", "note": "Weekly allowance", "created_at": "2026-02-03T12:00:00Z" }`

---

## Decision 6: Transaction History Pagination

**Decision**: Initial implementation without pagination; add if performance issues arise

**Rationale**:
- YAGNI: Most children will have <100 transactions initially
- SQLite handles small result sets efficiently
- Edge case (hundreds of transactions) acknowledged in spec but not immediate concern
- Constitution's "No premature optimization" principle

**Alternatives Considered**:
- Cursor-based pagination from start: Adds complexity before it's needed
- Limit to N most recent: Would hide history, not acceptable per spec

**Implementation Notes**:
- `ORDER BY created_at DESC` for newest-first display
- Monitor query performance; add LIMIT/OFFSET if needed
- Consider index on `(child_id, created_at)` for transaction table

---

## Decision 7: Handling Concurrent Modifications

**Decision**: SQLite's serialized writes with WAL mode provide sufficient protection

**Rationale**:
- Existing database setup uses maxOpenConns=1 for writes
- All balance updates are in transactions (INSERT transaction + UPDATE balance)
- SQLite WAL mode allows concurrent reads during writes
- Low concurrency expected (family app, not high-traffic)

**Alternatives Considered**:
- Row-level locking: SQLite doesn't support it
- Optimistic locking with version column: Overkill for expected load
- External queue: Way over-engineered

**Implementation Notes**:
- Each deposit/withdrawal is atomic: BEGIN → INSERT transaction → UPDATE balance → COMMIT
- If commit fails, entire operation rolls back
- Error returned to client for retry

---

## Existing Code Patterns to Follow

### Store Pattern (from child.go)
```go
type BalanceStore struct {
    db *DB
}

func NewBalanceStore(db *DB) *BalanceStore {
    return &BalanceStore{db: db}
}

func (s *BalanceStore) GetBalance(childID int64) (int64, error) {
    // Query child.balance_cents
}

func (s *BalanceStore) Deposit(childID, parentID int64, amountCents int64, note string) (*Transaction, error) {
    // BEGIN transaction
    // INSERT INTO transactions
    // UPDATE children SET balance_cents = balance_cents + amount
    // COMMIT
}
```

### Handler Pattern (from child.go)
```go
type BalanceHandler struct {
    balanceStore *store.BalanceStore
    txStore      *store.TransactionStore
    sessionStore *store.SessionStore
}

func (h *BalanceHandler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
    // Extract child ID from URL
    // Verify authorization (parent owns child OR child is self)
    // Get balance
    // writeJSON response
}
```

### Test Pattern (from child_test.go)
```go
func TestBalanceDeposit(t *testing.T) {
    db := openTestDB(t)
    store := NewBalanceStore(db)

    // Create test family and child
    family := createTestFamily(t, db)
    child := createTestChild(t, db, family.ID)

    // Test deposit
    tx, err := store.Deposit(child.ID, parent.ID, 1000, "Test deposit")
    require.NoError(t, err)
    assert.Equal(t, int64(1000), tx.AmountCents)

    // Verify balance updated
    balance, err := store.GetBalance(child.ID)
    require.NoError(t, err)
    assert.Equal(t, int64(1000), balance)
}
```

---

## Dependencies

No new dependencies required. Feature uses:
- `database/sql` (existing)
- `modernc.org/sqlite` (existing)
- `testify` (existing)
- React (existing)
- react-router-dom (existing)
