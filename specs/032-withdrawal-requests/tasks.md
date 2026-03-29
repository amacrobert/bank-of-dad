# Tasks: Withdrawal Requests

**Input**: Design documents from `/specs/032-withdrawal-requests/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included per constitution (Test-First Development principle).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Database migration, model definitions, and type scaffolding

- [x] T001 [P] Create migration in backend/migrations/013_withdrawal_requests.up.sql — withdrawal_requests table with partial unique index, update transaction_type check constraint to include 'withdrawal_request'
- [x] T002 [P] Create rollback migration in backend/migrations/013_withdrawal_requests.down.sql — drop withdrawal_requests table, restore original transaction_type check constraint
- [x] T003 Create WithdrawalRequest model and status constants in backend/models/withdrawal_request.go — struct with GORM tags matching data-model.md schema
- [x] T004 Add TransactionTypeWithdrawalRequest constant ("withdrawal_request") to backend/models/transaction.go
- [x] T005 [P] Add WithdrawalRequest, WithdrawalRequestStatus, and request/response types to frontend/src/types.ts

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Repository, handler skeleton, routes, and API client — MUST complete before user story work

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Create WithdrawalRequestRepo struct and constructor in backend/repositories/withdrawal_request_repo.go — NewWithdrawalRequestRepo(db *gorm.DB), GetByID method, PendingCountByFamily method
- [x] T007 Create withdrawal handler struct in backend/internal/withdrawal/handler.go — NewHandler with WithdrawalRequestRepo, TransactionRepo, ChildRepo, BalanceStore dependencies; writeJSON helper
- [x] T008 Register all withdrawal request routes in backend/main.go — child endpoints (POST/GET /api/child/withdrawal-requests, POST .../cancel), parent endpoints (GET /api/withdrawal-requests, POST .../approve, POST .../deny, GET .../pending/count) with requireAuth/requireParent middleware
- [x] T009 Add withdrawal request API functions to frontend/src/api.ts — submitWithdrawalRequest, getChildWithdrawalRequests, cancelWithdrawalRequest, getWithdrawalRequests, approveWithdrawalRequest, denyWithdrawalRequest, getPendingWithdrawalRequestCount

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — Child Requests a Withdrawal (Priority: P1) 🎯 MVP

**Goal**: Children can submit a withdrawal request with an amount and reason, see it in pending state

**Independent Test**: Log in as child, submit a withdrawal request, verify it appears as pending with correct amount and reason

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T010 [P] [US1] Write repo tests for Create method in backend/repositories/withdrawal_request_repo_test.go — test successful creation, duplicate pending rejection, insufficient balance scenarios
- [x] T011 [P] [US1] Write handler tests for HandleSubmitRequest in backend/internal/withdrawal/handler_test.go — test valid submission (201), insufficient balance (422), already pending (409), disabled account (403), invalid input (400)

### Implementation for User Story 1

- [x] T012 [US1] Implement Create method in backend/repositories/withdrawal_request_repo.go — insert with pending status, check for existing pending request (return conflict error), validate available balance via child balance lookup
- [x] T013 [US1] Implement HandleSubmitRequest in backend/internal/withdrawal/handler.go — parse request body, validate amount (>0, <=99999999) and reason (1-500 chars), check child auth context, check account not disabled, call repo Create, return 201 with withdrawal_request JSON
- [x] T014 [P] [US1] Create WithdrawalRequestForm component in frontend/src/components/WithdrawalRequestForm.tsx — amount input, reason textarea (required, max 500 chars), submit button disabled when balance is $0 or request pending, loading/error states; follows DepositForm/WithdrawForm pattern
- [x] T015 [US1] Integrate withdrawal request into frontend/src/pages/ChildDashboard.tsx — add "Request Withdrawal" button, show pending request status card if one exists, fetch pending request on mount, refresh after submission

**Checkpoint**: Child can submit a withdrawal request and see it pending. Run backend tests to verify.

---

## Phase 4: User Story 2 — Parent Reviews and Approves/Denies (Priority: P1) 🎯 MVP

**Goal**: Parents can see pending requests, approve (creating a transaction and deducting balance) or deny (with optional reason)

**Independent Test**: With a pending request in the system, log in as parent, view pending request, approve it and verify balance deduction; separately deny one and verify denial reason visible

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T016 [P] [US2] Write repo tests for Approve and Deny methods in backend/repositories/withdrawal_request_repo_test.go — test approve updates status/parent/transaction, deny records reason, invalid status transitions rejected
- [x] T017 [P] [US2] Write handler tests for HandleApprove and HandleDeny in backend/internal/withdrawal/handler_test.go — test approve with sufficient balance (200), approve with insufficient balance (422), approve with goal impact (409 then 200 with confirm), deny with reason (200), deny without reason (200), approve non-pending (409), approve disabled account (422)

### Implementation for User Story 2

- [x] T018 [US2] Implement Approve method in backend/repositories/withdrawal_request_repo.go — pending→approved status transition using RowsAffected==0 guard, set reviewed_by_parent_id, reviewed_at, transaction_id
- [x] T019 [US2] Implement Deny method in backend/repositories/withdrawal_request_repo.go — pending→denied status transition, set reviewed_by_parent_id, reviewed_at, denial_reason
- [x] T020 [US2] Implement ListByFamily method in backend/repositories/withdrawal_request_repo.go — query by family_id with optional status filter, join child name for display, order by created_at desc
- [x] T021 [US2] Implement HandleApprove in backend/internal/withdrawal/handler.go — verify parent auth, verify request belongs to family, check child available balance, check account not disabled, handle goal-impact warning (reuse balance handler pattern with confirm_goal_impact flag), create withdrawal_request transaction via TransactionRepo, call repo Approve with transaction ID, return updated request + new_balance_cents
- [x] T022 [US2] Implement HandleDeny in backend/internal/withdrawal/handler.go — verify parent auth, verify request belongs to family, parse optional denial reason (max 500 chars), call repo Deny, return updated request
- [x] T023 [US2] Implement HandlePendingCount in backend/internal/withdrawal/handler.go — verify parent auth, call repo PendingCountByFamily, return count JSON
- [x] T024 [P] [US2] Create PendingWithdrawalRequests component in frontend/src/components/PendingWithdrawalRequests.tsx — displays list of pending requests grouped by child, each with amount, reason, date, approve/deny buttons; deny opens modal for optional reason; approve shows confirmation; handles goal-impact warning flow
- [x] T025 [US2] Integrate pending requests into frontend/src/components/ManageChild.tsx — show pending request card with approve/deny actions alongside existing deposit/withdraw buttons, refresh balance on approve
- [x] T026 [US2] Add pending request badge to parent dashboard in frontend/src/pages/ParentDashboard.tsx — fetch pending count on mount, show indicator/badge when count > 0

**Checkpoint**: Full request→review→approve/deny flow works end-to-end. Run full backend test suite: `cd backend && go test -p 1 ./...`

---

## Phase 5: User Story 3 — Child Cancels a Pending Request (Priority: P2)

**Goal**: Children can cancel their own pending requests before parent acts on them

**Independent Test**: Create a pending request as child, cancel it, verify status changes to cancelled and it no longer appears in parent's pending queue

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T027 [P] [US3] Write repo tests for Cancel method in backend/repositories/withdrawal_request_repo_test.go — test successful cancel, cancel non-pending rejected, cancel by wrong child rejected
- [x] T028 [P] [US3] Write handler tests for HandleCancelRequest in backend/internal/withdrawal/handler_test.go — test cancel pending (200), cancel non-pending (409), cancel other child's request (404)

### Implementation for User Story 3

- [x] T029 [US3] Implement Cancel method in backend/repositories/withdrawal_request_repo.go — pending→cancelled status transition with child_id ownership check, RowsAffected==0 guard
- [x] T030 [US3] Implement HandleCancelRequest in backend/internal/withdrawal/handler.go — verify child auth, parse request ID from path, call repo Cancel, return updated request
- [x] T031 [US3] Add cancel button to pending request display in frontend/src/pages/ChildDashboard.tsx — show cancel option only for pending requests, confirm before cancelling, refresh view after cancel

**Checkpoint**: Child can cancel pending requests. Verify cancelled requests don't appear in parent's pending list.

---

## Phase 6: User Story 4 — Request History (Priority: P2)

**Goal**: Both parents and children can view past withdrawal requests with outcomes

**Independent Test**: Create requests in various states (approved, denied, cancelled), verify both child and parent views show complete history with correct statuses and details

### Tests for User Story 4 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T032 [P] [US4] Write repo tests for ListByChild method in backend/repositories/withdrawal_request_repo_test.go — test returns all statuses, status filter works, ordered by created_at desc
- [x] T033 [P] [US4] Write handler tests for HandleListRequests (child and parent variants) in backend/internal/withdrawal/handler_test.go — test child sees own requests only, parent sees family requests, status filter query param works

### Implementation for User Story 4

- [x] T034 [US4] Implement ListByChild method in backend/repositories/withdrawal_request_repo.go — query by child_id with optional status filter, order by created_at desc
- [x] T035 [US4] Implement HandleListRequests for child in backend/internal/withdrawal/handler.go — verify child auth, parse optional status query param, call repo ListByChild, return withdrawal_requests array
- [x] T036 [US4] Implement HandleListRequests for parent in backend/internal/withdrawal/handler.go — verify parent auth, parse optional status and child_id query params, call repo ListByFamily, return withdrawal_requests array with child names
- [x] T037 [P] [US4] Create WithdrawalRequestCard component in frontend/src/components/WithdrawalRequestCard.tsx — display request with amount, reason, status badge (color-coded: pending=honey, approved=forest, denied=terracotta, cancelled=bark-light), denial reason if denied, timestamps
- [x] T038 [US4] Add request history section to frontend/src/pages/ChildDashboard.tsx — list past requests using WithdrawalRequestCard, show empty state if no history
- [x] T039 [US4] Add request history to parent view in frontend/src/components/ManageChild.tsx — show past requests for the selected child using WithdrawalRequestCard

**Checkpoint**: Full history visible for both roles. All user stories complete.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, validation hardening, and final verification

- [x] T040 Add disabled account validation to submit and approve flows in backend/internal/withdrawal/handler.go — check child.IsDisabled before creating request and before approving
- [x] T041 Ensure withdrawal_request transactions display correctly in existing transaction list in frontend/src/components/ — update transaction display logic to show "Withdrawal Request" label for withdrawal_request type
- [x] T042 Run full backend test suite: `cd backend && go test -p 1 ./...`
- [x] T043 Run frontend validation: `cd frontend && npx tsc --noEmit && npm run build && npm run lint`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — child submit flow
- **US2 (Phase 4)**: Depends on Phase 2 + Phase 3 (needs pending requests to exist for testing)
- **US3 (Phase 5)**: Depends on Phase 2 + Phase 3 (needs pending requests to cancel)
- **US4 (Phase 6)**: Depends on Phase 2 (list endpoints are independent but benefit from US1+US2 data)
- **Polish (Phase 7)**: Depends on all user story phases complete

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational — no other story dependencies
- **US2 (P1)**: Requires US1 (needs pending requests to approve/deny) — forms the MVP together with US1
- **US3 (P2)**: Requires US1 (needs pending requests to cancel) — independent of US2
- **US4 (P2)**: Can start after Foundational — list endpoints are standalone; frontend integration benefits from US1/US2 components

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Repo methods before handler methods
- Backend before frontend
- Core implementation before integration into existing pages

### Parallel Opportunities

- T001 + T002 + T005 can run in parallel (different files)
- T010 + T011 can run in parallel (different test files)
- T016 + T017 can run in parallel (different test files)
- T027 + T028 can run in parallel (different test files)
- T032 + T033 can run in parallel (different test files)
- T014 can run in parallel with T012/T013 (frontend vs backend)
- T024 can run in parallel with T018-T023 (frontend component vs backend implementation)
- T037 can run in parallel with T034-T036 (frontend component vs backend implementation)

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (migration, models, types)
2. Complete Phase 2: Foundational (repo, handler, routes, API client)
3. Complete Phase 3: US1 — Child can submit requests
4. Complete Phase 4: US2 — Parent can approve/deny requests
5. **STOP and VALIDATE**: Full request→review→outcome flow works end-to-end
6. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 + US2 → Test end-to-end → Deploy/Demo (**MVP!**)
3. Add US3 → Child can cancel → Deploy/Demo
4. Add US4 → Full history views → Deploy/Demo
5. Polish → Edge cases, validation, cleanup → Final deploy

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires TDD — all test tasks must be completed and fail before implementation
- Money is int64 cents throughout — follow existing patterns
- Use RowsAffected==0 pattern from chore system for status transition guards
- Reuse goal-impact warning flow from existing balance handler for approve
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
