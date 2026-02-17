# Tasks: Combine Transaction Cards

**Input**: Design documents from `/specs/014-combine-transactions/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No frontend test framework in place. Manual testing only per acceptance scenarios.

**Organization**: Tasks are grouped by user story. US1 and US2 share the new component (TransactionsCard), so creation of the shared component is in a foundational phase. Integration into each dashboard is per-story and parallelizable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Foundational (Shared Component)

**Purpose**: Create the new combined TransactionsCard component that both dashboards will use

**⚠️ CRITICAL**: Both US1 and US2 depend on this component existing before integration

- [x] T001 Create `TransactionsCard` component in `frontend/src/components/TransactionsCard.tsx` that combines the data-fetching logic from `UpcomingPayments.tsx` (upcoming allowances + interest schedule fetching, `UpcomingPayment` interface, `periodsPerYear`, `estimateInterestCents`, `formatAmount` helpers) with the rendering logic from `TransactionHistory.tsx` (`typeConfig` icon/color mapping, `formatDate`, `formatAmount`, `getTypeLabel` helpers). The component should: (a) accept props `{ childId: number; balanceCents: number; interestRateBps: number; transactions: Transaction[] }`, (b) wrap everything in a single `<Card padding="md">` with title "Transactions", (c) render an "Upcoming" sub-heading section with upcoming payment entries using the existing badge/icon styling from `UpcomingPayments.tsx`, showing "No upcoming payments" when empty, (d) render a "Recent" sub-heading section with recent transaction entries using the existing row styling from `TransactionHistory.tsx`, showing "No transactions yet." when empty, (e) show a loading spinner while upcoming data is being fetched, (f) preserve all existing formatting: upcoming dates as "MMM D", recent dates as "MMM D, YYYY", upcoming amounts with `~` prefix for estimates, recent amounts with `+/-` prefix by type.

**Checkpoint**: TransactionsCard component is ready for integration into both dashboards

---

## Phase 2: User Story 1 - Parent Views Combined Transactions (Priority: P1) 🎯 MVP

**Goal**: Parent dashboard shows a single Transactions card when a child is selected

**Independent Test**: Log in as parent → select a child → verify single "Transactions" card with "Upcoming" and "Recent" sections displaying correct data

### Implementation for User Story 1

- [x] T002 [US1] Update `frontend/src/components/ManageChild.tsx` to replace `UpcomingPayments` and `TransactionHistory` with the new `TransactionsCard` component: (a) remove imports of `UpcomingPayments` and `TransactionHistory`, (b) add import for `TransactionsCard`, (c) replace lines 205-211 (the `<UpcomingPayments ... />` and the `<Card>` wrapping `<TransactionHistory>`) with a single `<TransactionsCard childId={child.id} balanceCents={currentBalance} interestRateBps={interestRateBps} transactions={transactions} />`

**Checkpoint**: Parent dashboard displays the combined Transactions card. Verify all 4 acceptance scenarios (both sections have data, only upcoming, only recent, neither).

---

## Phase 3: User Story 2 - Child Views Combined Transactions (Priority: P1)

**Goal**: Child dashboard shows a single Transactions card

**Independent Test**: Log in as child → verify single "Transactions" card with "Upcoming" and "Recent" sections displaying correct data

### Implementation for User Story 2

- [x] T003 [P] [US2] Update `frontend/src/pages/ChildDashboard.tsx` to replace `UpcomingPayments` and `TransactionHistory` with the new `TransactionsCard` component: (a) remove imports of `UpcomingPayments` and `TransactionHistory`, (b) remove import of `Card` if no longer used elsewhere in the file (check: it's still used for the hero balance card, so keep it), (c) add import for `TransactionsCard`, (d) replace lines 89-106 (the `<UpcomingPayments ... />` block and the `<Card>` wrapping "Recent Activity" heading + `<TransactionHistory>`) with a single `<TransactionsCard childId={user.user_id} balanceCents={balance} interestRateBps={interestRateBps} transactions={transactions} />`, wrapping it in the same `{!loadingData && ( ... )}` conditional as the current `UpcomingPayments`

**Checkpoint**: Child dashboard displays the combined Transactions card. Verify both acceptance scenarios (has data, empty state).

---

## Phase 4: User Story 3 - Remove Obsolete Components (Priority: P2)

**Goal**: Delete old separate card components and ensure no orphaned code remains

**Independent Test**: Verify that old component files are deleted, no imports reference them, and both dashboards still render correctly

### Implementation for User Story 3

- [x] T004 [P] [US3] Delete `frontend/src/components/UpcomingPayments.tsx`
- [x] T005 [P] [US3] Delete `frontend/src/components/TransactionHistory.tsx`
- [x] T006 [US3] Search the entire codebase for any remaining references to `UpcomingPayments` or `TransactionHistory` (imports, comments, etc.) and remove them. Verify the application builds without errors.

**Checkpoint**: No orphaned code remains. Application compiles and runs successfully.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — can start immediately
- **US1 (Phase 2)**: Depends on T001 (TransactionsCard must exist)
- **US2 (Phase 3)**: Depends on T001 (TransactionsCard must exist) — can run in parallel with US1
- **US3 (Phase 4)**: Depends on T002 and T003 (both dashboards must be updated before deleting old components)

### Parallel Opportunities

- **T002 and T003** can run in parallel (different files: `ManageChild.tsx` vs `ChildDashboard.tsx`)
- **T004 and T005** can run in parallel (independent file deletions)

### Within Each User Story

- US1: Single task (T002) — straightforward import swap
- US2: Single task (T003) — straightforward import swap
- US3: Delete files first (T004, T005 in parallel), then verify (T006)

---

## Parallel Example: US1 + US2

```bash
# After T001 completes, launch both integrations in parallel:
Task: "Update ManageChild.tsx to use TransactionsCard"         # T002 [US1]
Task: "Update ChildDashboard.tsx to use TransactionsCard"      # T003 [US2]
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create TransactionsCard (T001)
2. Complete Phase 2: Integrate into ManageChild.tsx (T002)
3. **STOP and VALIDATE**: Test parent dashboard independently
4. Proceed to US2 and US3

### Incremental Delivery

1. T001 → TransactionsCard exists
2. T002 + T003 (parallel) → Both dashboards updated
3. T004 + T005 + T006 → Old code removed
4. Final manual verification against all acceptance scenarios

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- No backend changes required — all existing API endpoints remain unchanged
- The `TransactionsCard` component reuses existing types (`Transaction`, `UpcomingAllowance`, `InterestSchedule`) from `types.ts` and existing API functions from `api.ts`
- One behavioral change: the current `UpcomingPayments` returns `null` when there are no upcoming payments (hides entirely). The new combined card always shows the "Upcoming" section with an empty state message per FR-005.
- Commit after each task or logical group
