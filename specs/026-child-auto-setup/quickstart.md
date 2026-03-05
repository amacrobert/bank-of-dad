# Quickstart: Child Auto-Setup

## What This Feature Does

Adds three optional fields to the Add Child form (Initial Deposit, Weekly Allowance, Annual Interest) so parents can fully configure a child's account in one step instead of four separate forms.

## Files to Modify

| File | Change |
|------|--------|
| `frontend/src/components/AddChildForm.tsx` | Add 3 optional input fields, post-creation API orchestration, updated success state |

## Key API Functions (already exist in `frontend/src/api.ts`)

```typescript
// Create child (existing)
post<ChildCreateResponse>("/children", { first_name, password, avatar? })

// Initial deposit (existing)
deposit(childId, { amount_cents, note: "Initial deposit" })

// Weekly allowance (existing)
setChildAllowance(childId, { amount_cents, frequency: "weekly", day_of_week, note: "Weekly allowance" })

// Annual interest (existing)
setInterest(childId, { interest_rate_bps, frequency: "monthly", day_of_month: 1 })
```

## Implementation Pattern

1. Add state for `initialDeposit`, `weeklyAllowance`, `annualInterest` (all strings, like existing amount inputs)
2. After successful child creation, sequentially call deposit → allowance → interest APIs (skipping any with zero/empty values)
3. Track setup errors separately from creation errors — child is kept even if setup partially fails
4. Update success card to show what was configured

## Input Patterns to Reuse

- **Currency**: `$` prefix + `type="number" step="0.01" min="0" max="999999.99"` (from DepositForm)
- **Percentage**: `type="number" step="0.01" min="0" max="100"` + `%` suffix (from InterestForm)
- **Conversion**: `Math.round(parseFloat(amount) * 100)` for cents, `Math.round(parseFloat(rate) * 100)` for basis points

## Verification

```bash
cd frontend && npx tsc --noEmit && npm run build && npm run lint
```
