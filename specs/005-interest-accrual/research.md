# Research: Interest Accrual

## R-001: Interest Rate Storage

**Decision**: Add `interest_rate_bps` column to the `children` table (per-child, integer basis points)

**Rationale**: The spec requires per-child interest rates (FR-002). Storing directly on the `children` table is simplest — one column, no join needed, follows the existing pattern of `balance_cents` being on `children`. Using basis points (integer, 500 = 5.00%) avoids floating-point precision issues with money calculations. The `addColumnIfNotExists` helper already supports this migration pattern.

**Alternatives considered**:
- Separate `interest_rates` table: More flexible for rate history, but YAGNI — the spec doesn't require rate change history, only the current rate
- `REAL` type for percentage: Floating-point introduces precision concerns for financial calculations
- Family-wide rate in `families` table: Doesn't satisfy FR-002 (per-child rates)

## R-002: Duplicate Prevention

**Decision**: Add `last_interest_at` column to `children` table, storing the timestamp of the last successful interest accrual. Before accruing, check that the current month/year differs from the last accrual month/year.

**Rationale**: FR-010 requires preventing duplicate accrual for the same period. Storing the last accrual timestamp on the child row is atomic — it's updated in the same database transaction as the interest deposit. Comparing month/year is simple and matches the monthly accrual requirement.

**Alternatives considered**:
- Separate `interest_accruals` tracking table: Extra complexity for a simple flag
- Checking transaction history for existing interest transactions: Slower query, relies on transaction data integrity rather than explicit tracking

## R-003: Transaction Type

**Decision**: Add `interest` to the transaction type CHECK constraint

**Rationale**: Interest payments are transactions recorded in the existing `transactions` table (FR-007). The existing types are `deposit`, `withdrawal`, `allowance`. Adding `interest` follows the same pattern. Requires table recreation to update the CHECK constraint (same pattern used in `migrateTransactionsCheckConstraint()`).

**Alternatives considered**:
- Separate `interest_transactions` table: Duplicates the transaction infrastructure; the existing table already has everything needed
- Using `deposit` type with a special note: Violates FR-008 (must be clearly distinguishable)

## R-004: Scheduler Pattern

**Decision**: Create `interest.Scheduler` following the same goroutine pattern as `allowance.Scheduler`

**Rationale**: The allowance scheduler (feature 003) provides a proven pattern — `time.NewTicker` with a stop channel, `ProcessDue()` method, started in `main.go`. Interest accrual uses the same pattern. The scheduler checks periodically (e.g., every hour) and processes children whose `last_interest_at` is not in the current month.

**Alternatives considered**:
- Cron-style scheduling: Requires additional dependency; the ticker pattern is already proven in this codebase
- Manual trigger only: Doesn't satisfy FR-004 (automatic monthly schedule)

## R-005: Interest Calculation

**Decision**: Calculate as `balance_cents * rate_bps / 12 / 10000`, rounded to nearest cent using standard banker's rounding

**Rationale**: FR-005 specifies `(current balance) × (annual rate / 12)`. Using basis points: rate_bps of 500 = 5.00%, so `balance_cents * 500 / 12 / 10000` gives the monthly interest in cents. Integer arithmetic throughout avoids floating-point issues. Results below 0.5 cents round to 0 and are skipped (per edge case spec).

**Alternatives considered**:
- Float-based calculation: Precision concerns with money
- Daily accrual accumulated monthly: More complex, spec says monthly is sufficient

## R-006: Interest Rate API

**Decision**: Add interest rate to existing child management endpoints rather than creating new endpoints

**Rationale**: The parent already manages children via family handlers. Adding `PUT /api/children/{id}/interest-rate` follows the existing endpoint pattern. The interest rate is returned as part of the child object in existing responses. Simplicity principle — no new endpoint groups needed.

**Alternatives considered**:
- Separate interest configuration API: Over-engineered for a single field

## R-007: Frontend Display

**Decision**: Add `interest` case to existing transaction type display logic and show interest rate in parent's child management view

**Rationale**: The frontend `TransactionHistory` component already handles multiple transaction types with labels. Adding an `interest` case with a distinct label ("Interest earned") and color satisfies US3 and US4 with minimal changes. The parent dashboard shows child cards — adding the interest rate display there is natural.

**Alternatives considered**:
- Separate interest dashboard page: YAGNI — interest transactions appear in the existing transaction history

## R-008: Parent ID for Interest Transactions

**Decision**: Use the family's parent ID as the `parent_id` on interest transactions

**Rationale**: The `transactions` table requires `parent_id NOT NULL`. For interest transactions (system-generated), we use the parent who owns the family. This maintains referential integrity and allows the parent to see interest transactions in their view.

**Alternatives considered**:
- Nullable parent_id: Would require schema change and break existing queries
- System user ID: Adds unnecessary complexity
