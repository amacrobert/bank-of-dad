# Tasks: Account Balances

**Input**: Design documents from `/specs/002-account-balances/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.yaml

**Tests**: Included per project constitution (Test-First Development required).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `backend/internal/`, `backend/main.go`
- **Frontend**: `frontend/src/`
- **Tests**: Co-located with implementation (`*_test.go`, `*.test.tsx`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database schema changes and shared types

- [x] T001 Add balance_cents column to children table migration in backend/internal/store/db.go
- [x] T002 Create transactions table with index in backend/internal/store/db.go
- [x] T003 [P] Add Transaction model and types in backend/internal/store/transaction.go
- [x] T004 [P] Add TypeScript types (Transaction, BalanceResponse, etc.) in frontend/src/types.ts

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Store layer with atomic transaction operations - MUST complete before any user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests First

- [x] T005 Write failing tests for TransactionStore.Deposit() in backend/internal/store/transaction_test.go
- [x] T006 [P] Write failing tests for TransactionStore.Withdraw() in backend/internal/store/transaction_test.go
- [x] T007 [P] Write failing tests for TransactionStore.ListByChild() in backend/internal/store/transaction_test.go
- [x] T008 [P] Write failing tests for ChildStore.GetBalance() in backend/internal/store/child_test.go

### Implementation

- [x] T009 Implement TransactionStore.Deposit() with atomic balance update in backend/internal/store/transaction.go
- [x] T010 Implement TransactionStore.Withdraw() with insufficient funds check in backend/internal/store/transaction.go
- [x] T011 Implement TransactionStore.ListByChild() ordered by created_at DESC in backend/internal/store/transaction.go
- [x] T012 Add GetBalance() method to ChildStore in backend/internal/store/child.go
- [x] T013 Run all store tests and verify they pass

**Checkpoint**: Store layer ready - balance operations work atomically, tests pass

---

## Phase 3: User Story 1 - Parent Views All Children's Balances (Priority: P1) 🎯 MVP

**Goal**: Parents see all their children's current balances on dashboard

**Independent Test**: Log in as parent, verify all children display with balance amounts

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T014 [P] [US1] Write failing test for GET /api/children returning balances in backend/internal/family/handler_test.go
- [x] T015 [P] [US1] Write failing test for parent authorization (can't see other family's children) in backend/internal/family/handler_test.go

### Backend Implementation for User Story 1

- [x] T016 [US1] Modify ListChildren handler to include balance_cents in response in backend/internal/family/handler.go
- [x] T017 [US1] Update ChildListResponse to include balance_cents in backend/internal/family/handler.go
- [x] T018 [US1] Run US1 backend tests and verify they pass

### Frontend Implementation for User Story 1

- [x] T019 [P] [US1] Add getChildrenWithBalances API function in frontend/src/api.ts
- [x] T020 [P] [US1] Create BalanceDisplay component in frontend/src/components/BalanceDisplay.tsx
- [x] T021 [US1] Update ParentDashboard to show balance for each child in frontend/src/pages/ParentDashboard.tsx
- [x] T022 [US1] Handle empty state (no children) with guidance message in frontend/src/pages/ParentDashboard.tsx

**Checkpoint**: Parents can view all children's balances - MVP complete and testable

---

## Phase 4: User Story 2 - Parent Adds Money to Child's Account (Priority: P1)

**Goal**: Parents can deposit money into any child's account with optional note

**Independent Test**: Parent deposits $10, child's balance increases by $10, transaction recorded

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T023 [P] [US2] Write failing test for POST /api/children/{id}/deposit in backend/internal/balance/handler_test.go
- [ ] T024 [P] [US2] Write failing test for deposit validation (amount > 0, max amount) in backend/internal/balance/handler_test.go
- [ ] T025 [P] [US2] Write failing test for deposit authorization (parent only, own family) in backend/internal/balance/handler_test.go

### Backend Implementation for User Story 2

- [ ] T026 [US2] Create balance handler package structure in backend/internal/balance/handler.go
- [ ] T027 [US2] Implement HandleDeposit endpoint in backend/internal/balance/handler.go
- [ ] T028 [US2] Add amount validation (positive, max 99999999 cents) in backend/internal/balance/handler.go
- [ ] T029 [US2] Add note validation (max 500 chars, trim whitespace) in backend/internal/balance/handler.go
- [ ] T030 [US2] Register POST /api/children/{id}/deposit route with RequireParent middleware in backend/main.go
- [ ] T031 [US2] Run US2 backend tests and verify they pass

### Frontend Implementation for User Story 2

- [ ] T032 [P] [US2] Add deposit API function in frontend/src/api.ts
- [ ] T033 [P] [US2] Create DepositForm component in frontend/src/components/DepositForm.tsx
- [ ] T034 [US2] Add deposit UI to ParentDashboard (form appears when child selected) in frontend/src/pages/ParentDashboard.tsx
- [ ] T035 [US2] Handle deposit success (show new balance, clear form) in frontend/src/pages/ParentDashboard.tsx
- [ ] T036 [US2] Handle deposit errors (validation, server errors) in frontend/src/components/DepositForm.tsx

**Checkpoint**: Parents can deposit money - core banking feature functional

---

## Phase 5: User Story 3 - Parent Removes Money from Child's Account (Priority: P2)

**Goal**: Parents can withdraw money from child's account (cannot go negative)

**Independent Test**: Parent withdraws $15 from $50 balance, balance becomes $35

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T037 [P] [US3] Write failing test for POST /api/children/{id}/withdraw in backend/internal/balance/handler_test.go
- [ ] T038 [P] [US3] Write failing test for insufficient funds error in backend/internal/balance/handler_test.go
- [ ] T039 [P] [US3] Write failing test for withdraw to exactly $0.00 (should succeed) in backend/internal/balance/handler_test.go

### Backend Implementation for User Story 3

- [ ] T040 [US3] Implement HandleWithdraw endpoint in backend/internal/balance/handler.go
- [ ] T041 [US3] Add insufficient funds error response with current balance in backend/internal/balance/handler.go
- [ ] T042 [US3] Register POST /api/children/{id}/withdraw route with RequireParent middleware in backend/main.go
- [ ] T043 [US3] Run US3 backend tests and verify they pass

### Frontend Implementation for User Story 3

- [ ] T044 [P] [US3] Add withdraw API function in frontend/src/api.ts
- [ ] T045 [P] [US3] Create WithdrawForm component in frontend/src/components/WithdrawForm.tsx
- [ ] T046 [US3] Add withdraw UI to ParentDashboard (alongside deposit) in frontend/src/pages/ParentDashboard.tsx
- [ ] T047 [US3] Handle insufficient funds error with clear message in frontend/src/components/WithdrawForm.tsx

**Checkpoint**: Parents can withdraw money - complete parent banking controls

---

## Phase 6: User Story 4 - Child Views Their Balance and Transaction History (Priority: P2)

**Goal**: Children see their balance and full transaction history (read-only)

**Independent Test**: Child logs in, sees current balance and list of all transactions

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T048 [P] [US4] Write failing test for GET /api/children/{id}/balance in backend/internal/balance/handler_test.go
- [ ] T049 [P] [US4] Write failing test for GET /api/children/{id}/transactions in backend/internal/balance/handler_test.go
- [ ] T050 [P] [US4] Write failing test for child can only view own balance (not siblings) in backend/internal/balance/handler_test.go
- [ ] T051 [P] [US4] Write failing test for transactions ordered newest-first in backend/internal/balance/handler_test.go

### Backend Implementation for User Story 4

- [ ] T052 [US4] Implement HandleGetBalance endpoint in backend/internal/balance/handler.go
- [ ] T053 [US4] Implement HandleGetTransactions endpoint in backend/internal/balance/handler.go
- [ ] T054 [US4] Add authorization check (child can only view self, parent can view own children) in backend/internal/balance/handler.go
- [ ] T055 [US4] Register GET /api/children/{id}/balance route with RequireAuth middleware in backend/main.go
- [ ] T056 [US4] Register GET /api/children/{id}/transactions route with RequireAuth middleware in backend/main.go
- [ ] T057 [US4] Run US4 backend tests and verify they pass

### Frontend Implementation for User Story 4

- [ ] T058 [P] [US4] Add getBalance and getTransactions API functions in frontend/src/api.ts
- [ ] T059 [P] [US4] Create TransactionHistory component in frontend/src/components/TransactionHistory.tsx
- [ ] T060 [US4] Update ChildDashboard to fetch and display balance in frontend/src/pages/ChildDashboard.tsx
- [ ] T061 [US4] Update ChildDashboard to fetch and display transaction history in frontend/src/pages/ChildDashboard.tsx
- [ ] T062 [US4] Handle empty transaction history state in frontend/src/components/TransactionHistory.tsx
- [ ] T063 [US4] Ensure no edit/add/remove controls visible to child in frontend/src/pages/ChildDashboard.tsx

**Checkpoint**: Children can view their finances - complete user experience

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [ ] T064 Run all backend tests (go test ./...) and verify 100% pass
- [ ] T065 Run all frontend tests and verify pass
- [ ] T066 Manual test: Parent creates deposit, child views updated balance
- [ ] T067 Manual test: Parent creates withdrawal, verify insufficient funds handling
- [ ] T068 Verify all success criteria from spec.md are met
- [ ] T069 Run quickstart.md validation scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 and US2 are both P1 priority - complete before P2 stories
  - US3 and US4 are both P2 priority - can proceed after P1 stories
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational - No dependencies (but same parent UI)
- **User Story 3 (P2)**: Can start after Foundational - Shares balance handler with US2
- **User Story 4 (P2)**: Can start after Foundational - Uses TransactionStore from Foundational

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Backend before frontend (API must exist for frontend to call)
- Run tests after implementation to verify pass
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 1: T003 and T004 can run in parallel (backend types, frontend types)
- Phase 2: T005, T006, T007, T008 tests can run in parallel
- Phase 3+: All tests marked [P] can run in parallel within each story
- Different user stories CAN be worked in parallel after Foundational (if team allows)

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational tests together:
Task: "Write failing tests for TransactionStore.Deposit()"
Task: "Write failing tests for TransactionStore.Withdraw()"
Task: "Write failing tests for TransactionStore.ListByChild()"
Task: "Write failing tests for ChildStore.GetBalance()"
```

## Parallel Example: User Story 2

```bash
# Launch all US2 tests together:
Task: "Write failing test for POST /api/children/{id}/deposit"
Task: "Write failing test for deposit validation"
Task: "Write failing test for deposit authorization"

# After backend complete, launch frontend components together:
Task: "Add deposit API function in frontend/src/api.ts"
Task: "Create DepositForm component"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (migrations, types)
2. Complete Phase 2: Foundational (TransactionStore, tests)
3. Complete Phase 3: User Story 1 (parent views balances)
4. **STOP and VALIDATE**: Test US1 independently
5. Parents can now see children with $0.00 balances

### Incremental Delivery

1. Setup + Foundational → Store layer ready
2. Add User Story 1 → Parents see balances → Demo ready
3. Add User Story 2 → Parents can deposit → Core banking works
4. Add User Story 3 → Parents can withdraw → Complete parent controls
5. Add User Story 4 → Children see their finances → Feature complete

### Suggested Execution

For single developer, sequential priority order:
1. Phase 1 → Phase 2 → Phase 3 (US1) → Phase 4 (US2) → Phase 5 (US3) → Phase 6 (US4) → Phase 7

---

## Task Summary

| Phase | Story | Task Count | Test Tasks | Impl Tasks |
|-------|-------|------------|------------|------------|
| Setup | - | 4 | 0 | 4 |
| Foundational | - | 9 | 4 | 5 |
| US1 (P1) | Parent Views Balances | 9 | 2 | 7 |
| US2 (P1) | Parent Deposits | 14 | 3 | 11 |
| US3 (P2) | Parent Withdraws | 11 | 3 | 8 |
| US4 (P2) | Child Views History | 16 | 4 | 12 |
| Polish | - | 6 | 0 | 6 |
| **Total** | | **69** | **16** | **53** |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (Red-Green-Refactor per constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All monetary values use cents (integers) - never use float
