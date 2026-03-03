# Research: Savings Goals

**Branch**: `025-savings-goals` | **Date**: 2026-03-02

## R-001: Fund Reservation Model

**Decision**: Reserved fund model — allocating money to a goal reduces the child's "available balance" while keeping the total balance (available + saved) unchanged.

**Rationale**: This teaches real budgeting and makes goal progress feel tangible. The existing `balance_cents` column on `children` continues to represent the total balance. A new computed value ("available balance" = total balance − sum of all active goal allocations) provides the split. Interest is calculated on total balance, not available balance.

**Alternatives considered**:
- Visual tracker only: Simpler but goals feel hollow — no real commitment of funds.
- Hybrid soft earmark: Advisory-only earmarking confuses the mental model — either funds are reserved or they aren't.

## R-002: Goal Allocation Storage

**Decision**: Store goal allocations as individual events in a `goal_allocations` table. Each allocation (or de-allocation) is a separate row recording amount, direction, and timestamp. The goal's `saved_cents` column is updated atomically with each allocation for fast reads.

**Rationale**: Preserves an audit trail of all allocation activity (FR-013). The denormalized `saved_cents` on the goal avoids summing allocations on every read. Atomic DB transactions ensure consistency between the allocation record and the goal's saved amount.

**Alternatives considered**:
- Sum allocations on read: Accurate but slow for frequent reads (dashboard loads).
- Only store current amount (no history): Loses the allocation history required by FR-013.

## R-003: Parent Withdrawal Impact on Goals

**Decision**: When a parent withdrawal would reduce the total balance below the sum of all goal allocations, the system warns the parent with a confirmation step. If confirmed, goal allocations are reduced proportionally across all active goals to fit the new balance.

**Rationale**: Parents have ultimate authority over the money. Blocking withdrawals would undermine parent control. Proportional reduction is fairer than draining a single goal.

**Alternatives considered**:
- Block withdrawal: Too restrictive on parent authority.
- Silent auto-reduce: Parent should know when they're impacting a child's goals.

## R-004: Available Balance Calculation

**Decision**: Available balance is computed as `children.balance_cents - SUM(savings_goals.saved_cents WHERE status = 'active')`. This is calculated on-read, not stored separately, to avoid dual-source-of-truth issues.

**Rationale**: The existing `balance_cents` field continues to be the authoritative total balance. Storing a separate `available_balance_cents` would create synchronization risk. The query is fast because a child has at most 5 active goals.

**Alternatives considered**:
- Store `available_balance_cents` column: Adds sync risk between two columns. Not worth it for a MAX 5-row SUM.

## R-005: Celebration Animation Approach

**Decision**: Use a lightweight CSS/JS confetti animation triggered on the frontend when a goal completion is detected. No backend involvement beyond marking the goal as completed.

**Rationale**: The celebration is a UI concern. The backend returns the goal's new status; the frontend detects the transition to "completed" and triggers the animation. This keeps the backend simple and the animation customizable.

**Alternatives considered**:
- Backend push notification: Over-engineered for a single-user session feature.
- Third-party animation library: Unnecessary dependency for a simple confetti effect.

## R-006: Existing Pattern Alignment

**Decision**: Follow existing project patterns exactly:
- New store file: `backend/internal/store/savings_goal.go` with `SavingsGoalStore`
- New handler package: `backend/internal/goals/handler.go` (mirrors `balance/`, `allowance/`, `interest/`)
- New migration: `backend/migrations/007_savings_goals.{up,down}.sql`
- Frontend types in `types.ts`, API functions in `api.ts`, new page components

**Rationale**: Consistency with existing patterns reduces cognitive load and maintenance burden. No new architectural concepts needed.
