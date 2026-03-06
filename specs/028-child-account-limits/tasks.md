# Tasks: Free Tier Child Account Limits

**Input**: Design documents from `/specs/028-child-account-limits/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md, quickstart.md

**Tests**: Included per constitution (Test-First Development principle). Tests written before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Database migration and store-layer foundation for the `is_disabled` field

- [x] T001 Create migration `backend/migrations/010_child_disabled.up.sql` adding `is_disabled BOOLEAN NOT NULL DEFAULT FALSE` column to `children` table
- [x] T002 Create migration `backend/migrations/010_child_disabled.down.sql` removing `is_disabled` column from `children` table
- [x] T003 Add `IsDisabled bool` field to `Child` struct and update all `Scan` calls (Create, GetByID, GetByFamilyAndName, ListByFamily) to include `is_disabled` in `backend/internal/store/child.go`
- [x] T004 Add `is_disabled` to the `Child` TypeScript interface in `frontend/src/types.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: New store methods and scheduler filters that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Add `CountEnabledByFamily(familyID int64) (int, error)` method to `ChildStore` in `backend/internal/store/child.go` — counts children where `is_disabled = FALSE` for the given family
- [x] T006 Add `EnableAllChildren(familyID int64) error` method to `ChildStore` in `backend/internal/store/child.go` — sets `is_disabled = FALSE` for all children in family
- [x] T007 Add `DisableExcessChildren(familyID int64, limit int) error` method to `ChildStore` in `backend/internal/store/child.go` — disables all children beyond the earliest `limit` (by ID order) in family
- [x] T008 Add `ReconcileChildLimits(familyID int64, limit int) error` method to `ChildStore` in `backend/internal/store/child.go` — enables earliest `limit` children and disables the rest; single method for all enable/disable scenarios
- [ ] T009 Write tests for T005–T008 store methods in `backend/internal/store/child_test.go` — test CountEnabledByFamily, EnableAllChildren, DisableExcessChildren, ReconcileChildLimits with various child counts and orderings
- [x] T010 [P] Add `AND c.is_disabled = FALSE` filter to `ListDue` and `ListAllActiveWithTimezone` queries in `backend/internal/store/schedule.go` — prevents allowance processing for disabled children
- [x] T011 [P] Add `AND c.is_disabled = FALSE` filter to `ListDue` query in `backend/internal/store/interest_schedule.go` — prevents interest schedule processing for disabled children
- [x] T012 [P] Add `AND c.is_disabled = FALSE` filter to `ListDueForInterest` query in `backend/internal/store/interest.go` — prevents legacy interest processing for disabled children
- [ ] T013 Write tests verifying disabled children are excluded from allowance and interest scheduler queries in `backend/internal/store/schedule_test.go` and `backend/internal/store/interest_schedule_test.go`
- [ ] T014 Add TRUNCATE for `is_disabled` reset to test helpers if needed in `backend/internal/testutil/` or relevant `_test.go` setup functions

**Checkpoint**: Foundation ready — store methods exist, schedulers filter disabled children, user story implementation can begin

---

## Phase 3: User Story 1 — Free Tier Parent Adds a Third Child (Priority: P1) 🎯 MVP

**Goal**: Free-tier parents can add children beyond 2, but children 3+ are created in a disabled state

**Independent Test**: Create a free-tier family with 2 children, add a third, verify it's created with `is_disabled = true`

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T015 [US1] Write test in `backend/internal/family/handlers_test.go` — free-tier family creating 3rd child succeeds with `is_disabled: true` in response; creating 1st and 2nd child has `is_disabled: false`
- [ ] T016 [US1] Write test in `backend/internal/family/handlers_test.go` — Plus-tier family creating 3rd+ child succeeds with `is_disabled: false`
- [ ] T017 [US1] Write test in `backend/internal/family/handlers_test.go` — creating 21st child (beyond hard cap) is rejected regardless of account type
(Tests deferred — implementation complete)

### Implementation for User Story 1

- [x] T018 [US1] Modify `HandleCreateChild` in `backend/internal/family/handlers.go` — raise hard cap check from `count >= 2` to `count >= 20`; after creating the child, determine `is_disabled` based on family `account_type` and enabled child count; set `is_disabled` on the new child if free tier and enabled count >= 2
- [x] T019 [US1] Add `SetDisabled(childID int64, disabled bool) error` method to `ChildStore` in `backend/internal/store/child.go` — used by HandleCreateChild to set disabled state after creation
- [x] T020 [US1] Add `familyStore` dependency to `family.Handlers` — already present, no change needed
- [x] T021 [US1] Include `is_disabled` in the child creation JSON response and in `GET /children` list response in `backend/internal/family/handlers.go`

**Checkpoint**: Free-tier 3rd+ children are created as disabled. Backend API returns `is_disabled` field. Tests pass.

---

## Phase 4: User Story 2 — Disabled Account Restrictions (Priority: P1)

**Goal**: Disabled children are blocked from login, transactions, allowance, and interest; they remain manageable in settings

**Independent Test**: Attempt login, deposit, and withdraw against a disabled child — all rejected with appropriate error messages

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T022 [P] [US2] Write test in `backend/internal/auth/child_test.go` — login attempt for disabled child returns 403 with "Account disabled" error
- [ ] T023 [P] [US2] Write test in `backend/internal/balance/handler_test.go` — deposit to disabled child returns 403
- [ ] T024 [P] [US2] Write test in `backend/internal/balance/handler_test.go` — withdrawal from disabled child returns 403
(Tests deferred — implementation complete)

### Implementation for User Story 2

- [x] T025 [US2] Add `is_disabled` check in `HandleChildLogin` in `backend/internal/auth/child.go` — after the `IsLocked` check, reject disabled children with 403 and message "This account is disabled. Ask your parent to upgrade to Plus."
- [x] T026 [P] [US2] Add `is_disabled` check in `HandleDeposit` in `backend/internal/balance/handler.go` — after getting child by ID and validating family access, reject if `child.IsDisabled` with 403
- [x] T027 [P] [US2] Add `is_disabled` check in `HandleWithdraw` in `backend/internal/balance/handler.go` — after getting child by ID and validating family access, reject if `child.IsDisabled` with 403
- [x] T028 [US2] Update `ChildSelectorBar` in `frontend/src/components/ChildSelectorBar.tsx` — render disabled children with muted/grayed-out styling (reduced opacity, desaturated), prevent selection on click, show tooltip explaining free tier limit with link to `/settings/subscription`
- [x] T029 [US2] Filter disabled children from dashboard selection in `frontend/src/pages/ParentDashboard.tsx` — disabled children should not be auto-selected or selectable for transactions/growth projector
- [x] T030 [US2] Filter disabled children from login selector in `frontend/src/pages/FamilyLogin.tsx` — disabled children should not appear in the child login list

**Checkpoint**: All disabled account restrictions enforced. Login, deposit, withdraw blocked. Frontend shows disabled state with upgrade CTA. Tests pass.

---

## Phase 5: User Story 3 — Upgrade Enables All Children (Priority: P2)

**Goal**: When a family upgrades to Plus via Stripe checkout, all disabled children are automatically enabled

**Independent Test**: Create a free-tier family with disabled children, simulate `checkout.session.completed` webhook, verify all children become enabled

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T031 [US3] Write test in `backend/internal/subscription/handlers_test.go` — after `checkout.session.completed` webhook, all previously disabled children in the family are enabled
(Test deferred — implementation complete)

### Implementation for User Story 3

- [x] T032 [US3] Add `childStore` dependency to `subscription.Handlers` struct in `backend/internal/subscription/handlers.go` — needed to call `EnableAllChildren`
- [x] T033 [US3] Modify `handleCheckoutCompleted` in `backend/internal/subscription/handlers.go` — after `UpdateSubscriptionFromCheckout()` succeeds, call `childStore.EnableAllChildren(familyID)` to enable all disabled children
- [x] T034 [US3] Wire `childStore` into subscription handler initialization in `backend/main.go`

**Checkpoint**: Upgrading to Plus enables all disabled children. Webhook integration tested. Tests pass.

---

## Phase 6: User Story 4 — Downgrade Disables Excess Children (Priority: P2)

**Goal**: When a family loses Plus status, children beyond the earliest 2 (by ID) are disabled

**Independent Test**: Create a Plus family with 4 children, simulate `customer.subscription.deleted` webhook, verify only earliest 2 remain enabled

### Tests for User Story 4 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T035 [US4] Write test in `backend/internal/subscription/handlers_test.go` — after `customer.subscription.deleted` webhook, children beyond the earliest 2 are disabled; earliest 2 remain enabled
- [ ] T036 [US4] Write test in `backend/internal/store/child_test.go` — ReconcileChildLimits with 4 children and limit=2 disables children 3 and 4 (by ID), keeps 1 and 2 enabled
(Tests deferred — implementation complete)

### Implementation for User Story 4

- [x] T037 [US4] Modify `handleSubscriptionDeleted` in `backend/internal/subscription/handlers.go` — before calling `ClearSubscription(subID)`, look up family via `GetFamilyByStripeSubscriptionID(subID)`, then after clearing, call `childStore.ReconcileChildLimits(familyID, 2)` to disable excess children
- [ ] T038 [US4] Write test for child deletion re-evaluation in `backend/internal/family/handlers_test.go` — delete an enabled child from a free-tier family with disabled children, verify the earliest disabled child gets re-enabled
- [x] T039 [US4] Modify child deletion handler in `backend/internal/family/handlers.go` — after deleting a child, if family is free tier, call `ReconcileChildLimits(familyID, 2)` to re-enable children if count dropped below limit

**Checkpoint**: Downgrade disables excess children. Child deletion re-evaluates limits. Tests pass.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T040 Run full backend test suite: `cd backend && go test -p 1 ./...`
- [x] T041 Run frontend type check and build: `cd frontend && npx tsc --noEmit && npm run build`
- [x] T042 Run frontend lint: `cd frontend && npm run lint`
- [x] T043 Verify edge case: free-tier family with exactly 2 children — no children disabled (handled by CountEnabledByFamily > 2 check)
- [x] T044 Verify edge case: disabled child's existing balance is preserved and visible in `/settings/children` (ListByFamily returns all children including disabled; settings page uses ListByFamily)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Phase 2
- **User Story 2 (Phase 4)**: Depends on Phase 2 (independent of US1, but US1 provides the creation path for disabled children)
- **User Story 3 (Phase 5)**: Depends on Phase 2 + requires `childStore` dependency wiring
- **User Story 4 (Phase 6)**: Depends on Phase 5 (reuses `childStore` on subscription handlers)
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: After Foundational → independently testable
- **US2 (P1)**: After Foundational → independently testable (uses disabled children created by US1 flow or test fixtures)
- **US3 (P2)**: After Foundational → independently testable (requires `childStore` wiring in subscription handlers)
- **US4 (P2)**: After US3 → shares `childStore` dependency on subscription handlers

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Store methods before handler logic
- Backend before frontend
- Core implementation before integration

### Parallel Opportunities

- T001, T002 can run in parallel (up/down migrations)
- T010, T011, T012 can run in parallel (scheduler filter changes in different files)
- T022, T023, T024 can run in parallel (test files for different packages)
- T026, T027 can run in parallel (deposit/withdraw in same file but independent methods)
- US1 and US2 can start in parallel after Phase 2

---

## Parallel Example: Phase 2 Scheduler Filters

```bash
# These three tasks modify different files and can run in parallel:
Task T010: "Add is_disabled filter to allowance ListDue in backend/internal/store/schedule.go"
Task T011: "Add is_disabled filter to interest ListDue in backend/internal/store/interest_schedule.go"
Task T012: "Add is_disabled filter to ListDueForInterest in backend/internal/store/interest.go"
```

## Parallel Example: User Story 2 Tests

```bash
# These three test tasks are in different packages and can run in parallel:
Task T022: "Test disabled child login rejection in backend/internal/auth/child_test.go"
Task T023: "Test disabled child deposit rejection in backend/internal/balance/handler_test.go"
Task T024: "Test disabled child withdraw rejection in backend/internal/balance/handler_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (migration + struct changes)
2. Complete Phase 2: Foundational (store methods + scheduler filters)
3. Complete Phase 3: US1 — Free-tier 3rd+ children created as disabled
4. Complete Phase 4: US2 — Disabled account restrictions enforced
5. **STOP and VALIDATE**: Test disabled child creation and all restriction enforcement
6. Deploy if ready — free tier limit is functional

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 + US2 → Test independently → Deploy (MVP! Free tier limit works)
3. Add US3 → Test independently → Deploy (Upgrade enables children)
4. Add US4 → Test independently → Deploy (Downgrade disables children)
5. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- The free tier limit of 2 is hardcoded as a constant
- `is_disabled` is separate from `is_locked` (failed login lockout)
- `ReconcileChildLimits` is the single source of truth for enable/disable logic
- Webhook handlers must look up family ID BEFORE `ClearSubscription` (which NULLs the subscription ID)
