# Feature Specification: SQLite to PostgreSQL Migration

**Feature Branch**: `011-sqlite-to-postgres`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "SQLite to PostgreSQL database migration — hard cutover with no SQLite support remaining"
**Further context**: In file ./sqlite-to-postgres-migration-abstract.md

## User Scenarios & Testing *(mandatory)*

### User Story 1 - All Application Features Continue Working After Migration (Priority: P1)

As a parent or child user, I can continue using all existing Bank of Dad features — logging in, viewing balances, making deposits/withdrawals, managing allowances, and accruing interest — without any disruption or change in behavior after the database is migrated.

**Why this priority**: The migration must be invisible to end users. If any existing feature breaks, the migration has failed regardless of any other benefits.

**Independent Test**: Can be fully tested by exercising every existing user workflow (login, view balance, deposit, withdraw, set allowance, view transactions, set interest rate) and verifying identical behavior and data integrity.

**Acceptance Scenarios**:

1. **Given** the system has been migrated to the new database, **When** a parent logs in and views their children's accounts, **Then** all account data (names, balances, transaction history) is displayed correctly.
2. **Given** the system has been migrated, **When** a parent creates a deposit or withdrawal, **Then** the transaction is recorded and the child's balance updates correctly.
3. **Given** the system has been migrated, **When** a child logs in, **Then** they see their correct balance and transaction history.
4. **Given** the system has been migrated, **When** an allowance schedule fires, **Then** the correct amount is deposited to the child's account on the expected schedule.
5. **Given** the system has been migrated, **When** interest accrual runs, **Then** the correct interest amount is calculated and credited based on the configured rate and frequency.

---

### User Story 2 - Reliable Automated Testing Against Production Database (Priority: P2)

As a developer, I can run the full test suite against the same type of database used in production, so that tests catch real database-related bugs rather than passing due to differences between the test and production database engines.

**Why this priority**: Without test parity, regressions can ship undetected. Tests that pass against one database engine but fail against another provide false confidence.

**Independent Test**: Can be verified by running the test suite against the production database engine and confirming all tests pass with proper isolation (no test-to-test state leakage).

**Acceptance Scenarios**:

1. **Given** a developer has the application's database dependency running locally, **When** they run the test suite, **Then** all tests pass against the production database engine.
2. **Given** multiple tests run in sequence, **When** one test creates or modifies data, **Then** subsequent tests are not affected (test isolation is maintained).
3. **Given** a shared test helper exists, **When** a new test file is created, **Then** the developer can set up a test database with a single function call rather than duplicating setup code.

---

### User Story 3 - Seamless Local Development and CI (Priority: P3)

As a developer, I can start the full application stack locally with a single command and have CI automatically validate all changes against the production database engine, so that development and deployment are reliable and reproducible.

**Why this priority**: Developer productivity and deployment confidence depend on the local environment and CI pipeline matching production. This is foundational but lower priority than correctness.

**Independent Test**: Can be verified by running the containerized stack from scratch and confirming the backend starts, connects to the database, and serves requests; and by confirming the CI pipeline runs tests against the production database engine.

**Acceptance Scenarios**:

1. **Given** a developer clones the repository fresh, **When** they start the containerized stack, **Then** the database starts, the backend connects, migrations run, and the application is ready to serve requests.
2. **Given** a developer pushes a branch, **When** CI runs, **Then** tests execute against the production database engine with proper service dependencies.
3. **Given** the containerized stack is running, **When** the developer stops and restarts it, **Then** previously stored data persists across restarts.

---

### Edge Cases

- What happens if the database is unavailable when the application starts? The application should fail fast with a clear error message rather than silently degrading.
- What happens if a database migration has already been applied? Re-running migrations should be idempotent — already-applied migrations are skipped without error.
- What happens if a test fails mid-execution? The database should not be left in a dirty state that causes subsequent tests to fail.
- How does the system handle concurrent writes? The new database engine natively supports concurrent read/write access, so the previous single-writer workaround should be removed.
- What happens if the `DATABASE_URL` environment variable is missing? The application should fail with a clear error message indicating the missing configuration.
- How are financial values stored? All monetary amounts must be stored with exact precision (no floating-point rounding errors), consistent with the application's existing cents-based model.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST connect to a PostgreSQL database using a connection string read from the `DATABASE_URL` environment variable.
- **FR-002**: System MUST run database schema migrations automatically on startup, and skip already-applied migrations without error.
- **FR-003**: System MUST use a single database connection pool (removing the previous read/write split workaround that was specific to SQLite's concurrency model).
- **FR-004**: System MUST store all monetary values with exact decimal precision — no floating-point types for financial data.
- **FR-005**: System MUST preserve all existing data constraints: foreign keys, uniqueness, check constraints, and NOT NULL rules.
- **FR-006**: System MUST preserve all existing application behavior: authentication, session management, transactions, allowance scheduling, interest accrual, and child account management.
- **FR-007**: System MUST provide a shared test helper that sets up an isolated test database, runs migrations, and cleans up after each test.
- **FR-008**: System MUST include the database as a service in the containerized development stack, with health checks and persistent storage.
- **FR-009**: System MUST include the database as a service in the CI pipeline so that automated tests run against the same database engine used in production.
- **FR-010**: System MUST NOT retain any references to or dependencies on the previous SQLite database engine. This is a complete cutover.
- **FR-011**: System MUST fail with a clear error message if the database connection cannot be established at startup.
- **FR-012**: System MUST provide a `.env.example` file with a sensible default `DATABASE_URL` for local development.

### Key Entities

All existing key entities are preserved without change. The migration affects storage and connectivity, not the domain model:

- **Family**: Organizational unit grouping parents and children.
- **Parent**: Adult user who manages child accounts. Authenticated via external identity provider.
- **Child**: Minor user with a balance, transaction history, and optional allowance/interest schedules. Authenticated via password.
- **Transaction**: A financial event (deposit, withdrawal, allowance, interest) affecting a child's balance.
- **Allowance Schedule**: A recurring rule for automatic deposits to a child's account.
- **Interest Schedule**: A recurring rule for automatic interest accrual on a child's balance.
- **Session**: An authentication token with expiration, scoped to a user and family.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of existing application features work identically after migration — no user-visible behavior changes.
- **SC-002**: 100% of existing tests pass against the new database engine without modification to test assertions (only infrastructure/setup changes).
- **SC-003**: The full test suite completes in under 60 seconds on a standard development machine.
- **SC-004**: A developer can start the full stack from a fresh clone with a single container orchestration command in under 30 seconds.
- **SC-005**: No references to the previous database engine remain anywhere in the codebase.
- **SC-006**: All financial calculations produce identical results with exact decimal precision — no rounding differences introduced by the migration.
- **SC-007**: The CI pipeline runs the full test suite against the production database engine on every push.

## Assumptions

- The application does not need to support running against both SQLite and PostgreSQL simultaneously. This is a hard cutover.
- No existing production data needs to be migrated (data migration/ETL is out of scope). The new schema is created fresh.
- The `DATABASE_URL` format follows the standard PostgreSQL connection string convention.
- Local development will use a containerized database instance managed by the existing container orchestration setup.
- The existing cents-based integer model for money (`int64` representing cents) will be mapped to an appropriate exact-precision database type.
- Test isolation will be achieved by truncating tables between tests rather than using transaction rollback, since some tests may need to test transaction behavior explicitly.

## Out of Scope

- Migrating existing production data from SQLite to PostgreSQL (no ETL process).
- Adding new features, tables, or application behavior as part of this migration.
- Supporting multiple database backends or a database abstraction layer.
- Performance benchmarking or query optimization beyond ensuring existing behavior is preserved.
- Changes to the frontend — this migration is entirely backend/infrastructure.
