# Tasks: GORM Backend Refactor

**Input**: Design documents from `/specs/029-gorm-backend-refactor/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per constitution (Test-First Development principle). Repository tests are written before implementation.

**Organization**: Tasks grouped by user story. US1→US2→US3→US4 are sequential dependencies (models before repos, repos before handler migration, migration before store removal).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add GORM dependencies and create directory structure

- [x] T001 Add GORM dependencies to backend/go.mod: `go get gorm.io/gorm gorm.io/driver/postgres`
- [x] T000 Create directory structure: `backend/models/` and `backend/repositories/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: GORM connection setup and test infrastructure that MUST be complete before any user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T000 Create GORM connection setup in backend/repositories/db.go — implement `Open(dsn string) (*gorm.DB, error)` that wraps `gorm.Open()` with postgres driver, extracts underlying `*sql.DB` to run existing `golang-migrate` migrations, and returns `*gorm.DB`. Port migration logic from backend/internal/store/postgres.go. Do NOT use GORM AutoMigrate.
- [x] T000 Create shared test helper in backend/repositories/test_helpers_test.go — implement `testDB()` returning `*gorm.DB` connected to `bankofdad_test` database via `TEST_DATABASE_URL` env var. Include TRUNCATE helper that clears all tables with `RESTART IDENTITY CASCADE`. Reference existing pattern in backend/internal/store/store_test_helpers_test.go.

**Checkpoint**: Foundation ready — model and repository implementation can begin

---

## Phase 3: User Story 1 — Define GORM Models for All Entities (Priority: P1) 🎯 MVP

**Goal**: All 11 database entities represented as GORM model structs in `backend/models/`, matching the existing PostgreSQL schema exactly.

**Independent Test**: Import the `models` package, confirm all structs compile with correct GORM tags, field types match DB columns, and associations are defined.

### Implementation for User Story 1

- [x] T000 [P] [US1] Create Family model in backend/models/family.go — define `Family` struct with all fields from data-model.md (ID, Slug, Timezone, AccountType, Stripe fields, BankName, CreatedAt). Include GORM tags for primaryKey, uniqueIndex, not null, defaults. Add `HasMany` associations for Parents and Children. Set `TableName()` to "families".
- [x] T000 [P] [US1] Create Parent model in backend/models/parent.go — define `Parent` struct (ID, GoogleID, Email, DisplayName, FamilyID, CreatedAt). Include GORM tags and `BelongsTo` Family association.
- [x] T000 [P] [US1] Create Child model in backend/models/child.go — define `Child` struct (ID, FamilyID, FirstName, PasswordHash, IsLocked, IsDisabled, FailedLoginAttempts, BalanceCents int64, InterestRateBps, LastInterestAt, Avatar, Theme, CreatedAt, UpdatedAt). Use composite `uniqueIndex:idx_family_child` on FamilyID+FirstName. Add `BelongsTo` Family and `HasMany` associations.
- [x] T000 [P] [US1] Create Transaction model in backend/models/transaction.go — define `Transaction` struct (ID, ChildID, ParentID, AmountCents int64, TransactionType, Note, ScheduleID, CreatedAt). Include `BelongsTo` associations for Child, Parent, and optional AllowanceSchedule.
- [x] T000 [P] [US1] Create AllowanceSchedule model in backend/models/allowance_schedule.go — define `AllowanceSchedule` struct (ID, ChildID, ParentID, AmountCents int64, Frequency, DayOfWeek *int, DayOfMonth *int, Note, Status, NextRunAt, CreatedAt, UpdatedAt). Include `BelongsTo` Child/Parent and `HasMany` Transactions.
- [x] T00 [P] [US1] Create InterestSchedule model in backend/models/interest_schedule.go — define `InterestSchedule` struct (ID, ChildID uniqueIndex, ParentID, Frequency, DayOfWeek *int, DayOfMonth *int, Status, NextRunAt, CreatedAt, UpdatedAt). Include `BelongsTo` Child/Parent.
- [x] T00 [P] [US1] Create RefreshToken model in backend/models/refresh_token.go — define `RefreshToken` struct (ID, TokenHash uniqueIndex, UserType, UserID, FamilyID, ExpiresAt, CreatedAt). No GORM associations (polymorphic user ref).
- [x] T00 [P] [US1] Create AuthEvent model in backend/models/auth_event.go — define `AuthEvent` struct (ID, EventType, UserType, UserID *int64, FamilyID *int64, IPAddress, Details *string, CreatedAt). No GORM associations (audit log).
- [x] T00 [P] [US1] Create StripeWebhookEvent model in backend/models/webhook_event.go — define `StripeWebhookEvent` struct with `StripeEventID string` as primaryKey (not SERIAL), EventType, ProcessedAt. Set `TableName()` to "stripe_webhook_events".
- [x] T00 [P] [US1] Create SavingsGoal model in backend/models/savings_goal.go — define `SavingsGoal` struct (ID, ChildID, Name, TargetCents int64, SavedCents int64, Emoji *string, Status, CompletedAt *time.Time, CreatedAt, UpdatedAt). Include `BelongsTo` Child and `HasMany` GoalAllocations.
- [x] T00 [P] [US1] Create GoalAllocation model in backend/models/goal_allocation.go — define `GoalAllocation` struct (ID, GoalID, ChildID, AmountCents int64, CreatedAt). Include `BelongsTo` SavingsGoal and Child.

**Checkpoint**: All 11 models compile. `cd backend && go build ./models/` succeeds. Models match existing schema per data-model.md.

---

## Phase 4: User Story 2 — Create Repository Layer for Core Entities (Priority: P1)

**Goal**: Repository structs in `backend/repositories/` that replace all `internal/store/` methods with GORM-based equivalents.

**Independent Test**: Integration tests for each repository pass against test database with identical behavior to existing store tests.

### Tests for User Story 2 ⚠️

> **NOTE: Write tests FIRST per TDD. Reference existing store tests for expected behavior. Tests must FAIL before implementation.**

- [x] T00 [P] [US2] Write tests for FamilyRepo in backend/repositories/family_repo_test.go — cover Create, GetByID, GetBySlug, SlugExists, SuggestSlugs, UpdateTimezone, UpdateBankName, Stripe-related methods, UpdateAccountType. Mirror test cases from backend/internal/store/family_test.go. Use testDB() helper.
- [x] T00 [P] [US2] Write tests for ParentRepo in backend/repositories/parent_repo_test.go — cover GetOrCreate (new + existing), GetByID, SetFamily, GetByFamilyID. Mirror backend/internal/store/parent_test.go.
- [x] T00 [P] [US2] Write tests for ChildRepo in backend/repositories/child_repo_test.go — cover Create (including duplicate name error), GetByID, GetByName, ListByFamily, GetBalance, all Update methods, IncrementFailedLogins, Lock, ResetFailedLogins, SetDisabled, CountByFamily, DeleteAll. Mirror backend/internal/store/child_test.go.
- [x] T00 [P] [US2] Write tests for TransactionRepo in backend/repositories/transaction_repo_test.go — cover Deposit (balance update + transaction record), Withdraw (including insufficient funds), ListByChild (with pagination), GetMonthlySummary, CreateAllowanceTransaction, CreateInterestTransaction. Mirror backend/internal/store/transaction_test.go.
- [x] T00 [P] [US2] Write tests for ScheduleRepo in backend/repositories/schedule_repo_test.go — cover Create, GetByID, ListByChild, ListByFamily, Update, UpdateStatus, Delete, ListDue (with past/future dates), UpdateNextRun. Mirror backend/internal/store/schedule_test.go.
- [x] T00 [P] [US2] Write tests for InterestScheduleRepo in backend/repositories/interest_schedule_repo_test.go — cover Create, GetByChildID, ListByFamily, Update, UpdateStatus, Delete, ListDue, UpdateNextRun. Mirror backend/internal/store/interest_schedule_test.go.
- [x] T00 [P] [US2] Write tests for InterestRepo in backend/repositories/interest_repo_test.go — cover AccrueInterest (balance update + transaction + interest calculation from basis points), UpdateLastInterestAt. Mirror backend/internal/store/interest_test.go.
- [x] T00 [P] [US2] Write tests for RefreshTokenRepo in backend/repositories/refresh_token_repo_test.go — cover Create, GetByHash, DeleteByHash, DeleteByUser, DeleteExpired. Mirror backend/internal/store/refresh_token_test.go.
- [x] T00 [P] [US2] Write tests for AuthEventRepo in backend/repositories/auth_event_repo_test.go — cover Log, ListByFamily. Mirror backend/internal/store/auth_event_test.go.
- [x] T00 [P] [US2] Write tests for WebhookEventRepo in backend/repositories/webhook_event_repo_test.go — cover Exists, Create (including idempotency). Mirror backend/internal/store/webhook_event_test.go.
- [x] T00 [P] [US2] Write tests for SavingsGoalRepo in backend/repositories/savings_goal_repo_test.go — cover Create, GetByID, ListByChild, Update, Delete, Allocate (with goal saved_cents update), Deallocate. Mirror backend/internal/store/savings_goal_test.go.
- [x] T00 [P] [US2] Write tests for GoalAllocationRepo in backend/repositories/goal_allocation_repo_test.go — cover ListByGoal, ListByChild. Mirror backend/internal/store/savings_goal_test.go (allocation tests).

### Implementation for User Story 2

- [x] T00 [P] [US2] Implement FamilyRepo in backend/repositories/family_repo.go — all methods per contracts/repository-interfaces.md. Use GORM builder for CRUD, `db.Raw()` for SuggestSlugs. Handle `gorm.ErrRecordNotFound` → `(nil, nil)`. Handle duplicate slug errors with user-friendly message.
- [x] T00 [P] [US2] Implement ParentRepo in backend/repositories/parent_repo.go — all methods per contracts. GetOrCreate uses GORM's `FirstOrCreate` or manual check+insert. Return `(parent, isNew, error)`.
- [x] T00 [P] [US2] Implement ChildRepo in backend/repositories/child_repo.go — all methods per contracts. Use `db.Transaction()` for DeleteAll. Handle duplicate `(family_id, first_name)` constraint with user-friendly error. GetBalance returns int64 cents.
- [x] T00 [P] [US2] Implement TransactionRepo in backend/repositories/transaction_repo.go — all methods per contracts. Deposit/Withdraw use `db.Transaction()` for atomicity (update child balance + insert transaction). Use `db.Raw()` for GetMonthlySummary (to_char aggregation). Define `MonthlySummary` helper struct in this file.
- [x] T00 [P] [US2] Implement ScheduleRepo in backend/repositories/schedule_repo.go — all methods per contracts. ListByFamily requires JOIN with children table. ListDue filters by `next_run_at <= now AND status = 'active'`.
- [x] T00 [P] [US2] Implement InterestScheduleRepo in backend/repositories/interest_schedule_repo.go — all methods per contracts. Similar patterns to ScheduleRepo. ListByFamily requires JOIN with children table.
- [x] T00 [P] [US2] Implement InterestRepo in backend/repositories/interest_repo.go — all methods per contracts. AccrueInterest uses `db.Transaction()` to calculate interest from basis points, update child balance, insert interest transaction, and update last_interest_at atomically.
- [x] T00 [P] [US2] Implement RefreshTokenRepo in backend/repositories/refresh_token_repo.go — all methods per contracts. DeleteExpired deletes tokens where `expires_at < NOW()` and returns count of deleted rows.
- [x] T00 [P] [US2] Implement AuthEventRepo in backend/repositories/auth_event_repo.go — all methods per contracts. Log inserts a new auth event. ListByFamily orders by created_at DESC with limit.
- [x] T00 [P] [US2] Implement WebhookEventRepo in backend/repositories/webhook_event_repo.go — all methods per contracts. Exists checks by primary key (stripe_event_id). Create inserts with idempotency.
- [x] T00 [P] [US2] Implement SavingsGoalRepo in backend/repositories/savings_goal_repo.go — all methods per contracts. Allocate/Deallocate use `db.Transaction()` to update goal's saved_cents and insert goal_allocation atomically.
- [x] T00 [P] [US2] Implement GoalAllocationRepo in backend/repositories/goal_allocation_repo.go — all methods per contracts. ListByGoal and ListByChild order by created_at DESC.
- [x] T00 [US2] Run full repository test suite: `cd backend && go test -p 1 ./repositories/...` — all tests must pass

**Checkpoint**: All 12 repositories implemented and tested. `go test -p 1 ./repositories/...` passes. Each repo method produces identical behavior to its store counterpart.

---

## Phase 5: User Story 3 — Migrate Handlers to Use Repositories (Priority: P2)

**Goal**: All handler packages depend on repositories instead of store. No API behavior changes.

**Independent Test**: Full existing test suite passes with handlers wired to repositories. No `internal/store` imports in handler packages.

### Implementation for User Story 3

- [x] T00 [US3] Update backend/main.go — replace `store.Open()` with `repositories.Open()`, instantiate all repository structs instead of store structs, pass repositories to handler constructors. Keep the `*gorm.DB` instance for all repos.
- [x] T00 [P] [US3] Migrate backend/internal/family/ handlers — change constructor to accept `*repositories.FamilyRepo`, `*repositories.ParentRepo`, `*repositories.ChildRepo`, `*repositories.AuthEventRepo`. Update all method calls from store types to repository types. Update imports.
- [x] T00 [P] [US3] Migrate backend/internal/auth/ handlers — change constructor to accept repository types for ParentRepo, FamilyRepo, ChildRepo, RefreshTokenRepo, AuthEventRepo. Update all method calls and imports.
- [x] T00 [P] [US3] Migrate backend/internal/balance/ handlers — change constructor to accept TransactionRepo, ChildRepo, InterestRepo, InterestScheduleRepo, SavingsGoalRepo. Update all method calls and imports.
- [x] T00 [P] [US3] Migrate backend/internal/allowance/ handlers — change constructor to accept ScheduleRepo, ChildRepo, FamilyRepo. Update all method calls and imports.
- [x] T00 [P] [US3] Migrate backend/internal/interest/ handlers and scheduler — change constructor to accept InterestScheduleRepo, InterestRepo, ChildRepo. Update scheduler goroutine to use repository methods.
- [x] T00 [P] [US3] Migrate backend/internal/goals/ handlers — change constructor to accept SavingsGoalRepo, GoalAllocationRepo, ChildRepo. Update all method calls and imports.
- [x] T00 [P] [US3] Migrate backend/internal/settings/ handlers — change constructor to accept FamilyRepo. Update all method calls and imports.
- [x] T00 [P] [US3] Migrate backend/internal/subscription/ handlers — change constructor to accept FamilyRepo, WebhookEventRepo, ChildRepo. Update all method calls and imports.
- [x] T050 [P] [US3] Migrate backend/internal/contact/ handlers — update constructor to accept any repository dependencies. Update imports.
- [x] T051 [P] [US3] Migrate backend/internal/allowance/ processor/scheduler — update the allowance processor goroutine to use ScheduleRepo and TransactionRepo instead of store types.
- [x] T052 [US3] Update backend/internal/testutil/ — update shared test helpers to work with `*gorm.DB` if any test utilities reference the DB connection type.
- [x] T053 [US3] Run full test suite: `cd backend && go test -p 1 ./...` — all existing tests must pass with repository-backed handlers

**Checkpoint**: All handlers use repositories. No handler package imports `internal/store`. Full test suite passes.

---

## Phase 6: User Story 4 — Retire the Store Package (Priority: P3)

**Goal**: Remove `backend/internal/store/` entirely. Single data access pattern via repositories.

**Independent Test**: Project compiles and all tests pass with store package deleted. No imports of `internal/store` anywhere.

### Implementation for User Story 4

- [x] T054 [US4] Verify no remaining imports of `internal/store` in any non-store package — run `grep -r "internal/store" backend/ --include="*.go"` excluding the store package itself. Fix any remaining references.
- [x] T055 [US4] Remove backend/internal/store/ directory entirely — delete all store files (postgres.go, family.go, child.go, parent.go, transaction.go, schedule.go, interest_schedule.go, interest.go, refresh_token.go, auth_event.go, webhook_event.go, savings_goal.go, and all corresponding test files).
- [x] T056 [US4] Run full test suite after store removal: `cd backend && go test -p 1 ./...` — confirm everything compiles and passes without the store package.

**Checkpoint**: Store package removed. Project compiles. All tests pass. `backend/models/` and `backend/repositories/` are the sole data access layer.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T057 Run `cd backend && go vet ./...` to verify no issues across the codebase
- [x] T058 Run `cd backend && go build ./...` to verify clean compilation of all packages
- [x] T059 Verify no GORM AutoMigrate usage anywhere: search for "AutoMigrate" in backend/ — must return zero results
- [x] T060 Verify int64 cents preserved: spot-check models and repository methods for money fields — all must be int64, no float64 or decimal types

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (GORM deps installed)
- **US1 Models (Phase 3)**: Depends on Phase 1 (package structure exists)
- **US2 Repositories (Phase 4)**: Depends on Phase 2 (db.go, test helpers) AND Phase 3 (models)
- **US3 Handler Migration (Phase 5)**: Depends on Phase 4 (all repositories implemented and tested)
- **US4 Store Removal (Phase 6)**: Depends on Phase 5 (all handlers migrated)
- **Polish (Phase 7)**: Depends on Phase 6

### User Story Dependencies

- **US1 (P1 — Models)**: After Foundational → all 11 model tasks run in parallel
- **US2 (P1 — Repositories)**: After US1 complete → all 12 test tasks in parallel, then all 12 impl tasks in parallel
- **US3 (P2 — Handler Migration)**: After US2 complete → T041 (main.go) first, then T042-T052 in parallel
- **US4 (P3 — Store Removal)**: After US3 complete → sequential (verify → delete → test)

### Within Each User Story

- US1: All model files are independent → full parallel
- US2: Tests written first (TDD), then implementation. Tests for different repos are parallel. Implementations for different repos are parallel.
- US3: main.go wiring first (T041), then individual handler packages in parallel
- US4: Sequential: verify → remove → test

### Parallel Opportunities

- **Phase 3**: All 11 model tasks (T005–T015) run in parallel
- **Phase 4 tests**: All 12 test tasks (T016–T027) run in parallel
- **Phase 4 impl**: All 12 implementation tasks (T028–T039) run in parallel
- **Phase 5**: After T041, tasks T042–T051 run in parallel (different handler packages)

---

## Parallel Example: User Story 1 (Models)

```text
# Launch all 11 model files in parallel:
Task: T005 "Create Family model in backend/models/family.go"
Task: T006 "Create Parent model in backend/models/parent.go"
Task: T007 "Create Child model in backend/models/child.go"
Task: T008 "Create Transaction model in backend/models/transaction.go"
Task: T009 "Create AllowanceSchedule model in backend/models/allowance_schedule.go"
Task: T010 "Create InterestSchedule model in backend/models/interest_schedule.go"
Task: T011 "Create RefreshToken model in backend/models/refresh_token.go"
Task: T012 "Create AuthEvent model in backend/models/auth_event.go"
Task: T013 "Create StripeWebhookEvent model in backend/models/webhook_event.go"
Task: T014 "Create SavingsGoal model in backend/models/savings_goal.go"
Task: T015 "Create GoalAllocation model in backend/models/goal_allocation.go"
```

## Parallel Example: User Story 2 (Repository Tests)

```text
# Launch all 12 test files in parallel (TDD - write tests first):
Task: T016 "Write tests for FamilyRepo in backend/repositories/family_repo_test.go"
Task: T017 "Write tests for ParentRepo in backend/repositories/parent_repo_test.go"
Task: T018 "Write tests for ChildRepo in backend/repositories/child_repo_test.go"
...then all implementation tasks T028-T039 in parallel
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (deps + dirs)
2. Complete Phase 2: Foundational (db.go + test helpers)
3. Complete Phase 3: US1 — All models
4. Complete Phase 4: US2 — All repositories with tests
5. **STOP and VALIDATE**: All repository tests pass, behavioral equivalence confirmed
6. This is a meaningful MVP — the new data layer exists and is proven correct

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. US1 (Models) → Models compile, schema match confirmed
3. US2 (Repositories) → Full repository layer tested independently
4. US3 (Handler Migration) → Application fully wired to new layer, all tests pass
5. US4 (Store Removal) → Clean codebase, single data access pattern
6. Polish → Final validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- TDD enforced per constitution: write repo tests (T016–T027) before implementations (T028–T039)
- All tests run with `go test -p 1` due to shared test database
- Repository method names mirror store method names to minimize handler changes
- Use `db.Raw()` for complex SQL (monthly summaries, slug suggestions, interest calculations)
- Return `(nil, nil)` for not-found — match existing store convention
- Money fields are always int64 (cents) — never float64
