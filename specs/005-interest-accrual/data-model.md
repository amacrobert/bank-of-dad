# Data Model: Interest Accrual

## Schema Changes

### Modified Table: `children`

Two new columns added to the existing `children` table:

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `interest_rate_bps` | INTEGER | 0 | Annual interest rate in basis points (500 = 5.00%). Range: 0–10000 (0%–100%) |
| `last_interest_at` | DATETIME | NULL | Timestamp of the last successful interest accrual for this child. NULL means never accrued. |

**Migration**: Use `addColumnIfNotExists()` helper (idempotent).

### Modified Table: `transactions`

The `transaction_type` CHECK constraint must be updated to include `interest`:

**Before**: `CHECK(transaction_type IN ('deposit', 'withdrawal', 'allowance'))`
**After**: `CHECK(transaction_type IN ('deposit', 'withdrawal', 'allowance', 'interest'))`

**Migration**: Table recreation pattern (same as `migrateTransactionsCheckConstraint()` from feature 002).

### Interest Transaction Fields

Interest transactions use the existing `transactions` table with these field mappings:

| Field | Value for Interest Transactions |
|-------|--------------------------------|
| `child_id` | The child receiving interest |
| `parent_id` | The parent who owns the family |
| `amount_cents` | Calculated interest amount (always positive) |
| `transaction_type` | `"interest"` |
| `note` | Interest rate description, e.g., "5.00% annual rate" |
| `schedule_id` | NULL (not associated with an allowance schedule) |
| `created_at` | Timestamp of accrual |

## Entity Relationships

```
families (1) ──── (N) children
                      ├── interest_rate_bps   (new)
                      └── last_interest_at     (new)
                          │
children (1) ──── (N) transactions
                      └── transaction_type = 'interest' (new type)
```

## Validation Rules

- `interest_rate_bps` MUST be between 0 and 10000 (inclusive)
- Interest is only accrued when `interest_rate_bps > 0` AND `balance_cents > 0`
- `last_interest_at` is updated atomically with the interest transaction insertion
- Duplicate prevention: accrual skipped if `last_interest_at` is in the current calendar month

## Interest Calculation

```
monthly_interest_cents = round(balance_cents * interest_rate_bps / 12 / 10000)
```

- If `monthly_interest_cents` rounds to 0, no transaction is created
- Rounding: standard half-up (0.5 rounds to 1)
- All arithmetic uses integers (cents and basis points) — no floating point
