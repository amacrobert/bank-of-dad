# Tasks: Allowance Scheduling

**Input**: Design documents from `/specs/003-allowance-scheduling/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/api.yaml, research.md, quickstart.md

**Tests**: Included per constitution requirement (Test-First Development). Write tests FIRST, confirm they FAIL, then implement.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database migration and shared type definitions needed by all user stories

- [x] T001 Run database migration: create `allowance_schedules` table and add `schedule_id` column to `transactions` table per `specs/003-allowance-scheduling/data-model.md` migration script in `backend/internal/store/db.go` (or existing migration location)
- [x] T002 [P] Add `AllowanceSchedule` model, `Frequency` type, `ScheduleStatus` type, and response types to `backend/internal/store/schedule.go` per data-model.md Go model definitions
- [x] T003 [P] Add `TransactionTypeAllowance` constant and `ScheduleID` field to existing Transaction model in `backend/internal/store/transaction.go`
- [x] T004 [P] Add TypeScript types (`Frequency`, `ScheduleStatus`, `AllowanceSchedule`, `ScheduleWithChild`, `ScheduleListResponse`, `CreateScheduleRequest`, `UpdateScheduleRequest`, `UpcomingAllowance`, `UpcomingAllowancesResponse`) to `frontend/src/types.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Date calculation logic and store layer that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational Phase

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T005 [P] Write tests for `nextWeeklyDate`, `nextBiweeklyDate`, `nextMonthlyDate`, `daysInMonth`, and `calculateNextRun` helper functions in `backend/internal/allowance/schedule_calc_test.go` — cover: normal cases, end-of-month clamping (31st in Feb, Apr, Jun, Sep, Nov), leap year Feb 29, same-day-as-today (should go to next occurrence), biweekly 14-day cycle
- [x] T006 [P] Write tests for `ScheduleStore` CRUD operations (Create, GetByID, ListByParentFamily, Update, Delete, UpdateNextRunAt, UpdateStatus) in `backend/internal/store/schedule_test.go` — verify next_run_at is calculated on Create, verify all fields round-trip correctly, verify ListByParentFamily joins child_first_name

### Implementation for Foundational Phase

- [x] T007 Implement date calculation helper functions (`calculateNextRun`, `nextWeeklyDate`, `nextBiweeklyDate`, `nextMonthlyDate`, `daysInMonth`) in `backend/internal/allowance/schedule_calc.go` per research.md Decision 4 and quickstart.md patterns
- [x] T008 Implement `ScheduleStore` with methods: `Create`, `GetByID`, `ListByParentFamily`, `Update`, `Delete`, `ListDue`, `UpdateNextRunAt`, `UpdateStatus` in `backend/internal/store/schedule.go` — use existing read/write DB connection pattern per quickstart.md store pattern
- [x] T009 Run foundational tests (T005, T006) — all must pass green

**Checkpoint**: Foundation ready — date math and store CRUD verified. User story implementation can begin.

---

## Phase 3: User Story 1 — Parent Sets Up Weekly Allowance (Priority: P1) 🎯 MVP

**Goal**: Parent can create a weekly allowance schedule for a child. Background scheduler automatically deposits the amount on the scheduled day. Both parent and child see the transaction in history.

**Independent Test**: Create a weekly $10 Friday schedule → scheduler processes it → transaction appears with type "allowance" and schedule_id set.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T010 [P] [US1] Write handler test for `POST /api/schedules` (create weekly schedule) in `backend/internal/allowance/handler_test.go` — test: valid create returns 201 with schedule JSON including next_run_at; invalid frequency returns 400; missing day_of_week for weekly returns 400; child not in parent's family returns 404; unauthenticated returns 401; child role returns 403
- [x] T011 [P] [US1] Write scheduler tests in `backend/internal/allowance/scheduler_test.go` — test: `ProcessDueSchedules` creates transaction with type "allowance" and correct schedule_id; updates child balance; advances next_run_at to next week; skips paused schedules; processes multiple due schedules; handles missed schedules (next_run_at in the past)

### Implementation for User Story 1

- [x] T012 [US1] Create `backend/internal/allowance/handler.go` with `Handler` struct, `NewHandler` constructor, and `HandleCreateSchedule` method — validate request body (amount_cents 1-99999999, frequency enum, day_of_week 0-6 for weekly), verify parent session, verify child belongs to parent's family, call ScheduleStore.Create, return 201 with JSON
- [x] T013 [US1] Create `backend/internal/allowance/scheduler.go` with `Scheduler` struct, `NewScheduler` constructor, `Start` method (goroutine with ticker + stop channel per session cleanup pattern), `ProcessDueSchedules` method (query ListDue, execute each), and `executeSchedule` method (create "allowance" transaction with schedule_id, update child balance, calculate and set next next_run_at)
- [x] T014 [US1] Register `POST /api/schedules` route and start scheduler goroutine in `backend/main.go` — create ScheduleStore, create allowance.Handler, create allowance.Scheduler, wire route with auth middleware, start scheduler with 1-minute interval and stop channel
- [x] T015 [P] [US1] Add `createSchedule` API function to `frontend/src/api.ts` — POST to `/api/schedules` with CreateScheduleRequest body, return AllowanceSchedule
- [x] T016 [US1] Create `frontend/src/components/ScheduleForm.tsx` — form with child selector dropdown, amount input (dollars, converted to cents), frequency radio buttons (weekly/biweekly/monthly), day-of-week selector (shown for weekly), optional note field, submit button. On submit call createSchedule API, show success/error feedback
- [x] T017 [US1] Run all US1 tests (T010, T011) — all must pass green

**Checkpoint**: Parent can create weekly schedules, scheduler auto-deposits. Core MVP functional.

---

## Phase 4: User Story 2 — Parent Views and Manages Schedules (Priority: P1)

**Goal**: Parent can view all allowance schedules for their family, edit amount/frequency, pause/resume, and delete schedules.

**Independent Test**: Create multiple schedules → view list with child names, amounts, frequencies, next dates → edit one → pause one → delete one → verify changes.

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T018 [P] [US2] Write handler tests for `GET /api/schedules` (list), `GET /api/schedules/{id}` (get), `PUT /api/schedules/{id}` (update), `DELETE /api/schedules/{id}` (delete), `POST /api/schedules/{id}/pause` (pause), `POST /api/schedules/{id}/resume` (resume) in `backend/internal/allowance/handler_test.go` — test: list returns schedules with child_first_name; update changes amount and recalculates next_run_at; pause sets status to "paused"; resume sets status to "active" and recalculates next_run_at from today; delete removes schedule; 404 for wrong family's schedule; already-paused returns 400; already-active resume returns 400

### Implementation for User Story 2

- [x] T019 [US2] Add `HandleListSchedules` (GET /api/schedules — returns ScheduleListResponse with child names), `HandleGetSchedule` (GET /api/schedules/{id}), `HandleUpdateSchedule` (PUT /api/schedules/{id} — validate partial update fields, recalculate next_run_at if frequency/day changed), `HandleDeleteSchedule` (DELETE /api/schedules/{id} — returns 204), `HandlePauseSchedule` (POST /api/schedules/{id}/pause — set status paused, clear next_run_at), `HandleResumeSchedule` (POST /api/schedules/{id}/resume — set status active, recalculate next_run_at from today) to `backend/internal/allowance/handler.go` (implemented in T012)
- [x] T020 [US2] Register all schedule management routes in `backend/main.go`: GET /api/schedules, GET /api/schedules/{id}, PUT /api/schedules/{id}, DELETE /api/schedules/{id}, POST /api/schedules/{id}/pause, POST /api/schedules/{id}/resume — all with parent auth middleware (implemented in T014)
- [x] T021 [P] [US2] Add API functions to `frontend/src/api.ts`: `getSchedules`, `getSchedule`, `updateSchedule`, `deleteSchedule`, `pauseSchedule`, `resumeSchedule`
- [x] T022 [US2] Create `frontend/src/components/ScheduleList.tsx` — display table/list of all schedules (child name, amount formatted as dollars, frequency, next deposit date, status badge active/paused). Include action buttons: Edit, Pause/Resume toggle, Delete (with confirmation). Edit opens ScheduleForm in edit mode. Wire to API functions.
- [x] T023 [US2] Add allowance schedule management link/navigation to `frontend/src/pages/ParentDashboard.tsx` — link to schedule list view, integrate ScheduleList and ScheduleForm components (can be on same page or separate route)
- [x] T024 [US2] Run all US2 tests (T018) — all must pass green

**Checkpoint**: Full parent management CRUD. Parents can create, view, edit, pause, resume, and delete schedules.

---

## Phase 5: User Story 3 — Biweekly and Monthly Frequencies (Priority: P2)

**Goal**: Parent can create biweekly or monthly allowance schedules with correct date handling including end-of-month clamping.

**Independent Test**: Create monthly schedule for 31st → verify February deposits on 28th (or 29th leap year). Create biweekly schedule → verify 14-day intervals maintained.

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T025 [P] [US3] Write handler tests for creating biweekly schedule (requires day_of_week) and monthly schedule (requires day_of_month, rejects day_of_week) in `backend/internal/allowance/handler_test.go` — test: monthly with day_of_month returns 201; biweekly with day_of_week returns 201; monthly without day_of_month returns 400; biweekly without day_of_week returns 400
- [x] T026 [P] [US3] Write scheduler tests for biweekly execution (advances next_run_at by 14 days) and monthly execution (advances to correct day next month, clamping for short months) in `backend/internal/allowance/scheduler_test.go`

### Implementation for User Story 3

- [x] T027 [US3] Update `HandleCreateSchedule` validation in `backend/internal/allowance/handler.go` to accept "biweekly" and "monthly" frequencies — for biweekly: require day_of_week; for monthly: require day_of_month (1-31), reject day_of_week. Update `HandleUpdateSchedule` to handle frequency changes between types. (implemented in T012)
- [x] T028 [US3] Update `ScheduleForm.tsx` in `frontend/src/components/ScheduleForm.tsx` to show day-of-month selector (1-31) when monthly is selected, and day-of-week selector when biweekly is selected. Conditionally show/hide the appropriate day picker based on frequency. (implemented in T016)
- [x] T029 [US3] Run all US3 tests (T025, T026) — all must pass green

**Checkpoint**: All three frequencies (weekly, biweekly, monthly) fully functional with correct date math.

---

## Phase 6: User Story 4 — Child Sees Upcoming Allowance (Priority: P3)

**Goal**: Child can see when their next allowance arrives and for how much on their dashboard.

**Independent Test**: Child with active schedule sees "Next allowance: $10 on Friday". Child with no schedule sees no allowance info.

### Tests for User Story 4 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T030 [P] [US4] Write store test for `ListActiveByChild` method in `backend/internal/store/schedule_test.go` — returns only active schedules for specified child, sorted by next_run_at (implemented in T006)
- [x] T031 [P] [US4] Write handler test for `GET /api/children/{childId}/upcoming-allowances` in `backend/internal/allowance/handler_test.go` — test: returns upcoming allowances sorted by date; child can only see their own; parent can see any child in family; no active schedules returns empty array

### Implementation for User Story 4

- [x] T032 [US4] Add `ListActiveByChild(childID int64)` method to `ScheduleStore` in `backend/internal/store/schedule.go` — query active schedules for child, return sorted by next_run_at (implemented in T008)
- [x] T033 [US4] Add `HandleGetUpcomingAllowances` method to `Handler` in `backend/internal/allowance/handler.go` — get child ID from URL, verify authorization (child sees own, parent sees family), call ListActiveByChild, map to UpcomingAllowancesResponse, return JSON (implemented in T012)
- [x] T034 [US4] Register `GET /api/children/{childId}/upcoming-allowances` route in `backend/main.go` with auth middleware (both parent and child roles) (implemented in T014)
- [x] T035 [P] [US4] Add `getUpcomingAllowances` API function to `frontend/src/api.ts` (implemented in T015)
- [x] T036 [US4] Create `frontend/src/components/UpcomingAllowances.tsx` — display upcoming allowance info for nearest scheduled deposits. If no active schedules, render nothing.
- [x] T037 [US4] Integrate `UpcomingAllowances` component into `frontend/src/pages/ChildDashboard.tsx` — fetch upcoming allowances on mount, display component above transaction history
- [x] T038 [US4] Run all US4 tests (T030, T031) — all must pass green

**Checkpoint**: Children can see their upcoming allowance. All four user stories complete.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories, edge cases, and cleanup

- [x] T039 Verify edge case: scheduled deposits after system downtime — tested in TestScheduler_ProcessDueSchedules_HandlesMissed (schedule with past next_run_at is processed)
- [ ] T040 Verify edge case: deleting a child cascades to delete their allowance schedules — confirm ON DELETE CASCADE works with existing child deletion flow
- [ ] T041 Verify allowance transactions display correctly in existing transaction history views for both parent and child dashboards — "allowance" type should render with clear automatic-deposit indicator
- [x] T042 Run full backend test suite: `cd backend && go test ./... -v` — all tests pass
- [x] T043 Run frontend build verification: `cd frontend && npm run build` — no TypeScript errors

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup (T001-T004) — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) completion
- **User Story 2 (Phase 4)**: Depends on US1 (Phase 3) — US2 extends the handler with management endpoints
- **User Story 3 (Phase 5)**: Depends on US2 (Phase 4) — US3 extends validation and UI for additional frequencies
- **User Story 4 (Phase 6)**: Depends on Foundational (Phase 2) — can run in parallel with US1-US3 if desired
- **Polish (Phase 7)**: Depends on all user stories being complete

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Store layer before handlers
- Handlers before routes
- Backend before frontend
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

Within Phase 1:
- T002, T003, T004 can all run in parallel (different files)

Within Phase 2:
- T005, T006 can run in parallel (different test files)

Within each User Story phase:
- Test tasks marked [P] can run in parallel
- Frontend API functions marked [P] can run alongside backend route registration

Cross-phase:
- US4 (Phase 6) can run in parallel with US1-US3 since it only depends on Foundational phase

---

## Parallel Example: User Story 1

```bash
# Launch US1 tests in parallel:
Task: "Write handler test for POST /api/schedules in backend/internal/allowance/handler_test.go"
Task: "Write scheduler tests in backend/internal/allowance/scheduler_test.go"

# After tests written, implement sequentially:
Task: "Create handler.go with HandleCreateSchedule"
Task: "Create scheduler.go with ProcessDueSchedules"
Task: "Register route and start scheduler in main.go"

# Frontend can start after backend route is registered:
Task: "Add createSchedule to frontend/src/api.ts"  (parallel with route registration)
Task: "Create ScheduleForm.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (migration + types)
2. Complete Phase 2: Foundational (date math + store CRUD)
3. Complete Phase 3: User Story 1 (create weekly schedule + auto-deposit)
4. **STOP and VALIDATE**: Create a weekly schedule, verify scheduler deposits on time
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Full parent management → Deploy/Demo
4. Add User Story 3 → All frequencies supported → Deploy/Demo
5. Add User Story 4 → Child visibility → Deploy/Demo
6. Polish → Edge cases verified, full test suite green

### Suggested MVP Scope

**User Story 1 only** (Phases 1-3, tasks T001-T017). This delivers the core value: parents can create a weekly allowance that auto-deposits. Everything else is incremental.
