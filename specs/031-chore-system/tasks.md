# Tasks: Chore & Task System

**Input**: Design documents from `/specs/031-chore-system/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/api-endpoints.md, research.md

**Tests**: Included per constitution (Test-First Development principle).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Database migration and GORM models shared by all user stories

- [x] T001 Create database migration in `backend/migrations/011_chores.up.sql` — create `chores`, `chore_assignments`, and `chore_instances` tables with all columns, indexes, constraints, and foreign keys per data-model.md
- [x] T002 Create down migration in `backend/migrations/011_chores.down.sql` — drop `chore_instances`, `chore_assignments`, `chores` tables in order
- [x] T003 Define GORM models (Chore, ChoreAssignment, ChoreInstance) and type constants (ChoreRecurrence, ChoreInstanceStatus) in `backend/models/chore.go` per data-model.md
- [x] T004 Add `TransactionTypeChore TransactionType = "chore"` constant to `backend/models/transaction.go`
- [x] T005 Add `"chore"` to the `TransactionType` in `frontend/src/types.ts` and add `Chore`, `ChoreAssignment`, `ChoreInstance`, `ChoreInstanceStatus` TypeScript interfaces

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Repository layer that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Write tests for ChoreRepo (Create, GetByID, ListByFamily, CreateAssignment, DeleteAssignment) in `backend/repositories/chore_repo_test.go` — verify CRUD, family scoping, and unique assignment constraint
- [x] T007 Implement ChoreRepo in `backend/repositories/chore_repo.go` — Create, GetByID, ListByFamily (with assignments + pending count), CreateAssignment, DeleteAssignment, Delete
- [x] T008 Write tests for ChoreInstanceRepo (CreateInstance, GetByID, ListByChild, MarkComplete, Approve, Reject, Expire) in `backend/repositories/chore_instance_repo_test.go` — verify status transitions, reward snapshotting, and family scoping
- [x] T009 Implement ChoreInstanceRepo in `backend/repositories/chore_instance_repo.go` — CreateInstance (snapshots reward_cents from chore), GetByID, ListByChild (grouped by status), ListPendingByFamily, MarkComplete, Approve (sets reviewed_at, transaction_id), Reject (resets to available, clears completed_at), ExpireByPeriod
- [x] T010 Write test for DepositChore in `backend/repositories/transaction_repo_test.go` — verify atomic deposit with "chore" transaction type and correct note
- [x] T011 Add DepositChore method to `backend/repositories/transaction_repo.go` — same pattern as DepositAllowance but with TransactionTypeChore and no schedule_id

**Checkpoint**: Foundation ready — repository layer complete with passing tests

---

## Phase 3: User Story 1 — Parent Creates a Chore (Priority: P1) 🎯 MVP

**Goal**: Parents can create one-time chores, assign them to children, and view their chore list

**Independent Test**: Parent creates a chore → it appears in GET /api/chores with correct assignments

### Tests for User Story 1

- [x] T012 [P] [US1] Write handler tests for POST /api/chores in `backend/internal/chore/handler_test.go` — test valid creation, multi-child assignment, validation errors (blank name, negative amount, invalid recurrence, child not in family), and 403 for child users
- [x] T013 [P] [US1] Write handler tests for GET /api/chores in `backend/internal/chore/handler_test.go` — test list returns chores with assignments and pending counts, empty list, family scoping

### Implementation for User Story 1

- [x] T014 [US1] Create chore Handler struct with repo dependencies and constructor in `backend/internal/chore/handler.go` — include ChoreRepo, ChoreInstanceRepo, TransactionRepo, ChildRepo, FamilyRepo; add writeJSON and ErrorResponse helpers
- [x] T015 [US1] Implement HandleCreateChore in `backend/internal/chore/handler.go` — parse request, validate name (1-100), reward_cents (0-99999999), recurrence, child_ids (all in family); create chore + assignments; create initial instances for one-time chores; return 201
- [x] T016 [US1] Implement HandleListChores in `backend/internal/chore/handler.go` — list chores by family with assignments and pending instance counts
- [x] T017 [US1] Wire chore handler and routes (POST /api/chores, GET /api/chores) in `backend/main.go` — initialize repos, create handler, register routes with requireParent middleware
- [x] T018 [US1] Add ChoreForm component in `frontend/src/components/ChoreForm.tsx` — form with name, description, reward amount (dollar input converted to cents), child selector (checkboxes from children list), recurrence dropdown (one-time only for MVP); client-side validation
- [x] T019 [US1] Create ParentChores page in `frontend/src/pages/ParentChores.tsx` — list all chores with assignments and status; include "Add Chore" button that opens ChoreForm; display chore cards with name, reward, assigned children
- [x] T020 [US1] Add parent chore route to `frontend/src/App.tsx` — add `/chores` route pointing to ParentChores page; add navigation link in parent dashboard/nav

**Checkpoint**: Parent can create and view chores. MVP deliverable.

---

## Phase 4: User Story 2 — Child Completes a Chore (Priority: P1)

**Goal**: Children can view their assigned chores and mark them as complete

**Independent Test**: Child sees chore in "available" list → marks complete → chore moves to "pending" list

### Tests for User Story 2

- [x] T021 [P] [US2] Write handler tests for GET /api/child/chores in `backend/internal/chore/handler_test.go` — test returns instances grouped by status (available/pending/completed), empty state, child can only see own instances
- [x] T022 [P] [US2] Write handler tests for POST /api/child/chores/{id}/complete in `backend/internal/chore/handler_test.go` — test status transition to pending_approval, 400 if not available, 404 if wrong child, 403 for parent users

### Implementation for User Story 2

- [x] T023 [US2] Implement HandleChildListChores in `backend/internal/chore/handler.go` — query instances by child_id grouped by status, join chore name/description; return available, pending, completed arrays
- [x] T024 [US2] Implement HandleCompleteChore in `backend/internal/chore/handler.go` — validate instance belongs to child, status is "available"; update to "pending_approval", set completed_at; clear any previous rejection_reason
- [x] T025 [US2] Wire child chore routes (GET /api/child/chores, POST /api/child/chores/{id}/complete) in `backend/main.go` — register with requireAuth middleware, verify child user type in handler
- [x] T026 [US2] Create ChoreCard component in `frontend/src/components/ChoreCard.tsx` — display chore name, description, reward amount; show status badge (available/pending/completed); "Mark Done" button for available chores; show rejection reason if present
- [x] T027 [US2] Create ChildChores page in `frontend/src/pages/ChildChores.tsx` — fetch and display chore instances grouped by status (available first, then pending, then completed); empty state message; call complete endpoint on button click
- [x] T028 [US2] Add child chore route to `frontend/src/App.tsx` — add `/chores` route for child users pointing to ChildChores page; add navigation link in child dashboard/nav

**Checkpoint**: Children can view and complete chores. Full create→complete flow works.

---

## Phase 5: User Story 3 — Parent Approves or Rejects Completion (Priority: P1)

**Goal**: Parents can approve pending chores (triggering deposit) or reject them (returning to available)

**Independent Test**: Approve a pending chore → verify deposit in child's transaction history and balance increase; reject → verify chore returns to available

### Tests for User Story 3

- [x] T029 [P] [US3] Write handler tests for GET /api/chores/pending in `backend/internal/chore/handler_test.go` — test returns pending instances across all children with chore/child names, empty list
- [x] T030 [P] [US3] Write handler tests for POST /api/chore-instances/{id}/approve in `backend/internal/chore/handler_test.go` — test deposit created with correct amount and "chore" type, balance updated, 400 if not pending, 403 if child disabled, 404 if wrong family
- [x] T031 [P] [US3] Write handler tests for POST /api/chore-instances/{id}/reject in `backend/internal/chore/handler_test.go` — test status reset to available, rejection reason saved, completed_at cleared, no transaction created, 400 if not pending

### Implementation for User Story 3

- [x] T032 [US3] Implement HandleListPending in `backend/internal/chore/handler.go` — query pending instances by family, join chore name and child name
- [x] T033 [US3] Implement HandleApprove in `backend/internal/chore/handler.go` — validate instance is pending_approval; check child not disabled; call DepositChore with snapshotted reward_cents and note "Chore: {chore_name}"; update instance status to approved, set reviewed_at, reviewed_by_parent_id, transaction_id; skip deposit if reward_cents is 0; return instance + new_balance
- [x] T034 [US3] Implement HandleReject in `backend/internal/chore/handler.go` — validate instance is pending_approval; parse optional reason (max 500 chars); reset status to available, clear completed_at, set rejection_reason, set reviewed_at and reviewed_by_parent_id; return updated instance
- [x] T035 [US3] Wire approval routes (GET /api/chores/pending, POST /api/chore-instances/{id}/approve, POST /api/chore-instances/{id}/reject) in `backend/main.go` — register with requireParent middleware
- [x] T036 [US3] Create ChoreApprovalQueue component in `frontend/src/components/ChoreApprovalQueue.tsx` — list pending instances with child name, chore name, reward, completed_at; approve/reject buttons per instance; optional rejection reason input (text field shown on reject click); show success/error feedback
- [x] T037 [US3] Integrate ChoreApprovalQueue into ParentChores page in `frontend/src/pages/ParentChores.tsx` — show approval queue section above/alongside chore list; show pending count badge; refresh chore list after approve/reject

**Checkpoint**: Complete chore lifecycle works end-to-end: create → complete → approve/reject → deposit.

---

## Phase 6: User Story 4 — Parent Creates a Recurring Chore (Priority: P2)

**Goal**: Recurring chores (daily/weekly/monthly) auto-generate instances and expire missed ones

**Independent Test**: Create weekly chore → scheduler generates instance for current period → approve → next period instance appears on next tick

### Tests for User Story 4

- [x] T038 [P] [US4] Write scheduler tests in `backend/internal/chore/scheduler_test.go` — test instance generation for daily/weekly/monthly chores respecting family timezone; test expiry of available instances past period_end; test skipping disabled children; test skipping inactive chores; test no duplicate instances for same period

### Implementation for User Story 4

- [x] T039 [US4] Implement period calculation helpers in `backend/internal/chore/scheduler.go` — CalculatePeriodBounds(recurrence, dayOfWeek, dayOfMonth, now, loc) returns (periodStart, periodEnd) for daily/weekly/monthly; timezone-aware using family IANA timezone
- [x] T040 [US4] Implement ChoreScheduler struct with Start method in `backend/internal/chore/scheduler.go` — goroutine with time.NewTicker and stop channel; two jobs per tick: GenerateInstances and ExpireInstances
- [x] T041 [US4] Implement GenerateInstances in `backend/internal/chore/scheduler.go` — find all active recurring chores; for each chore+assignment, calculate current period; if no instance exists for (chore_id, child_id, period_start), create one with snapshotted reward_cents; skip disabled children
- [x] T042 [US4] Implement ExpireInstances in `backend/internal/chore/scheduler.go` — find all "available" instances where period_end < now; update status to "expired"
- [x] T043 [US4] Wire ChoreScheduler in `backend/main.go` — create stop channel, initialize scheduler with repos, start with 5-minute interval, defer close stop channel
- [x] T044 [US4] Implement HandleActivate and HandleDeactivate in `backend/internal/chore/handler.go` — toggle is_active on chore; return updated chore
- [x] T045 [US4] Wire activate/deactivate routes (PATCH /api/chores/{id}/activate, PATCH /api/chores/{id}/deactivate) in `backend/main.go` — register with requireParent middleware
- [x] T046 [US4] Update ChoreForm in `frontend/src/components/ChoreForm.tsx` — add recurrence selector (one-time/daily/weekly/monthly); show day_of_week picker for weekly; show day_of_month picker for monthly; update create API call with recurrence fields
- [x] T047 [US4] Update ParentChores page in `frontend/src/pages/ParentChores.tsx` — show recurrence badge on chore cards; show active/inactive toggle for recurring chores (calls PATCH activate/deactivate)
- [x] T048 [US4] Update ChildChores page in `frontend/src/pages/ChildChores.tsx` — show period dates on recurring chore instances; filter to show only current period's available instance (not future)

**Checkpoint**: Recurring chores auto-generate and expire. Full recurring lifecycle works.

---

## Phase 7: User Story 5 — Child Views Chore Earnings History (Priority: P3)

**Goal**: Children see total chore earnings and a list of completed chores with amounts

**Independent Test**: Child with approved chores → earnings view shows correct total and list

### Tests for User Story 5

- [x] T049 [P] [US5] Write handler test for GET /api/child/chores/earnings in `backend/internal/chore/handler_test.go` — test total_earned_cents sums correctly, chores_completed count, recent list ordered by approved_at desc, empty state

### Implementation for User Story 5

- [x] T050 [US5] Implement HandleChildEarnings in `backend/internal/chore/handler.go` — query approved instances for child; aggregate total_earned_cents and chores_completed count; return recent approved instances with chore name, reward, and approved_at
- [x] T051 [US5] Wire earnings route (GET /api/child/chores/earnings) in `backend/main.go` — register with requireAuth middleware
- [x] T052 [US5] Add earnings summary section to ChildChores page in `frontend/src/pages/ChildChores.tsx` — show total earned, chores completed count, and recent earnings list; empty state with encouraging message

**Checkpoint**: Children can see their chore earnings history.

---

## Phase 8: User Story 6 — Parent Edits or Deletes a Chore (Priority: P3)

**Goal**: Parents can update chore details or delete chores; changes affect future instances only

**Independent Test**: Edit reward amount → new instance uses new amount, existing approved transaction unchanged; delete chore → chore disappears, transactions preserved

### Tests for User Story 6

- [x] T053 [P] [US6] Write handler tests for PUT /api/chores/{id} in `backend/internal/chore/handler_test.go` — test name/description/reward/recurrence updates, validation errors, 404 for wrong family, verify existing instances unaffected
- [x] T054 [P] [US6] Write handler tests for DELETE /api/chores/{id} in `backend/internal/chore/handler_test.go` — test chore deleted, pending instances cancelled, approved transactions preserved, 404 for wrong family

### Implementation for User Story 6

- [x] T055 [US6] Add Update method to ChoreRepo in `backend/repositories/chore_repo.go` — update name, description, reward_cents, recurrence, day_of_week, day_of_month; set updated_at
- [x] T056 [US6] Implement HandleUpdateChore in `backend/internal/chore/handler.go` — parse request, validate fields, verify chore belongs to family; call repo Update; return updated chore
- [x] T057 [US6] Implement HandleDeleteChore in `backend/internal/chore/handler.go` — verify chore belongs to family; cancel pending instances (set status to expired); delete chore (cascades to assignments); return 204
- [x] T058 [US6] Wire edit/delete routes (PUT /api/chores/{id}, DELETE /api/chores/{id}) in `backend/main.go`
- [x] T059 [US6] Add edit mode to ChoreForm in `frontend/src/components/ChoreForm.tsx` — pre-populate fields from existing chore; call PUT endpoint on save; show delete button with confirmation dialog
- [x] T060 [US6] Add edit/delete actions to chore cards in `frontend/src/pages/ParentChores.tsx` — edit icon opens ChoreForm in edit mode; delete icon shows confirmation; refresh list after action

**Checkpoint**: Full CRUD lifecycle for chores is complete.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final integration, edge cases, and cleanup

- [x] T061 [P] Update TRUNCATE list in test helpers to include `chore_instances, chore_assignments, chores` in `backend/internal/testutil/` or store test helper files
- [x] T062 Write end-to-end chore lifecycle integration test in `backend/internal/chore/handler_test.go` — single test that: creates a chore (POST /api/chores), marks it complete as child (POST /api/child/chores/{id}/complete), approves as parent (POST /api/chore-instances/{id}/approve), then verifies: child balance increased by reward_cents, transaction exists with type "chore" and correct note, instance status is "approved" with transaction_id set
- [x] T063 [P] Add chore navigation links to both parent and child nav/sidebar components — ensure consistent placement with existing nav items
- [x] T064 Run full backend test suite (`cd backend && go test -p 1 ./...`) and fix any failures
- [x] T065 Run frontend checks (`cd frontend && npx tsc --noEmit && npm run build && npm run lint`) and fix any issues
- [ ] T066 Run quickstart.md manual validation — execute the full manual test flow end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2
- **US2 (Phase 4)**: Depends on Phase 2 (independent of US1 at backend level, but frontend benefits from US1 being done first for nav/routing)
- **US3 (Phase 5)**: Depends on Phase 2 (needs instances in pending_approval state, so testing requires US1+US2 flow, but code is independent)
- **US4 (Phase 6)**: Depends on Phase 2 (scheduler is independent of handlers)
- **US5 (Phase 7)**: Depends on Phase 2 (needs approved instances, so testing requires US3 flow)
- **US6 (Phase 8)**: Depends on Phase 2 + T007 (needs ChoreRepo.Update which could be in Phase 2, but is story-specific). Activate/deactivate routes handled in US4
- **Polish (Phase 9)**: Depends on all completed stories

### User Story Dependencies

- **US1 (P1)**: No story dependencies — foundational only
- **US2 (P1)**: No story dependencies — but instances must exist (created in US1 or directly via repo)
- **US3 (P1)**: No story dependencies — but instances must be in pending_approval state
- **US4 (P2)**: No story dependencies — scheduler is independent of handlers
- **US5 (P3)**: No story dependencies — reads approved instances from DB
- **US6 (P3)**: No story dependencies — CRUD operations on chores

### Recommended Execution Order

Sequential (solo developer): Phase 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9

### Within Each User Story

1. Tests FIRST — write and verify they FAIL
2. Backend implementation (handler methods)
3. Route wiring in main.go
4. Frontend components and pages
5. Verify tests PASS

### Parallel Opportunities

**Phase 1**: T001+T002 (migrations) can parallel with T003+T004+T005 (models/types)
**Phase 2**: T006+T007 (ChoreRepo) can parallel with T010+T011 (DepositChore); T008+T009 after T007
**Phase 3**: T012+T013 (tests) in parallel
**Phase 5**: T029+T030+T031 (tests) in parallel
**Phase 6**: T038 (scheduler tests) independent of handler tests
**Phase 8**: T053+T054 (tests) in parallel

---

## Parallel Example: User Story 1

```text
# Write tests in parallel:
T012: "Handler tests for POST /api/chores in backend/internal/chore/handler_test.go"
T013: "Handler tests for GET /api/chores in backend/internal/chore/handler_test.go"

# Then implement sequentially:
T014: "Handler struct and constructor"
T015: "HandleCreateChore"
T016: "HandleListChores"
T017: "Wire routes in main.go"

# Then frontend in parallel:
T018: "ChoreForm component"
T019: "ParentChores page"
T020: "Add route to App.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (migrations + models)
2. Complete Phase 2: Foundational (repos + tests)
3. Complete Phase 3: User Story 1 (create + list chores)
4. **STOP and VALIDATE**: Parent can create and view chores
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Parent creates chores → Deploy/Demo (MVP!)
3. Add US2 → Child completes chores → Deploy/Demo
4. Add US3 → Parent approves/rejects → Deploy/Demo (core loop complete!)
5. Add US4 → Recurring chores → Deploy/Demo
6. Add US5+US6 → Earnings + edit/delete → Deploy/Demo (feature complete)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires TDD: write tests first, verify they fail, then implement
- Money is always int64 cents — never use floats
- All chore queries must be family-scoped (prevent cross-family data access)
- Disabled children: block approval and instance generation, allow viewing
- Reward snapshotting: instance.reward_cents is set at creation, not at approval
- Commit after each task or logical group
