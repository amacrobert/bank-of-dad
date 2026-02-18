# API Contracts: Savings Growth Projector

**Feature**: 016-savings-growth-projector
**Date**: 2026-02-17

## No New API Endpoints

This feature requires **no new backend endpoints**. All projection calculations happen client-side using data from existing endpoints.

## Existing Endpoints Used

### GET /api/children/{childId}/balance

Returns current balance and interest rate configuration.

**Response** (existing):
```json
{
  "child_id": 1,
  "first_name": "Alice",
  "balance_cents": 10050,
  "interest_rate_bps": 500,
  "interest_rate_display": "5.00%",
  "next_interest_at": "2026-02-24T00:00:00Z"
}
```

**Used for**: Starting balance (`balance_cents`) and interest rate (`interest_rate_bps`).

### GET /api/children/{childId}/allowance

Returns the child's allowance schedule (or 404 if none exists).

**Response** (existing):
```json
{
  "id": 1,
  "child_id": 1,
  "parent_id": 1,
  "amount_cents": 1000,
  "frequency": "weekly",
  "day_of_week": 6,
  "status": "active",
  "next_run_at": "2026-02-22T00:00:00Z",
  "created_at": "2026-01-15T12:00:00Z",
  "updated_at": "2026-01-15T12:00:00Z"
}
```

**Used for**: Allowance amount (`amount_cents`), frequency, and active/paused status.

### GET /api/children/{childId}/interest-schedule

Returns the child's interest compounding schedule (or 404 if none exists).

**Response** (existing):
```json
{
  "id": 1,
  "child_id": 1,
  "parent_id": 1,
  "frequency": "weekly",
  "day_of_week": 0,
  "status": "active",
  "next_run_at": "2026-02-23T00:00:00Z",
  "created_at": "2026-01-15T12:00:00Z",
  "updated_at": "2026-01-15T12:00:00Z"
}
```

**Used for**: Interest compounding frequency and active/paused status.

## Client-Side Contract: Projection Engine

### Function: calculateProjection

**Input**: `ProjectionConfig` (see data-model.md)

**Output**: `ProjectionResult` (see data-model.md)

**Behavior**:
1. Adjust starting balance by one-time deposit/withdrawal from scenario
2. Iterate week-by-week through the time horizon
3. At each week, apply any due allowance deposits (based on frequency)
4. At each week, apply any due interest (based on compounding frequency and rate)
5. Subtract weekly spending from scenario
6. Floor balance at $0 (no negatives)
7. Track cumulative interest, allowance, and spending totals
8. Record depletion week if balance reaches $0
9. Return time series + summary breakdown
