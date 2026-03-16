# Feature Specification: GORM Backend Refactor

**Feature Branch**: `029-gorm-backend-refactor`
**Created**: 2026-03-16
**Status**: Draft
**Input**: User description: "As a developer, I want better organization and abstractions of the backend codebase. Use GORM to manage database entities. Put models in backend/models/ and repository methods in backend/repositories/. Use repository methods for database interactions."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Define GORM Models for All Entities (Priority: P1)

As a developer, I want all database entities represented as GORM model structs in a dedicated `models/` package so that entity definitions are centralized, self-documenting, and decoupled from data access logic.

**Why this priority**: Models are the foundation that repositories depend on. Without well-defined models, no repository can be built. This is the prerequisite for all other stories.

**Independent Test**: Can be verified by importing the `models` package and confirming all entity structs compile, have correct GORM tags, and match the existing database schema (column names, types, constraints).

**Acceptance Scenarios**:

1. **Given** the existing database schema with 12 tables, **When** a developer imports the `models` package, **Then** a corresponding GORM-tagged struct exists for each table with field types matching the PostgreSQL column types.
2. **Given** a model struct, **When** a developer inspects its GORM tags, **Then** primary keys, foreign keys, nullable fields, and default values are correctly annotated.
3. **Given** entities with relationships (e.g., family has many children, child has many transactions), **When** a developer inspects the model, **Then** GORM association tags correctly express those relationships.

---

### User Story 2 - Create Repository Layer for Core Entities (Priority: P1)

As a developer, I want repository structs in a dedicated `repositories/` package that encapsulate all database interactions using GORM so that data access logic is organized by entity and uses a consistent abstraction.

**Why this priority**: The repository layer is the primary deliverable of this refactor — it replaces the current store package's raw SQL with GORM-based methods, providing the improved organization and abstraction the developer wants.

**Independent Test**: Can be verified by writing integration tests that call each repository method against a test database and confirming correct data is persisted and retrieved.

**Acceptance Scenarios**:

1. **Given** the existing store methods for an entity (e.g., families), **When** a developer looks at the corresponding repository, **Then** equivalent methods exist using GORM instead of raw SQL.
2. **Given** a repository struct, **When** a developer calls a create/read/update/delete method, **Then** the operation is performed through GORM and returns the correct model struct(s).
3. **Given** a repository method that previously used a raw SQL query with joins or aggregations, **When** that method is reimplemented, **Then** it uses GORM's query builder or preloading and produces identical results.

---

### User Story 3 - Migrate Handlers to Use Repositories (Priority: P2)

As a developer, I want all handler packages (auth, balance, family, allowance, interest, goals, settings, subscription, contact) to depend on repositories instead of the store package so that the codebase consistently uses the new abstraction layer.

**Why this priority**: Without updating the consumers, the new repositories would be unused. This story completes the refactor by wiring the new layer into the application.

**Independent Test**: Can be verified by running the full existing test suite — all tests pass with handlers now using repository methods, confirming behavioral equivalence.

**Acceptance Scenarios**:

1. **Given** a handler that currently depends on a store struct, **When** the migration is complete, **Then** it depends on a repository struct instead and no direct store imports remain.
2. **Given** the full application, **When** all handlers have been migrated, **Then** the existing end-to-end behavior is preserved (no user-facing changes).
3. **Given** the migrated codebase, **When** a developer searches for direct store usage in handler packages, **Then** no results are found (excluding test utilities).

---

### User Story 4 - Retire the Store Package (Priority: P3)

As a developer, I want the old `internal/store/` package removed once all consumers use repositories so that there is a single, clear data access pattern in the codebase.

**Why this priority**: Cleanup step that removes confusion from having two data access patterns. Lower priority because the system works correctly even if both exist temporarily.

**Independent Test**: Can be verified by confirming no non-test code imports `internal/store/` and the full test suite passes.

**Acceptance Scenarios**:

1. **Given** all handlers and schedulers use repositories, **When** the store package is removed, **Then** the project compiles and all tests pass.
2. **Given** the final codebase, **When** a developer looks at the project structure, **Then** `backend/models/` and `backend/repositories/` are the clear locations for entity definitions and data access.

---

### Edge Cases

- What happens when GORM's auto-generated SQL differs from hand-written SQL in edge cases (e.g., complex aggregations in transaction summaries or interest calculations)? Raw SQL via GORM's `db.Raw()` should be used where GORM's query builder cannot express the query.
- How are database migrations handled? GORM's auto-migration feature must NOT be used — the existing `golang-migrate` migration workflow is preserved. GORM models must match the schema defined by migrations, not drive it.
- What happens to the database connection setup? GORM wraps the database connection, so initialization changes from `sql.Open` to `gorm.Open` with the PostgreSQL driver.
- How are database transactions handled? Repositories must support GORM's `db.Transaction()` for operations that require atomicity (e.g., creating a goal allocation and updating a goal's current amount).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define GORM model structs for all 12 existing database tables in a `backend/models/` package.
- **FR-002**: Model structs MUST include GORM struct tags for column mapping, primary keys, foreign keys, nullability, and defaults that match the existing PostgreSQL schema exactly.
- **FR-003**: Model structs MUST define GORM associations (HasMany, BelongsTo, HasOne) for entity relationships (e.g., Family HasMany Children, Child HasMany Transactions).
- **FR-004**: System MUST provide repository structs in `backend/repositories/` that encapsulate all data access operations currently in `backend/internal/store/`.
- **FR-005**: Each repository MUST accept a `*gorm.DB` instance and expose methods equivalent to the existing store methods.
- **FR-006**: Repository methods MUST use GORM's query builder for standard CRUD operations and GORM's `Raw()`/`Exec()` for complex queries that cannot be cleanly expressed with the builder.
- **FR-007**: Repository methods MUST preserve the exact same behavior and return values as the store methods they replace — this is a refactor, not a feature change.
- **FR-008**: All handler packages MUST be updated to depend on repository structs instead of store structs.
- **FR-009**: The money representation (int64 cents) MUST be preserved in all models and repository methods.
- **FR-010**: Database migrations MUST continue to use `golang-migrate` — GORM's AutoMigrate MUST NOT be used.
- **FR-011**: Repository methods that require atomicity MUST use GORM's transaction support.
- **FR-012**: The existing test suite MUST pass after the migration with equivalent or better coverage.

### Key Entities

- **Family**: Core group entity. Has many Parents, Children. Identified by slug. Includes Stripe customer/subscription fields.
- **Parent**: Google OAuth user. Belongs to Family. Has email, Google ID, name.
- **Child**: Account holder. Belongs to Family. Has balance (int64 cents), interest rate, theme, pin hash, disabled flag.
- **Transaction**: Financial record. Belongs to Child. Types: deposit, withdrawal, allowance, interest, goal_deposit, goal_withdrawal.
- **AllowanceSchedule**: Recurring payment config. Belongs to Child. Frequency: weekly, biweekly, monthly.
- **InterestSchedule**: Recurring interest config. Belongs to Child. Frequency: monthly, quarterly, annually.
- **RefreshToken**: Auth token. Belongs to Parent. Stores hashed token, expiry, revocation status.
- **AuthEvent**: Audit log. Belongs to Parent. Records login events.
- **StripeWebhookEvent**: Idempotency tracker for Stripe webhooks. Keyed by event ID.
- **SavingsGoal**: Target savings. Belongs to Child. Has name, target amount, current amount, emoji, target date.
- **GoalAllocation**: Audit trail for goal fund movements. Belongs to SavingsGoal and Child.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All existing tests pass after the refactor with no reduction in test coverage percentage.
- **SC-002**: No handler or scheduler package directly imports the old store package after migration is complete.
- **SC-003**: Every database interaction in the codebase goes through a repository method — no raw SQL outside the repository layer.
- **SC-004**: Developer onboarding to a new entity follows a clear two-step pattern: define model in `models/`, implement repository in `repositories/`.
- **SC-005**: The number of lines of boilerplate SQL code is reduced by at least 30% compared to the current store implementation, as measured by total lines in data access files.

## Assumptions

- The existing `golang-migrate` migration workflow is preserved — GORM is used only for runtime data access, not schema management.
- The GORM PostgreSQL driver (`gorm.io/driver/postgres`) is compatible with the existing PostgreSQL 17 instance and connection configuration.
- Existing test helpers and test database setup will be adapted to work with GORM's `*gorm.DB` instead of `*sql.DB`.
- The `internal/store/postgres.go` connection initialization will be replaced with GORM's `gorm.Open()` using the same DSN/connection string.
- Schedulers (allowance processor, interest accrual, session cleanup) will also be updated to use repositories.
- No API contract changes — all HTTP request/response shapes remain identical.
