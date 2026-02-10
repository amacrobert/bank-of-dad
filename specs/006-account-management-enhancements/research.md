# Research: Account Management Enhancements

## R1: Interest Accrual Scheduling — Reuse vs. New Table

**Decision**: Create a new `interest_schedules` table that mirrors the structure of `allowance_schedules`.

**Rationale**: The interest schedule has the same fields (frequency, day_of_week, day_of_month, next_run_at, status) but serves a fundamentally different purpose. Reusing `allowance_schedules` would overload that table and create confusion between deposit schedules and interest calculation schedules. A dedicated table is cleaner, follows single responsibility, and avoids complex type-discrimination columns.

**Alternatives considered**:
- Add a `schedule_type` column to `allowance_schedules`: Rejected — mixes two concerns, complicates queries and constraints, and the allowance scheduler would need to filter by type.
- Store schedule fields directly on `children` table: Rejected — would add 5+ columns to an already growing table.

## R2: Interest Proration for Non-Monthly Frequencies

**Decision**: Use annual rate / periods-per-year for proration. Weekly = rate/52, biweekly = rate/26, monthly = rate/12.

**Rationale**: Simple, standard approach. Matches user expectations — "5% annually, paid weekly" means 5%/52 per week. The existing `ApplyInterest` method divides by 12; it needs to accept a divisor parameter or the caller needs to pass the period count.

**Approach**: Modify `ApplyInterest` to accept a `periodsPerYear` int parameter. The scheduler determines this from the schedule's frequency.

## R3: One Allowance Per Child — Enforcement Strategy

**Decision**: Add a UNIQUE constraint on `child_id` in `allowance_schedules` and enforce in application code.

**Rationale**: Database constraint prevents races and ensures data integrity. Application-level check provides a friendly error message. Both layers work together.

**Migration concern**: If any existing children have multiple allowances, the UNIQUE constraint migration will fail. The migration must first check for duplicates and handle them (keep the most recently created one, delete others) or skip the constraint if duplicates exist. Given this is a personal/family app, duplicates are unlikely but must be handled safely.

**Approach**: Migration checks for duplicates first. If found, keeps the newest schedule per child and deletes others. Then adds the UNIQUE constraint.

## R4: Fetching Allowance Schedule for a Child

**Decision**: Add `GetByChildID(childID int64)` method to `ScheduleStore` to fetch a child's single allowance.

**Rationale**: The existing `ListActiveByChild` returns a slice. With the one-per-child constraint, we need a method that returns a single schedule (or nil). The new method queries without the `status = 'active'` filter since the UI should show paused schedules too — the parent needs to see the current state.

## R5: Parent Transaction History — Backend vs. Frontend Only

**Decision**: No backend changes needed. The existing `GET /api/children/{id}/transactions` endpoint already works for parents (returns all transactions for a child, requires auth and same-family check). The frontend just needs to call it from ManageChild.

**Rationale**: The balance handler already supports both parent and child access to transactions. The only change is adding the TransactionHistory component to the ManageChild view and fetching transaction data.

## R6: Interest Rate Pre-population — Data Flow

**Decision**: Fetch interest rate from `GET /api/children/{id}/balance` endpoint, which already returns `interest_rate_bps`. Use this to populate the InterestRateForm on mount.

**Rationale**: The balance endpoint already includes interest rate data (added in 005). No new endpoints needed. The InterestRateForm component already receives `currentRateBps` as a prop — we just need to ensure ManageChild fetches and passes it correctly.

**Current issue**: ManageChild already fetches getBalance and passes interestRateBps to InterestRateForm, but the form uses `useState` initialized from the prop, so if the prop loads async, the form may show "0.00" initially then not update. Need to ensure the form re-initializes when the prop changes (use `key` prop or `useEffect`).

## R7: Interest Schedule — Child Dashboard Visibility

**Decision**: Add a new backend endpoint `GET /api/children/{childId}/upcoming-interest` that returns the next interest payment date, or extend the balance response to include `next_interest_at`.

**Rationale**: The child dashboard already displays `UpcomingAllowances`. For interest, we need the next scheduled date. The simplest approach is to include `next_interest_at` in the balance response (already fetched by ChildDashboard), since it's a single date from the interest schedule's `next_run_at`.

**Decision**: Extend `BalanceResponse` to include `next_interest_at` field, populated from the interest schedule's `next_run_at`.

## R8: Removing Standalone Allowance Section

**Decision**: Remove `ScheduleForm` and `ScheduleList` from `ParentDashboard.tsx`. Replace with inline allowance management within `ManageChild.tsx`.

**Rationale**: With one allowance per child, a family-wide schedule list is unnecessary. Each child's allowance is managed individually. The standalone components can be removed or the logic reused within a new `ChildAllowanceForm` component.

**Approach**: Create a new `ChildAllowanceForm` component that combines schedule creation, editing, pause/resume, and deletion in a single form within ManageChild. The standalone `ScheduleForm`, `ScheduleList`, and their section in ParentDashboard will be removed.
