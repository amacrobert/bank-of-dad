# Data Model: Savings Growth Projector

**Feature**: 016-savings-growth-projector
**Date**: 2026-02-17

## Overview

This feature introduces **no new database tables or backend changes**. All data models are client-side TypeScript interfaces representing computed projection data and transient user inputs.

## Existing Entities Used (Read-Only)

### BalanceResponse (existing)
- `balance_cents`: int64 — child's current balance in cents
- `interest_rate_bps`: int — annual interest rate in basis points (e.g., 500 = 5%)
- `interest_rate_display`: string — formatted rate (e.g., "5.00%")

### AllowanceSchedule (existing)
- `amount_cents`: int64 — per-payment amount in cents
- `frequency`: "weekly" | "biweekly" | "monthly"
- `status`: "active" | "paused"

### InterestSchedule (existing)
- `frequency`: "weekly" | "biweekly" | "monthly"
- `status`: "active" | "paused"

## New Client-Side Entities

### ScenarioInputs
Transient user-adjustable parameters. Not persisted.

| Field | Type | Default | Validation |
|-------|------|---------|------------|
| weeklySpendingCents | number | 0 | >= 0 |
| oneTimeDepositCents | number | 0 | >= 0 |
| oneTimeWithdrawalCents | number | 0 | >= 0, <= current balance |
| horizonMonths | number | 12 | enum: 3, 6, 12, 24, 60 |

### ProjectionDataPoint
A single point in the projected time series.

| Field | Type | Description |
|-------|------|-------------|
| weekIndex | number | Weeks from now (0 = today) |
| date | string | ISO date string for this week |
| balanceCents | number | Projected balance at this point (floored at 0) |

### ProjectionResult
The complete output of a projection calculation.

| Field | Type | Description |
|-------|------|-------------|
| dataPoints | ProjectionDataPoint[] | Time series for the graph |
| finalBalanceCents | number | Balance at end of horizon |
| totalInterestCents | number | Cumulative interest earned |
| totalAllowanceCents | number | Cumulative allowance received |
| totalSpendingCents | number | Cumulative spending deducted |
| startingBalanceCents | number | Effective starting balance (current + deposit - withdrawal) |
| depletionWeek | number | null | Week when balance hits 0, if applicable |

### ProjectionConfig
Inputs to the projection engine derived from existing data + scenario.

| Field | Type | Description |
|-------|------|-------------|
| currentBalanceCents | number | From BalanceResponse |
| interestRateBps | number | From BalanceResponse |
| interestFrequency | Frequency | null | From InterestSchedule (null if none/paused) |
| allowanceAmountCents | number | From AllowanceSchedule (0 if none/paused) |
| allowanceFrequency | Frequency | null | From AllowanceSchedule (null if none/paused) |
| scenario | ScenarioInputs | User-provided scenario adjustments |

## State Transitions

None. All entities are computed values with no lifecycle or persistence.

## Relationships

```
BalanceResponse ──┐
AllowanceSchedule ├── ProjectionConfig ── calculateProjection() ── ProjectionResult
InterestSchedule ──┘                                                    │
ScenarioInputs ────┘                                              ProjectionDataPoint[]
```
