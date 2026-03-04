# Tasks: Savings Goals

**Input**: Design documents from `/specs/025-savings-goals/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md

**Tests**: Included — constitution mandates Test-First Development (TDD) for all feature work.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database schema and test infrastructure for savings goals

- [x] T001 Create migration `backend/migrations/007_savings_goals.up.sql` — create `savings_goals` table (id SERIAL PK, child_id INT NOT NULL REFERENCES children(id) ON DELETE CASCADE, name TEXT NOT NULL, target_cents BIGINT NOT NULL CHECK(target_cents > 0), saved_cents BIGINT NOT NULL DEFAULT 0 CHECK(saved_cents >= 0), emoji TEXT, target_date DATE, status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','completed')), completed_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()) and `goal_allocations` table (id SERIAL PK, goal_id INT NOT NULL REFERENCES savings_goals(id) ON DELETE CASCADE, child_id INT NOT NULL REFERENCES children(id) ON DELETE CASCADE, amount_cents BIGINT NOT NULL CHECK(amount_cents != 0), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()) with indexes: idx_savings_goals_child_status(child_id, status), idx_savings_goals_child_created(child_id, created_at DESC), idx_goal_allocations_goal(goal_id, created_at DESC). Create corresponding `backend/migrations/007_savings_goals.down.sql` that drops both tables.
- [x] T002 Update `backend/internal/testutil/db.go` — add `goal_allocations, savings_goals` to the TRUNCATE statement (before `transactions` to respect FK order)

**Checkpoint**: Migration can be applied and test DB cleanup includes new tables

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Base store, handler, types, and routing that all user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Create `backend/internal/store/savings_goal.go` — define `SavingsGoal` struct (ID, ChildID, Name, TargetCents, SavedCents int64; Emoji *string; TargetDate *time.Time; Status string; CompletedAt *time.Time; CreatedAt, UpdatedAt time.Time) and `GoalAllocation` struct (ID, GoalID, ChildID, AmountCents int64; CreatedAt time.Time). Create `SavingsGoalStore` struct with `db *sql.DB` field and `NewSavingsGoalStore(db *sql.DB)` constructor. Add a `scanGoal` helper that scans a row into a SavingsGoal handling nullable fields (emoji, target_date, completed_at) with sql.NullString/sql.NullTime.
- [x] T004 [P] Create `backend/internal/goals/handler.go` — define `Handler` struct with `goalStore *store.SavingsGoalStore`, `childStore *store.ChildStore` fields. Create `NewHandler` constructor. Add `writeJSON` and `ErrorResponse` helpers matching the pattern in `backend/internal/balance/handler.go`.
- [x] T005 [P] Add savings goal TypeScript types to `frontend/src/types.ts` — add `SavingsGoal` interface (id, child_id, name, target_cents, saved_cents, emoji?: string, target_date?: string, status: 'active'|'completed', completed_at?: string, created_at, updated_at), `GoalAllocation` interface (id, goal_id, amount_cents, created_at), `SavingsGoalsResponse` (goals: SavingsGoal[], available_balance_cents, total_saved_cents), `AllocateResponse` (goal: SavingsGoal, available_balance_cents, completed: boolean), `DeleteGoalResponse` (released_cents, available_balance_cents). Update `BalanceResponse` to include optional `available_balance_cents?: number`, `total_saved_cents?: number`, `active_goals_count?: number`.
- [x] T006 [P] Add savings goal API functions to `frontend/src/api.ts` — add `getSavingsGoals(childId)`, `createSavingsGoal(childId, data)`, `updateSavingsGoal(childId, goalId, data)`, `deleteSavingsGoal(childId, goalId)`, `allocateToGoal(childId, goalId, amountCents)`, `getGoalAllocations(childId, goalId)` using existing get/post/put/request patterns.
- [x] T007 Register savings goal routes and store in `backend/main.go` — instantiate `SavingsGoalStore`, create goals `Handler`, register routes: `GET /api/children/{id}/savings-goals` (requireAuth), `POST /api/children/{id}/savings-goals` (requireAuth), `PUT /api/children/{id}/savings-goals/{goalId}` (requireAuth), `DELETE /api/children/{id}/savings-goals/{goalId}` (requireAuth), `POST /api/children/{id}/savings-goals/{goalId}/allocate` (requireAuth), `GET /api/children/{id}/savings-goals/{goalId}/allocations` (requireAuth).

**Checkpoint**: Foundation ready — store/handler files exist, routes registered, frontend types and API functions defined. All user story implementation can now begin.

---

## Phase 3: User Story 1 — Create a Savings Goal (Priority: P1) MVP

**Goal**: Children can create savings goals with a name, target amount, optional emoji, and optional target date. Goals appear in a list on a dedicated goals page.

**Independent Test**: Child logs in, creates a goal with name and target amount, verifies it appears in the goals list at 0% progress.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T008 [P] [US1] Write store tests in `backend/internal/store/savings_goal_test.go` — test `Create` (success with all fields, success with only required fields, name too long returns error), `GetByID` (found, not found returns nil), `ListByChild` (returns active then completed ordered by created_at, empty list), `CountActiveByChild` (counts only active goals, ignores completed)
- [x] T009 [P] [US1] Write handler tests in `backend/internal/goals/handler_test.go` — test `HandleCreate` (success 201 with valid body, 400 missing name, 400 missing target_cents, 400 target_cents <= 0, 400 name > 50 chars, 403 parent cannot create, 409 already has 5 active goals) and `HandleList` (200 returns goals list, 200 empty list, 403 child cannot see sibling goals, 200 parent can see family child goals)

### Implementation for User Story 1

- [x] T010 [US1] Implement `SavingsGoalStore.Create`, `GetByID`, `ListByChild`, `CountActiveByChild` in `backend/internal/store/savings_goal.go` — Create inserts a row and returns the created goal via RETURNING; GetByID queries by id using scanGoal; ListByChild queries WHERE child_id=$1 ORDER BY status ASC (active first), created_at DESC; CountActiveByChild returns count WHERE child_id=$1 AND status='active'. Run tests to verify they pass.
- [x] T011 [US1] Implement `HandleCreate` and `HandleList` in `backend/internal/goals/handler.go` — HandleCreate: require child user type, parse child ID from URL, verify child owns the account (userID == childID && familyID match), decode request body, validate name (1-50 chars), target_cents (> 0, ≤ 99999999), check CountActiveByChild < 5 (409 if at max), call store.Create, return 201. HandleList: require auth, parse child ID, verify access (child can see own, parent can see family child), call store.ListByChild, compute available_balance_cents and total_saved_cents from the goals, return 200. Run tests to verify they pass.
- [x] T012 [P] [US1] Create `frontend/src/components/GoalProgressRing.tsx` — circular SVG progress ring component. Props: percent (0-100), size (default 64px), strokeWidth (default 6px). Render a background circle and a foreground arc using stroke-dasharray/stroke-dashoffset. Use theme CSS variables (text-forest for progress fill, border-sand for background track). Show percentage text in center. Include smooth CSS transition on the progress arc.
- [x] T013 [P] [US1] Create `frontend/src/components/GoalCard.tsx` — Card-based component displaying a single goal. Props: goal (SavingsGoal), onAllocate callback, onEdit callback, onDelete callback. Show emoji (or default target icon from lucide-react), goal name, GoalProgressRing, amount saved vs target (e.g., "$20.00 / $45.00"), and progress percentage. Use Card component with padding="sm". Style with theme colors (text-forest, text-bark, text-bark-light, border-sand).
- [x] T014 [US1] Create `frontend/src/components/GoalForm.tsx` — form component for creating (and later editing) goals. Props: onSubmit callback, onCancel callback, optional initialGoal for edit mode. Fields: name (Input, required, maxLength 50), target amount (dollar input with $ prefix, required, min 0.01), emoji (optional text input), target date (optional date input). Convert dollar amount to cents on submit. Show validation errors. Use Button primary for submit, secondary for cancel. Follow DepositForm pattern for error handling and loading state.
- [x] T015 [US1] Create `frontend/src/pages/SavingsGoalsPage.tsx` — page component for child goals management. Fetch goals via `getSavingsGoals(user.user_id)`. Show "My Savings Goals" heading. Show "Add Goal" button (disabled if 5 active goals, with tooltip). Render GoalCard for each active goal. Toggle GoalForm visibility when "Add Goal" is tapped. On successful creation, refresh goals list. Use `max-w-[480px] mx-auto space-y-6 animate-fade-in-up` layout matching ChildDashboard pattern.
- [x] T016 [US1] Add `/child/goals` route to `frontend/src/App.tsx` inside the child `AuthenticatedLayout` routes. Add a navigation link to the goals page from the child dashboard — a small card or button below the balance card in `frontend/src/pages/ChildDashboard.tsx` with a target/goal icon (from lucide-react) and "Savings Goals" label that links to `/child/goals`.

**Checkpoint**: Child can create goals, see them listed on the goals page, and navigate to/from the dashboard. Goals show 0% progress. This is the MVP.

---

## Phase 4: User Story 2 — Allocate Money Toward a Goal (Priority: P1)

**Goal**: Children can allocate money from their available balance toward a goal. Progress updates visually. Dashboard shows total balance prominently with available/saved breakdown.

**Independent Test**: Child creates a goal, allocates $10 toward it, verifies progress updates and available balance decreases by $10.

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T017 [P] [US2] Write store tests in `backend/internal/store/savings_goal_test.go` — test `Allocate` (success: saved_cents increases, allocation record created; error: amount exceeds available balance; error: amount is 0; error: goal not found), `GetAvailableBalance` (returns balance_cents minus sum of active goals' saved_cents; returns full balance when no goals), `ListAllocationsByGoal` (returns allocations ordered by created_at DESC)
- [x] T018 [P] [US2] Write handler tests in `backend/internal/goals/handler_test.go` — test `HandleAllocate` (200 success with positive amount, 200 success with negative amount for de-allocation, 400 amount is 0, 400 exceeds available balance, 400 de-allocation exceeds saved_cents, 403 parent cannot allocate, 404 goal not found) and `HandleListAllocations` (200 returns allocation list, 200 parent can view)

### Implementation for User Story 2

- [x] T019 [US2] Implement `SavingsGoalStore.Allocate` in `backend/internal/store/savings_goal.go` — within a DB transaction: verify goal exists and is active, for positive amounts verify available balance >= amount (query children.balance_cents minus SUM of active goals' saved_cents), for negative amounts verify goal's saved_cents >= abs(amount), update savings_goals.saved_cents, insert goal_allocations record, update savings_goals.updated_at. Return updated goal. Commit atomically.
- [x] T020 [US2] Implement `SavingsGoalStore.GetAvailableBalance` and `ListAllocationsByGoal` in `backend/internal/store/savings_goal.go` — GetAvailableBalance: query `SELECT c.balance_cents - COALESCE(SUM(sg.saved_cents), 0) FROM children c LEFT JOIN savings_goals sg ON sg.child_id = c.id AND sg.status = 'active' WHERE c.id = $1 GROUP BY c.balance_cents`. ListAllocationsByGoal: query goal_allocations WHERE goal_id=$1 ORDER BY created_at DESC. Run tests to verify they pass.
- [x] T021 [US2] Implement `HandleAllocate` and `HandleListAllocations` in `backend/internal/goals/handler.go` — HandleAllocate: require child user type, parse child ID and goal ID, verify ownership, decode amount_cents, validate non-zero, call store.Allocate, return updated goal + available_balance_cents + completed flag. HandleListAllocations: require auth, verify access (child own or parent family), call store.ListAllocationsByGoal, return allocations list. Run tests to verify they pass.
- [x] T022 [US2] Update `HandleGetBalance` in `backend/internal/balance/handler.go` — after fetching balance, also call `goalStore.GetAvailableBalance(childID)` and `goalStore.CountActiveByChild(childID)` to include `available_balance_cents`, `total_saved_cents` (balance - available), and `active_goals_count` in the response. Add `goalStore *store.SavingsGoalStore` to the balance Handler struct and update NewHandler constructor. Update handler instantiation in `backend/main.go` to pass goalStore.
- [x] T023 [US2] Add allocation UI to `frontend/src/components/GoalCard.tsx` — add an "Add Funds" button that reveals an inline amount input (dollar format with $ prefix). On submit, call `allocateToGoal(childId, goalId, amountCents)`. Show success animation (brief scale pulse on progress ring). Handle errors (insufficient funds message). Also add a "Remove Funds" option that sends a negative amount for de-allocation.
- [x] T024 [US2] Update `frontend/src/components/BalanceDisplay.tsx` — add optional `breakdown` prop with `availableCents` and `savedCents` fields. When provided, render a secondary line below the main balance showing "Available: $X.XX · Saved: $X.XX" in smaller text (text-sm text-bark-light). Total balance remains the large prominent display.
- [x] T025 [US2] Update `frontend/src/pages/ChildDashboard.tsx` — fetch balance response and use `available_balance_cents` and `total_saved_cents` from the updated BalanceResponse. Pass `breakdown` prop to BalanceDisplay when `total_saved_cents > 0`. The total balance (balance_cents) stays as the hero number.

**Checkpoint**: Child can allocate/de-allocate money to goals, progress updates, dashboard shows balance breakdown. Available balance prevents over-allocation.

---

## Phase 5: User Story 3 — View Goal Progress with Visual Feedback (Priority: P1)

**Goal**: Goals display with rich, theme-aware progress visualization. Near-completion goals show encouraging visual cues. Overdue target dates are indicated.

**Independent Test**: Create goals at 0%, 50%, 75%, and 100% progress levels. Verify each displays correctly with appropriate visual treatment across different themes.

### Implementation for User Story 3

- [x] T026 [US3] Enhance `frontend/src/components/GoalProgressRing.tsx` — use CSS custom properties (var(--color-forest) for fill, var(--color-sand) for track) so the ring automatically adapts to the child's selected theme. Add a smooth animated transition when percent changes (CSS transition on stroke-dashoffset, ~0.6s ease-out). Add a `milestone` prop that when true (>=75%) applies a subtle glow/pulse animation around the ring.
- [x] T027 [US3] Enhance `frontend/src/components/GoalCard.tsx` — when progress >= 75%, add a visual "almost there" indicator: a small sparkle icon (from lucide-react Sparkles) next to the progress percentage with encouraging text color (text-amber). When the goal has a target_date that has passed and status is still 'active', show a small "Overdue" badge in text-terracotta.
- [x] T028 [US3] Add goals summary section to `frontend/src/pages/ChildDashboard.tsx` — below the balance card, show a compact goals overview: up to 3 active goals rendered as mini goal cards (emoji + name + small GoalProgressRing size=40). Add a "View All Goals" link to `/child/goals`. Only render this section if the child has active goals. Use a horizontal scrollable row or compact grid layout.

**Checkpoint**: Goals look visually engaging with theme colors, milestone effects, and overdue indicators. Dashboard shows a goals preview.

---

## Phase 6: User Story 4 — Achieve a Goal (Priority: P2)

**Goal**: When allocation makes saved_cents >= target_cents, the goal is marked completed with a celebration animation. Completed goals appear in a separate section.

**Independent Test**: Create a goal with $5 target, allocate $5, verify celebration plays and goal moves to completed section with achievement date.

### Tests for User Story 4

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T029 [P] [US4] Write store test in `backend/internal/store/savings_goal_test.go` — test that `Allocate` with amount that makes saved_cents >= target_cents sets status='completed' and completed_at to non-nil time. Test that over-allocation (saved > target) also triggers completion. Test that allocating to an already completed goal returns an error.
- [x] T030 [P] [US4] Write handler test in `backend/internal/goals/handler_test.go` — test that `HandleAllocate` returns `completed: true` in response when allocation triggers completion. Verify the goal object in response has status='completed' and completed_at set.

### Implementation for User Story 4

- [x] T031 [US4] Update `SavingsGoalStore.Allocate` in `backend/internal/store/savings_goal.go` — after updating saved_cents, check if new saved_cents >= target_cents. If so, also UPDATE status='completed' AND completed_at=NOW() within the same transaction. Return the updated goal with completion info. Run tests to verify.
- [x] T032 [US4] Update `HandleAllocate` response in `backend/internal/goals/handler.go` — set `completed: true` in `AllocateResponse` when the goal transitions to completed status. Run tests to verify.
- [x] T033 [US4] Create `frontend/src/components/ConfettiCelebration.tsx` — lightweight confetti animation component. Props: show (boolean), onComplete callback. When `show` is true, render ~50 small colored squares/circles that animate from center outward with random trajectories using CSS keyframe animations. Use theme colors (var(--color-forest), var(--color-amber), var(--color-terracotta), var(--color-sage)). Auto-dismiss after ~2 seconds. No external dependencies — pure CSS animations.
- [x] T034 [US4] Add completed goals section to `frontend/src/pages/SavingsGoalsPage.tsx` — below active goals, render a "Completed Goals" section when completed goals exist. Show the 5 most recently completed goals (sorted by completed_at DESC) with a checkmark badge, goal name, amount achieved, and completion date. Add a "View All" button that expands to show all completed goals. Completed goal cards should have a muted/faded style (opacity-70 or similar).
- [x] T035 [US4] Integrate ConfettiCelebration into `frontend/src/pages/SavingsGoalsPage.tsx` — after an allocation response returns `completed: true`, trigger the ConfettiCelebration. Show a congratulatory message ("Goal Achieved!" with the goal name). Refresh the goals list after celebration completes to move the goal to the completed section.

**Checkpoint**: Goal completion triggers celebration, completed goals appear in their own section with achievement dates.

---

## Phase 7: User Story 5 — Manage Goals (Priority: P2)

**Goal**: Children can edit goal details (name, target, emoji, date) and delete goals. Deleting releases allocated funds. Reducing target below saved amount triggers completion.

**Independent Test**: Create a goal, edit its name and target, verify changes persist. Delete the goal, verify it disappears and allocated funds return to available balance.

### Tests for User Story 5

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T036 [P] [US5] Write store tests in `backend/internal/store/savings_goal_test.go` — test `Update` (success partial update name only, success update all fields, auto-complete when target reduced below saved, error goal not found, error goal already completed), `Delete` (success releases saved_cents by setting it to 0 before delete, error goal not found, error goal already completed — cannot delete completed goals)
- [x] T037 [P] [US5] Write handler tests in `backend/internal/goals/handler_test.go` — test `HandleUpdate` (200 success, 400 name too long, 400 target_cents <= 0, 403 parent cannot update, 404 not found or completed), `HandleDelete` (200 returns released_cents and available_balance_cents, 403 parent cannot delete, 404 not found or completed)

### Implementation for User Story 5

- [x] T038 [US5] Implement `SavingsGoalStore.Update` in `backend/internal/store/savings_goal.go` — accept partial update fields (name *string, targetCents *int64, emoji *string, emojiSet bool, targetDate *time.Time, targetDateSet bool). Build dynamic UPDATE query for provided fields. If target_cents is reduced to <= saved_cents, also set status='completed' and completed_at=NOW(). Return updated goal. Only allow updates on active goals.
- [x] T039 [US5] Implement `SavingsGoalStore.Delete` in `backend/internal/store/savings_goal.go` — within a DB transaction: verify goal exists and is active, read saved_cents, DELETE the goal (cascades to allocations), return released_cents. Only allow deletion of active goals. Run tests to verify.
- [x] T040 [US5] Implement `HandleUpdate` and `HandleDelete` in `backend/internal/goals/handler.go` — HandleUpdate: require child user type, parse IDs, verify ownership, decode partial update body, validate fields (name 1-50 chars, target_cents > 0 and ≤ 99999999), call store.Update, return updated goal (200). HandleDelete: require child user type, parse IDs, verify ownership, call store.Delete, compute new available_balance_cents, return released_cents + available_balance_cents (200). Run tests to verify.
- [x] T041 [US5] Update `frontend/src/components/GoalForm.tsx` — support edit mode via `initialGoal` prop. When provided, pre-populate all fields with existing goal data (convert cents to dollars for display). Change submit button label to "Save Changes". On submit, call `updateSavingsGoal` API instead of `createSavingsGoal`.
- [x] T042 [US5] Update `frontend/src/components/GoalCard.tsx` — add edit and delete actions. Add an overflow menu (three dots icon or gear icon) with "Edit" and "Delete" options. "Edit" opens GoalForm in edit mode (inline or modal). "Delete" shows a confirmation dialog ("Delete this goal? $X.XX will return to your available balance."). On delete confirmation, call `deleteSavingsGoal` API and refresh the goals list.

**Checkpoint**: Children can edit goal details, delete goals with fund release, and target reduction triggers auto-completion.

---

## Phase 8: User Story 6 — Parent Views Child's Goals (Priority: P3)

**Goal**: Parents can see their children's savings goals and progress from the parent dashboard (read-only).

**Independent Test**: Child creates goals, parent logs in and views the child's account — goals are visible with names, amounts, and progress.

### Tests for User Story 6

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T043 [US6] Write handler test in `backend/internal/goals/handler_test.go` — verify `HandleList` returns goals when called by a parent for a child in their family (200), returns 403 for a parent trying to view a child not in their family. Verify parent CANNOT call HandleCreate, HandleUpdate, HandleDelete, HandleAllocate (all return 403).

### Implementation for User Story 6

- [x] T044 [US6] Verify and adjust `HandleList` auth in `backend/internal/goals/handler.go` — ensure parent access works correctly: parent with matching familyID can call GET list. Explicitly block parent from POST create, PUT update, DELETE, and POST allocate endpoints (return 403 with "Only children can manage their own goals" message). Run tests to verify.
- [x] T045 [US6] Add goals summary to parent child view in `frontend/src/pages/ParentDashboard.tsx` — when viewing a child's details, fetch `getSavingsGoals(childId)` and display a "Savings Goals" section showing the child's active goals with compact GoalCards (read-only — no edit/delete/allocate buttons). Show goal names, progress rings, and amounts. If no goals, show "No savings goals yet" message. Show completed goals count if any ("3 goals achieved").

**Checkpoint**: Parents can see each child's goals from their dashboard without any ability to modify them.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Withdrawal goal impact warning (FR-016), edge case handling, and final validation

- [x] T046 Write tests for withdrawal goal impact in `backend/internal/balance/handler_test.go` — test that a withdrawal bringing balance below total goal allocations returns a `goal_impact_warning` error with affected_goals list when `confirm_goal_impact` is not true. Test that re-submitting with `confirm_goal_impact: true` processes the withdrawal and proportionally reduces goal allocations. Test that a withdrawal NOT impacting goals proceeds normally without warning.
- [x] T047 Implement `SavingsGoalStore.ReduceGoalsProportionally` in `backend/internal/store/savings_goal.go` — given a child_id and total_to_release int64, reduce each active goal's saved_cents proportionally (each goal loses `saved_cents * total_to_release / total_saved` rounded, with remainder adjustment on last goal). Record de-allocation entries in goal_allocations for each affected goal. All within a single DB transaction.
- [x] T048 Update `HandleWithdraw` in `backend/internal/balance/handler.go` — before processing withdrawal, compute available_balance after withdrawal. If it would go negative (meaning goals would be impacted): check for `confirm_goal_impact: true` in request body. If not confirmed, return 409 with `goal_impact_warning` error including list of affected goals with current and projected saved_cents. If confirmed, call `ReduceGoalsProportionally` after the withdrawal succeeds. Add `confirm_goal_impact` field to the withdrawal request struct.
- [x] T049 Run full backend test suite (`cd backend && go test -p 1 ./...`) and fix any failures
- [x] T050 Run frontend type check and build (`cd frontend && npx tsc --noEmit && npm run build`) and fix any errors

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS all user stories
- **User Stories (Phases 3–8)**: All depend on Phase 2 completion
  - **US1 (Phase 3)**: No dependencies on other stories — MVP
  - **US2 (Phase 4)**: Depends on US1 (needs goals to exist to allocate toward)
  - **US3 (Phase 5)**: Depends on US1 (enhances components created in US1)
  - **US4 (Phase 6)**: Depends on US2 (completion triggered by allocation)
  - **US5 (Phase 7)**: Depends on US1 (needs goals to edit/delete)
  - **US6 (Phase 8)**: Depends on US1 (needs goals to view)
- **Polish (Phase 9)**: Depends on US2 completion (withdrawal warning relates to allocations)

### Independent Story Tracks (after Phase 2)

```
Phase 2 ──► US1 ──► US2 ──► US4
                ├──► US3
                ├──► US5
                └──► US6
```

- US3, US5, US6 can run in parallel after US1
- US4 must follow US2
- US2 must follow US1

### Within Each User Story

1. Tests MUST be written and FAIL before implementation
2. Store methods before handler methods
3. Handler methods before frontend components
4. Core components before page integration

### Parallel Opportunities

- **Phase 2**: T003, T004, T005, T006 all target different files — run in parallel
- **Phase 3**: T008 + T009 (tests in parallel), then T010 → T011 (sequential), T012 + T013 + T014 (parallel frontend components), then T015 → T016 (sequential)
- **Phase 4**: T017 + T018 (tests in parallel), then T019 → T020 → T021 → T022 (sequential backend), T023 + T024 (parallel frontend)
- **Phase 6**: T029 + T030 (tests in parallel)
- **Phase 7**: T036 + T037 (tests in parallel)

---

## Parallel Example: User Story 1

```bash
# Launch both test files in parallel:
Task T008: "Write store tests for Create, GetByID, ListByChild, CountActiveByChild"
Task T009: "Write handler tests for HandleCreate and HandleList"

# After tests written, implement store (must be before handler):
Task T010: "Implement SavingsGoalStore.Create, GetByID, ListByChild, CountActiveByChild"

# Then handler (depends on store):
Task T011: "Implement HandleCreate and HandleList"

# Launch frontend components in parallel (different files):
Task T012: "Create GoalProgressRing component"
Task T013: "Create GoalCard component"
Task T014: "Create GoalForm component"

# Then page integration (depends on components):
Task T015: "Create SavingsGoalsPage"
Task T016: "Add route and dashboard navigation link"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (migration + test infra)
2. Complete Phase 2: Foundational (store/handler/types/routes)
3. Complete Phase 3: User Story 1 (create + list goals)
4. **STOP and VALIDATE**: Child can create goals and see them listed
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Create goals → **MVP!**
3. Add US2 → Allocate funds + balance breakdown → Core feature complete
4. Add US3 → Visual polish + theme integration → Visually appealing
5. Add US4 → Goal completion + celebration → Reward loop complete
6. Add US5 → Edit/delete/de-allocate → Full goal management
7. Add US6 → Parent visibility → Parent oversight
8. Add Polish → Withdrawal warnings → Edge cases handled

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- TDD is mandatory per constitution — all test tasks must FAIL before implementation
- Money is always int64 cents — convert dollars to/from cents at UI boundary only
- Max 5 active goals enforced at application layer (CountActiveByChild check)
- Interest continues on total balance (balance_cents) — no changes to interest calculation
- Available balance = balance_cents - SUM(active goals' saved_cents) — computed, not stored
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
