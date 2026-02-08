# Tasks: Interest Accrual

**Input**: Design documents from `/specs/005-interest-accrual/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md, quickstart.md

**Tests**: Included — constitution requires Test-First Development for all financial logic.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database migrations and shared data layer for interest accrual

- [x] T001 Add `interest_rate_bps` and `last_interest_at` columns to `children` table via migration in `backend/internal/store/sqlite.go`
- [x] T002 Update `transactions` table CHECK constraint to include `interest` type using table recreation pattern in `backend/internal/store/sqlite.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Store layer for interest operations — MUST be complete before any user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T003 Write store tests for `SetInterestRate` (set rate, update rate, set to zero, validation 0-10000 bps) in `backend/internal/store/interest_test.go`
- [x] T004 Write store tests for `ApplyInterest` (correct calculation, balance update, transaction creation, last_interest_at update, atomicity) in `backend/internal/store/interest_test.go`
- [x] T005 Write store tests for edge cases: zero balance skipped, zero rate skipped, sub-cent rounds to zero skipped, duplicate prevention (same month skipped) in `backend/internal/store/interest_test.go`
- [x] T006 Write store test for `ListDueForInterest` (returns children with rate > 0, balance > 0, last_interest_at not in current month) in `backend/internal/store/interest_test.go`

### Implementation

- [x] T007 Implement `InterestStore` with `SetInterestRate(childID int64, rateBps int)` method in `backend/internal/store/interest.go`
- [x] T008 Implement `ListDueForInterest()` method that finds children eligible for accrual (rate > 0, balance > 0, last_interest_at not current month) in `backend/internal/store/interest.go`
- [x] T009 Implement `ApplyInterest(childID int64, parentID int64, rateBps int)` method that atomically: calculates interest (`balance_cents * rate_bps / 12 / 10000`), inserts interest transaction, updates balance_cents, updates last_interest_at — in a single DB transaction in `backend/internal/store/interest.go`
- [x] T010 Run store tests and verify all pass: `cd backend && go test ./internal/store/... -run Interest -v`

**Checkpoint**: Store layer complete — all interest data operations work correctly

---

## Phase 3: User Story 1 — Parent Configures Interest Rate (Priority: P1) 🎯 MVP

**Goal**: Parents can set, update, and disable per-child interest rates via API and UI

**Independent Test**: Set interest rate via `PUT /api/children/{id}/interest-rate`, verify rate is persisted and returned in balance response

### Tests

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T011 [P] [US1] Write handler tests for `PUT /api/children/{id}/interest-rate`: success (200), validation error (400 for rate < 0 or > 10000), forbidden (403 wrong family), set to zero disables in `backend/internal/interest/handler_test.go`
- [x] T012 [P] [US1] Write test for enhanced balance response: verify `interest_rate_bps` and `interest_rate_display` fields appear in `GET /api/children/{id}/balance` response in `backend/internal/balance/handler_test.go`

### Implementation

- [x] T013 [US1] Implement `SetInterestRate` HTTP handler: parse request, validate rate 0-10000, verify parent owns child's family, call store, return response with `interest_rate_bps` and `interest_rate_display` in `backend/internal/interest/handler.go`
- [x] T014 [US1] Enhance balance handler to include `interest_rate_bps` and `interest_rate_display` fields in `GET /api/children/{id}/balance` response in `backend/internal/balance/handler.go`
- [x] T015 [US1] Wire up `PUT /api/children/{id}/interest-rate` route with parent auth middleware in `backend/main.go`
- [x] T016 [US1] Run handler tests and verify all pass: `cd backend && go test ./internal/interest/... ./internal/balance/... -v`

**Checkpoint**: Parents can set interest rates via API and see them in balance responses

---

## Phase 4: User Story 2 — Automatic Interest Accrual (Priority: P1)

**Goal**: Background scheduler automatically calculates and credits monthly interest

**Independent Test**: Set an interest rate, trigger scheduler processing, verify balance increased and interest transaction created

### Tests

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T017 [P] [US2] Write scheduler tests: `ProcessDue` finds eligible children and applies interest, skips children already accrued this month, skips zero-balance children, skips zero-rate children in `backend/internal/interest/scheduler_test.go`
- [x] T018 [P] [US2] Write scheduler test for partial failure: if one child fails, others still process successfully in `backend/internal/interest/scheduler_test.go`

### Implementation

- [x] T019 [US2] Implement `interest.Scheduler` struct with `Start(interval, stop)` and `ProcessDue()` methods following allowance scheduler pattern (goroutine + ticker + stop channel) in `backend/internal/interest/scheduler.go`
- [x] T020 [US2] Wire up interest scheduler in `backend/main.go`: create scheduler, start with interval (e.g., 1 hour), defer stop on shutdown
- [x] T021 [US2] Run all backend tests and verify pass: `cd backend && go test ./... -v`

**Checkpoint**: Interest automatically accrues monthly via background scheduler

---

## Phase 5: User Story 3 — Child Sees Interest Earnings (Priority: P2)

**Goal**: Interest transactions appear distinctly in child's transaction history

**Independent Test**: After interest accrual, child views transaction history and sees "Interest earned" entries distinguishable from deposits/withdrawals

### Implementation

- [x] T022 [US3] Add `interest` to transaction type union in `frontend/src/types.ts`
- [x] T023 [US3] Add interest transaction display (label: "Interest earned", distinct color/icon) to `frontend/src/components/TransactionHistory.tsx`

**Checkpoint**: Children see interest transactions labeled clearly in their history

---

## Phase 6: User Story 4 — Parent Views Interest Activity (Priority: P2)

**Goal**: Parents see interest rate on child cards and interest transactions in history

**Independent Test**: Parent views child dashboard and sees interest rate displayed; views transaction history and sees interest entries with rate info

### Implementation

- [x] T024 [US4] Show interest rate on parent's ManageChild view (fetched from balance endpoint) in `frontend/src/components/ManageChild.tsx`
- [x] T025 [US4] Create `InterestRateForm` component for parents to set/update interest rate with input validation (0-100%) in `frontend/src/components/InterestRateForm.tsx`
- [x] T026 [US4] Wire `InterestRateForm` to `PUT /api/children/{id}/interest-rate` endpoint with success/error feedback in `frontend/src/components/InterestRateForm.tsx`

**Checkpoint**: Parents can configure and monitor interest from the UI

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Validation, cleanup, and final verification

- [x] T027 Run full backend test suite: `cd backend && go test ./... -v`
- [x] T028 Run frontend checks: `cd frontend && npm run lint && npx tsc --noEmit && npx vite build`
- [x] T029 Run quickstart.md manual verification steps (set rate via API, trigger accrual, check transactions)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2
- **US2 (Phase 4)**: Depends on Phase 2 (independent of US1 for backend, but logically US1 provides the rate-setting mechanism)
- **US3 (Phase 5)**: Depends on Phase 2 (frontend only — can start once transaction type exists in DB)
- **US4 (Phase 6)**: Depends on US1 (needs interest rate API endpoint)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational — no dependencies on other stories
- **US2 (P1)**: Can start after Foundational — independent of US1 (scheduler uses store directly)
- **US3 (P2)**: Can start after Foundational — frontend only, needs `interest` type in types.ts
- **US4 (P2)**: Depends on US1 — needs the interest rate API endpoint to exist

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Store layer before handlers
- Handlers before route wiring
- Backend before frontend (for API-dependent UI)

### Parallel Opportunities

- T003, T004, T005, T006 can all be written in parallel (test file, no implementation yet)
- T011 and T012 can run in parallel (different test files)
- T017 and T018 can run in parallel (same test file but independent test functions)
- T022 and T023 can run in parallel with T024-T026 (different frontend files, different stories)
- US1 and US2 backend work can proceed in parallel after Phase 2
- US3 and US4 frontend work can proceed in parallel after US1 backend is done

---

## Parallel Example: Phase 2 Store Tests

```bash
# Write all store tests in parallel (same file, but independent test functions):
Task T003: "Store tests for SetInterestRate in backend/internal/store/interest_test.go"
Task T004: "Store tests for ApplyInterest in backend/internal/store/interest_test.go"
Task T005: "Store tests for edge cases in backend/internal/store/interest_test.go"
Task T006: "Store tests for ListDueForInterest in backend/internal/store/interest_test.go"
```

## Parallel Example: US1 + US2 Handler Tests

```bash
# Write handler tests in parallel (different files):
Task T011: "Handler tests for interest rate endpoint in backend/internal/interest/handler_test.go"
Task T012: "Balance response tests in backend/internal/balance/handler_test.go"
```

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1: Setup (migrations)
2. Complete Phase 2: Foundational (store layer with tests)
3. Complete Phase 3: US1 — Parent sets interest rate via API
4. Complete Phase 4: US2 — Scheduler auto-accrues interest
5. **STOP and VALIDATE**: Test end-to-end: set rate → scheduler runs → balance increases → transaction appears
6. Deploy/demo backend-only MVP

### Incremental Delivery

1. Setup + Foundational → Data layer ready
2. US1 (Rate config API) → Test independently → Backend can accept rates
3. US2 (Scheduler) → Test independently → Interest accrues automatically
4. US3 (Child transaction display) → Test independently → Children see interest
5. US4 (Parent UI) → Test independently → Parents manage rates from UI
6. Polish → Full validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires Test-First Development — write tests before implementation
- Interest calculation uses integer arithmetic only: `balance_cents * rate_bps / 12 / 10000`
- Use past dates in test fixtures to avoid time-dependent failures
- Follow existing patterns: `NewInterestStore(db)`, `s.db.Write`/`s.db.Read`, `writeJSON` helper
- Commit after each task or logical group
