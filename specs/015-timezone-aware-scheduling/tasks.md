# Tasks: Timezone-Aware Scheduling

**Input**: Design documents from `/specs/015-timezone-aware-scheduling/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included per constitution (Test-First Development mandate). Core calculation tests written first.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Web app**: `backend/` (Go), `frontend/` (TypeScript/React)

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Core schedule calculation changes and store-layer modifications that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 Write timezone-aware unit tests for all schedule calculation functions (CalculateNextRun, CalculateNextRunAfterExecution, nextWeeklyDate, nextBiweeklyDate, nextMonthlyDate) with multiple timezones (UTC, America/New_York, America/Los_Angeles, Asia/Kolkata) and DST transition cases. Tests MUST FAIL before T002 implementation. File: `backend/internal/allowance/schedule_calc_test.go`

- [x] T002 Modify all schedule calculation functions to accept a `*time.Location` parameter instead of hardcoded `time.UTC`. Change `CalculateNextRun(sched, after, loc)`, `CalculateNextRunAfterExecution(sched, executedAt, loc)`, `nextWeeklyDate(dayOfWeek, after, loc)`, `nextBiweeklyDate(dayOfWeek, after, loc)`, `nextMonthlyDate(dayOfMonth, after, loc)`. Replace `after.UTC()` with `after.In(loc)` and `time.Date(..., time.UTC)` with `time.Date(..., loc)`. Verify T001 tests pass. File: `backend/internal/allowance/schedule_calc.go`

- [x] T003 Add `DueAllowanceSchedule` struct (embeds `AllowanceSchedule` + `FamilyTimezone string`). Modify `ListDue(now time.Time)` to return `[]DueAllowanceSchedule` by JOINing `allowance_schedules` with `children` and `families` tables to include `families.timezone`. Add `ListAllActiveWithTimezone()` method (same JOIN, no `next_run_at <= $1` filter) for startup recalculation. File: `backend/internal/store/schedule.go`

- [x] T004 [P] Add `DueInterestSchedule` struct (embeds `InterestSchedule` + `FamilyTimezone string`). Modify `ListDue(now time.Time)` to return `[]DueInterestSchedule` by JOINing `interest_schedules` with `children` and `families` tables to include `families.timezone`. Add `ListAllActiveWithTimezone()` method for startup recalculation. File: `backend/internal/store/interest_schedule.go`

**Checkpoint**: Core calculation is timezone-aware, store layer returns timezone with due schedules. All existing callers will need updating (Phase 2).

---

## Phase 2: User Story 1 — Scheduled Payments Happen at the Right Time (Priority: P1) 🎯 MVP

**Goal**: Allowance and interest payments fire at midnight in the family's configured timezone, not midnight UTC

**Independent Test**: Configure a family with timezone "America/New_York", set a weekly allowance for Wednesday, verify `next_run_at` is stored as `Wednesday 05:00:00 UTC` (midnight EST), not `Wednesday 00:00:00 UTC`

### Implementation for User Story 1

- [x] T005 [US1] Update `allowance.Scheduler` to accept `familyStore *store.FamilyStore` as a constructor dependency. In `ProcessDueSchedules`, update `ListDue` call to handle new `DueAllowanceSchedule` return type. In `executeSchedule`, use `sched.FamilyTimezone` to create `*time.Location` via `time.LoadLocation`, pass to `CalculateNextRunAfterExecution`. Fall back to `time.UTC` if timezone is empty or invalid. File: `backend/internal/allowance/scheduler.go`

- [x] T006 [P] [US1] Update `interest.Scheduler` to accept `familyStore *store.FamilyStore` via `SetFamilyStore` setter (matching existing `SetInterestScheduleStore` pattern). In `ProcessDueSchedules`, update `ListDue` call to handle new `DueInterestSchedule` return type. In `advanceNextRun`, use `sched.FamilyTimezone` to create `*time.Location`, pass to `CalculateNextRunAfterExecution`. File: `backend/internal/interest/scheduler.go`

- [x] T007 [US1] Add `RecalculateAllNextRuns()` method to `allowance.Scheduler`. This method calls `ListAllActiveWithTimezone()`, iterates all active schedules, loads `*time.Location` from `FamilyTimezone`, and recalculates `next_run_at` using `CalculateNextRun(sched, now, loc)`. Called from `Start()` before the ticker loop (one-time correction on startup). File: `backend/internal/allowance/scheduler.go`

- [x] T008 [P] [US1] Add `RecalculateAllNextRuns()` method to `interest.Scheduler`. Same pattern as allowance — calls `ListAllActiveWithTimezone()`, recalculates all active interest schedules using family timezone. Called from `Start()` before the ticker loop. File: `backend/internal/interest/scheduler.go`

- [x] T009 [US1] Update `allowance.Handler` to accept `familyStore *store.FamilyStore` as a constructor dependency. In `HandleCreateSchedule`, `HandleUpdateSchedule`, `HandleResumeSchedule`, `HandleSetChildAllowance`, and `HandleResumeChildAllowance`: look up timezone via `h.familyStore.GetTimezone(familyID)`, create `*time.Location` via `time.LoadLocation`, pass to `CalculateNextRun(sched, time.Now().UTC(), loc)`. Fall back to `time.UTC` if timezone lookup fails. File: `backend/internal/allowance/handler.go`

- [x] T010 [P] [US1] Update `interest.Handler` to accept `familyStore *store.FamilyStore` as a constructor dependency. In `HandleSetInterest`: look up timezone via `h.familyStore.GetTimezone(familyID)`, create `*time.Location`, pass to `calculateInterestNextRun`. Update `calculateInterestNextRun` to accept and forward `*time.Location` to `allowance.CalculateNextRun`. File: `backend/internal/interest/handler.go`

- [x] T011 [US1] Wire `familyStore` to all modified constructors in `main.go`. Pass `familyStore` to `allowance.NewHandler(scheduleStore, childStore, familyStore)`, `interest.NewHandler(interestStore, childStore, interestScheduleStore, familyStore)`, `allowance.NewScheduler(scheduleStore, txStore, childStore, familyStore)`. Call `interestScheduler.SetFamilyStore(familyStore)`. File: `backend/main.go`

- [x] T012 [US1] Run all backend tests to verify US1 changes compile and pass. Fix any test failures caused by changed function signatures (existing tests may need `time.UTC` passed as the new `loc` parameter). Command: `cd backend && go test -p 1 ./...`

**Checkpoint**: Backend scheduling is fully timezone-aware. Payments will fire at midnight in the family's timezone. Existing `next_run_at` values are corrected on startup.

---

## Phase 3: User Story 2 — Upcoming Dates Display Correctly (Priority: P1)

**Goal**: "Upcoming" allowance and interest dates display the correct calendar date in the family's timezone

**Independent Test**: Set family timezone to "America/New_York", create a Wednesday weekly allowance, verify the frontend shows the upcoming date as Wednesday (not Tuesday, which would happen if displayed using UTC)

### Implementation for User Story 2

- [x] T013 [US2] Create `TimezoneContext.tsx` with `TimezoneProvider` component and `useTimezone()` hook. Provider fetches timezone from `getSettings()` API on mount, provides `{ timezone: string, loading: boolean }`. Default timezone is `"UTC"` while loading. Handle auth-required case (only fetch when user is logged in). File: `frontend/src/context/TimezoneContext.tsx`

- [x] T014 [US2] Wrap the app root component with `TimezoneProvider` so all child components can access the family timezone via `useTimezone()`. File: `frontend/src/App.tsx` (or equivalent root — find the top-level component that renders routes)

- [x] T015 [US2] Update `TransactionsCard.tsx` upcoming payment date formatting to use family timezone. Import `useTimezone`, call it in the component, pass `{ timeZone: timezone }` to `toLocaleDateString` in the upcoming payments section (line ~193). File: `frontend/src/components/TransactionsCard.tsx`

- [x] T016 [P] [US2] Update `ChildAllowanceForm.tsx` `formatNextRun` function to accept timezone parameter and pass `{ timeZone: timezone }` to `toLocaleDateString`. Import and use `useTimezone()` hook. File: `frontend/src/components/ChildAllowanceForm.tsx`

- [x] T017 [P] [US2] Update `InterestForm.tsx` `formatNextRun` function to accept timezone parameter and pass `{ timeZone: timezone }` to `toLocaleDateString`. Import and use `useTimezone()` hook. File: `frontend/src/components/InterestForm.tsx`

**Checkpoint**: All "upcoming" and "next run" dates display in the family's timezone. Combined with US1 backend changes, displayed dates now match the actual payment schedule day.

---

## Phase 4: User Story 3 — All Date Displays Are Timezone-Consistent (Priority: P2)

**Goal**: Transaction dates and all other displayed dates use the family's timezone for consistency

**Independent Test**: Create a transaction at 11pm Pacific time on March 15, verify it displays as "Mar 15" in a Pacific-timezone family (not "Mar 16" which UTC would show)

### Implementation for User Story 3

- [x] T018 [US3] Update `TransactionsCard.tsx` `formatRecentDate` function to accept timezone parameter and pass `{ timeZone: timezone }` to `toLocaleDateString`. Use `useTimezone()` hook value (already imported in T015). File: `frontend/src/components/TransactionsCard.tsx`

**Checkpoint**: All dates across the application — upcoming schedules, next run dates, and transaction history — consistently display in the family's configured timezone.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final verification and cleanup

- [x] T019 Run full backend test suite and fix any failures. Command: `cd backend && go test -p 1 ./...`
- [x] T020 [P] Run frontend lint and fix any issues. Command: `cd frontend && npm run lint`
- [x] T021 Verify edge cases: test with a family whose timezone is empty/null (should fall back to UTC), test DST transition dates in schedule_calc_test.go, test non-hour-offset timezone (Asia/Kolkata). File: `backend/internal/allowance/schedule_calc_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately. BLOCKS all user stories.
- **US1 (Phase 2)**: Depends on Phase 1 completion. Core backend changes.
- **US2 (Phase 3)**: Backend portion satisfied by US1 (Phase 2). Frontend context is independent of backend. Can START Phase 3 in parallel with Phase 2 (T013-T014 don't depend on backend), but T015-T017 benefit from correct backend values.
- **US3 (Phase 4)**: Depends on T013 (TimezoneContext created in US2). One additional frontend change.
- **Polish (Phase 5)**: Depends on all phases complete.

### User Story Dependencies

- **US1 (P1)**: Depends only on Foundational phase. No dependencies on other stories.
- **US2 (P1)**: Frontend context (T013-T014) can start after Foundational. Date formatting (T015-T017) should follow US1 for correct backend values, but is independently testable.
- **US3 (P2)**: Depends on US2 (TimezoneContext must exist). Single task (T018).

### Within Each User Story

- Tests written FIRST (T001 before T002)
- Store changes before scheduler/handler changes (T003-T004 before T005-T010)
- Scheduler/handler changes before wiring (T005-T010 before T011)
- Context creation before component usage (T013 before T015-T017)

### Parallel Opportunities

**Phase 1:**
- T003 and T004 can run in parallel (different files: schedule.go vs interest_schedule.go)

**Phase 2:**
- T005 and T006 can run in parallel (different files: allowance/scheduler.go vs interest/scheduler.go)
- T007 and T008 can run in parallel (same files as above, but additive methods)
- T009 and T010 can run in parallel (different files: allowance/handler.go vs interest/handler.go)

**Phase 3:**
- T016 and T017 can run in parallel (different files: ChildAllowanceForm.tsx vs InterestForm.tsx)

---

## Parallel Example: User Story 1

```bash
# After Phase 1 completes, launch scheduler changes in parallel:
Task T005: "Update allowance.Scheduler in backend/internal/allowance/scheduler.go"
Task T006: "Update interest.Scheduler in backend/internal/interest/scheduler.go"

# Then launch handler changes in parallel:
Task T009: "Update allowance.Handler in backend/internal/allowance/handler.go"
Task T010: "Update interest.Handler in backend/internal/interest/handler.go"
```

## Parallel Example: User Story 2

```bash
# After TimezoneContext created (T013-T014), launch component updates in parallel:
Task T016: "Update ChildAllowanceForm.tsx to use timezone"
Task T017: "Update InterestForm.tsx to use timezone"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Foundational (T001-T004)
2. Complete Phase 2: US1 — Backend scheduling (T005-T012)
3. **STOP and VALIDATE**: Verify `next_run_at` values are timezone-correct in database
4. Payments now fire at the correct time — core bug is fixed

### Incremental Delivery

1. Phase 1 + Phase 2 → Backend scheduling correct (MVP!)
2. Add Phase 3 → Upcoming dates display correctly in frontend
3. Add Phase 4 → All dates consistent in frontend
4. Each phase adds user-facing value without breaking previous

### Single Developer (Recommended Path)

1. T001 → T002 → T003 → T004 (Foundational, sequential + parallel where marked)
2. T005+T006 → T007+T008 → T009+T010 → T011 → T012 (US1)
3. T013 → T014 → T015+T016+T017 (US2)
4. T018 (US3)
5. T019+T020 → T021 (Polish)

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks in same phase
- [Story] label maps task to specific user story for traceability
- US1 and US2 share backend infrastructure but have distinct frontend needs
- No database migrations needed — all changes are behavioral
- Startup recalculation (T007-T008) ensures existing data is corrected on deploy
- Fallback to UTC ensures system works even if timezone is missing
- Commit after each task or logical group
