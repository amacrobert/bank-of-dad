# Tasks: Account Management Enhancements

**Input**: Design documents from `/specs/006-account-management-enhancements/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md, quickstart.md

**Tests**: Included — constitution requires Test-First Development for all financial logic.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database migrations and shared data layer changes

- [ ] T001 Add `interest_schedules` table migration in `backend/internal/store/sqlite.go`
- [ ] T002 Add UNIQUE constraint migration for `allowance_schedules.child_id` (with deduplication of existing data) in `backend/internal/store/sqlite.go`
- [ ] T003 Add `GetByChildID(childID)` method to `ScheduleStore` in `backend/internal/store/schedule.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: InterestScheduleStore and modified ApplyInterest — MUST be complete before user stories

**CRITICAL**: No user story work can begin until this phase is complete

### Tests

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T004 [P] Write store tests for `InterestScheduleStore` CRUD (Create, GetByChildID, Update, Delete) in `backend/internal/store/interest_schedule_test.go`
- [ ] T005 [P] Write store tests for `InterestScheduleStore.ListDue` and `UpdateNextRunAt` in `backend/internal/store/interest_schedule_test.go`
- [ ] T006 [P] Write store tests for `ApplyInterest` with `periodsPerYear` param (weekly=52, biweekly=26, monthly=12 proration) in `backend/internal/store/interest_test.go`
- [ ] T007 [P] Write store test for `ScheduleStore.GetByChildID` in `backend/internal/store/schedule_test.go`

### Implementation

- [ ] T008 Implement `InterestScheduleStore` with Create, GetByChildID, Update, Delete, ListDue, UpdateNextRunAt, UpdateStatus in `backend/internal/store/interest_schedule.go`
- [ ] T009 Modify `ApplyInterest` to accept `periodsPerYear int` parameter (replace hardcoded `/12`) in `backend/internal/store/interest.go`
- [ ] T010 Run foundational tests: `cd backend && go test ./internal/store/... -v`

**Checkpoint**: Store layer complete — all interest schedule and modified interest operations work correctly

---

## Phase 3: User Story 1 — Parent Views Child Transaction History (Priority: P1) MVP

**Goal**: Parents can see a child's full transaction history from the manage child view

**Independent Test**: Open manage child, see chronological transaction list with all types

### Implementation

- [ ] T011 [US1] Fetch and display transaction history in ManageChild: call `getTransactions(childId)` and render `TransactionHistory` component in `frontend/src/components/ManageChild.tsx`
- [ ] T012 [US1] Verify transaction history displays correctly for parent view (manual test or visual check)

**Checkpoint**: Parents see transaction history when managing a child

---

## Phase 4: User Story 2 — Interest Rate Form Pre-population (Priority: P1)

**Goal**: Interest rate form shows the current rate when opening manage child, not a stale default

**Independent Test**: Set rate to 5%, close, reopen — form shows "5.00"

### Implementation

- [ ] T013 [US2] Fix `InterestRateForm` to re-initialize when `currentRateBps` prop changes asynchronously (use `key` prop or `useEffect` sync) in `frontend/src/components/InterestRateForm.tsx`
- [ ] T014 [US2] Verify pre-population works: set rate, close manage child, reopen, confirm form shows saved rate

**Checkpoint**: Interest rate is always pre-populated correctly

---

## Phase 5: User Story 3 — Unified Child Management with Single Allowance (Priority: P1)

**Goal**: Each child has at most one allowance, managed inline within the manage child form. Standalone allowance section removed.

**Independent Test**: Create/edit/pause/remove allowance from manage child view; standalone section gone

### Tests

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T015 [P] [US3] Write handler tests for child-scoped allowance endpoints: `GET /api/children/{childId}/allowance`, `PUT /api/children/{childId}/allowance`, `DELETE /api/children/{childId}/allowance`, pause, resume in `backend/internal/allowance/handler_test.go`

### Implementation

- [ ] T016 [US3] Implement child-scoped allowance handlers: HandleGetChildAllowance, HandleSetChildAllowance (create-or-update), HandleDeleteChildAllowance, HandlePauseChildAllowance, HandleResumeChildAllowance in `backend/internal/allowance/handler.go`
- [ ] T017 [US3] Wire child-scoped allowance routes in `backend/main.go`
- [ ] T018 [US3] Run allowance handler tests: `cd backend && go test ./internal/allowance/... -v`
- [ ] T019 [US3] Add `InterestSchedule` type and child-scoped allowance API functions to `frontend/src/types.ts` and `frontend/src/api.ts`
- [ ] T020 [US3] Create `ChildAllowanceForm` component (create/edit/pause/resume/remove allowance, pre-populated when exists) in `frontend/src/components/ChildAllowanceForm.tsx`
- [ ] T021 [US3] Integrate `ChildAllowanceForm` into `ManageChild.tsx` — fetch child's allowance on mount, pass to form
- [ ] T022 [US3] Remove standalone allowance section (`ScheduleForm`, `ScheduleList`, related state) from `frontend/src/pages/ParentDashboard.tsx`

**Checkpoint**: Allowance managed per-child inline; standalone section removed

---

## Phase 6: User Story 4 — Scheduled Interest Accrual (Priority: P2)

**Goal**: Parents configure when interest accrues using a schedule (weekly/biweekly/monthly), replacing fixed monthly accrual

**Independent Test**: Set interest schedule to monthly on 15th, verify scheduler credits interest on schedule

### Tests

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T023 [P] [US4] Write handler tests for interest schedule endpoints: `PUT /api/children/{childId}/interest-schedule`, `GET`, `DELETE` in `backend/internal/interest/handler_test.go`
- [ ] T024 [P] [US4] Write scheduler tests: ProcessDue uses interest_schedules table, applies correct proration, updates next_run_at in `backend/internal/interest/scheduler_test.go`

### Implementation

- [ ] T025 [US4] Implement interest schedule handlers: HandleSetInterestSchedule (create-or-update), HandleGetInterestSchedule, HandleDeleteInterestSchedule in `backend/internal/interest/handler.go`
- [ ] T026 [US4] Wire interest schedule routes in `backend/main.go`
- [ ] T027 [US4] Modify interest scheduler to use `InterestScheduleStore.ListDue()` instead of `InterestStore.ListDueForInterest()`, apply correct periodsPerYear, update next_run_at after accrual in `backend/internal/interest/scheduler.go`
- [ ] T028 [US4] Enhance balance handler to include `next_interest_at` field from interest schedule's `next_run_at` in `backend/internal/balance/handler.go`
- [ ] T029 [US4] Run interest handler and scheduler tests: `cd backend && go test ./internal/interest/... ./internal/balance/... -v`
- [ ] T030 [US4] Add interest schedule API functions to `frontend/src/api.ts`
- [ ] T031 [US4] Create `InterestScheduleForm` component (create/edit/delete schedule with frequency/day selection) in `frontend/src/components/InterestScheduleForm.tsx`
- [ ] T032 [US4] Integrate `InterestScheduleForm` into `ManageChild.tsx` — fetch interest schedule on mount, pass to form

**Checkpoint**: Interest accrues on parent-configured schedule

---

## Phase 7: User Story 5 — Child Sees Interest Rate and Next Payment (Priority: P3)

**Goal**: Children see their interest rate and next payment date on their dashboard

**Independent Test**: Child logs in, sees "5.00% annual interest" and "Next interest payment: March 15"

### Implementation

- [ ] T033 [US5] Update `BalanceResponse` type in `frontend/src/types.ts` to include `next_interest_at` field
- [ ] T034 [US5] Display interest rate and next payment date on child dashboard in `frontend/src/pages/ChildDashboard.tsx`

**Checkpoint**: Children see interest info on their dashboard

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Validation, cleanup, and final verification

- [ ] T035 Run full backend test suite: `cd backend && go test ./... -v`
- [ ] T036 Run frontend checks: `cd frontend && npx tsc --noEmit && npx vite build`
- [ ] T037 Run quickstart.md manual verification steps

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 (frontend-only — no backend changes needed)
- **US2 (Phase 4)**: Depends on Phase 2 (frontend-only — fix form pre-population)
- **US3 (Phase 5)**: Depends on Phase 2 (new backend endpoints + frontend component)
- **US4 (Phase 6)**: Depends on Phase 2 (new backend endpoints + scheduler changes + frontend)
- **US5 (Phase 7)**: Depends on US4 (needs next_interest_at from balance response)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Frontend-only — can start after Phase 2
- **US2 (P1)**: Frontend-only — can start after Phase 2, independent of US1
- **US3 (P1)**: Backend + Frontend — can start after Phase 2, independent of US1/US2
- **US4 (P2)**: Backend + Frontend — can start after Phase 2, independent of US1/US2/US3
- **US5 (P3)**: Frontend-only — depends on US4 (needs next_interest_at in balance response)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Store layer before handlers
- Handlers before route wiring
- Backend before frontend (for API-dependent UI)

### Parallel Opportunities

- T004, T005, T006, T007 can all be written in parallel (different test files)
- T015 can run in parallel with T023, T024 (different test files, different stories)
- US1 and US2 can proceed in parallel (both frontend-only, different components)
- US3 and US4 backend work can proceed in parallel (different handler/store files)

---

## Implementation Strategy

### MVP First (US1 + US2 + US3)

1. Complete Phase 1: Setup (migrations)
2. Complete Phase 2: Foundational (store layer with tests)
3. Complete Phase 3: US1 — Parent sees transaction history
4. Complete Phase 4: US2 — Interest rate pre-populated
5. Complete Phase 5: US3 — Unified allowance in manage child
6. **STOP and VALIDATE**: Test all P1 stories independently

### Incremental Delivery

1. Setup + Foundational → Data layer ready
2. US1 (Transaction history) → Parents see child activity
3. US2 (Interest rate form) → Correct pre-population
4. US3 (Single allowance) → Simplified allowance management
5. US4 (Scheduled interest) → Flexible interest scheduling
6. US5 (Child interest visibility) → Children informed
7. Polish → Full validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires Test-First Development — write tests before implementation
- Interest calculation uses integer arithmetic with periodsPerYear: `balance * rate / periodsPerYear / 10000`
- Use past dates in test fixtures to avoid time-dependent failures
- Follow existing patterns: store, handler, scheduler, schedule_calc
- Reuse `CalculateNextRun` and `CalculateNextRunAfterExecution` from `schedule_calc.go` for interest schedules
- Commit after each task or logical group
