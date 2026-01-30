# Tasks: User Authentication

**Input**: Design documents from `/specs/001-user-auth/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/auth-api.md, quickstart.md

**Tests**: REQUIRED by constitution (Test-First Development principle). Tests are written first and must FAIL before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `backend/` (Go 1.21)
- **Frontend**: `frontend/` (React + TypeScript + Vite)
- **Tests**: `backend/tests/contract/`, `backend/tests/integration/`, `backend/internal/*_test.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, dependency installation, and directory structure

- [x] T001 Add Go dependencies (`modernc.org/sqlite`, `golang.org/x/oauth2`, `golang.org/x/crypto`, `github.com/stretchr/testify`) to `backend/go.mod` and run `go mod tidy`
- [x] T002 Create backend directory structure: `backend/internal/auth/`, `backend/internal/family/`, `backend/internal/store/`, `backend/internal/middleware/`, `backend/internal/config/`, `backend/tests/contract/`, `backend/tests/integration/`
- [x] T003 [P] Add `react-router-dom` dependency to `frontend/package.json` and run `npm install`
- [x] T004 [P] Create frontend directory structure: `frontend/src/pages/`, `frontend/src/components/`
- [x] T005 [P] Create shared TypeScript types in `frontend/src/types.ts` (User, Child, Family, AuthState, ApiError)
- [x] T006 [P] Create typed fetch wrapper in `frontend/src/api.ts` with `credentials: 'include'`, JSON handling, and error typing per contracts/auth-api.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T007 Implement environment configuration loader in `backend/internal/config/config.go` (Google OAuth client ID/secret, redirect URL, database path, cookie domain, server port) reading from environment variables
- [x] T008 Implement SQLite connection manager in `backend/internal/store/sqlite.go` with WAL mode, read/write pools, PRAGMAs (`journal_mode=WAL`, `synchronous=NORMAL`, `busy_timeout=5000`, `foreign_keys=ON`), and schema migration (create all 5 tables from data-model.md)
- [x] T009 [P] Write unit tests for SQLite connection and migration in `backend/internal/store/sqlite_test.go` (verify tables created, PRAGMAs set, WAL mode active)
- [x] T010 Implement session store operations (Create, GetByToken, DeleteByToken, DeleteExpired) in `backend/internal/store/session.go` using 32-byte `crypto/rand` tokens, base64-encoded
- [x] T011 [P] Write unit tests for session store in `backend/internal/store/session_test.go` (create session, get valid session, reject expired session, delete session, cleanup expired)
- [x] T012 Implement auth middleware in `backend/internal/middleware/auth.go` that reads `session` cookie, validates token against sessions table, checks expiry, and attaches `user_type`, `user_id`, `family_id` to request context. Return 401 for invalid/expired sessions
- [x] T013 [P] Implement CORS middleware in `backend/internal/middleware/cors.go` replacing the inline CORS in `backend/main.go` (allow credentials, configurable origins)
- [x] T014 [P] Implement auth event logging middleware in `backend/internal/middleware/logging.go` that logs auth events to `auth_events` table (event_type, user_type, user_id, family_id, ip_address, details) without sensitive data per FR-018
- [x] T015 [P] Implement auth event store operations (LogEvent, GetEventsByFamily) in `backend/internal/store/auth_event.go`
- [x] T016 [P] Write unit tests for auth event store in `backend/internal/store/auth_event_test.go` (log event, query by family, verify no sensitive data fields)
- [x] T017 Implement health check endpoint `GET /api/health` and refactor `backend/main.go` to use new router structure with middleware chain (CORS → logging → routes, auth middleware applied per-route)
- [x] T018 [P] Update `frontend/src/App.tsx` to set up react-router-dom `BrowserRouter` with route definitions for all pages (/, /dashboard, /setup, /auth/callback, /:familySlug, /child/dashboard, *)
- [x] T019 [P] Implement `ProtectedRoute` component in `frontend/src/components/ProtectedRoute.tsx` that calls `GET /api/auth/me` to verify session, redirects to home if unauthenticated, and passes user context to children
- [x] T020 [P] Implement shared `Layout` component in `frontend/src/components/Layout.tsx` with navigation bar, user display name, and logout button (calls `POST /api/auth/logout`)
- [x] T021 [P] Implement `NotFound` page in `frontend/src/pages/NotFound.tsx` showing "This bank doesn't exist" message with link to create your own (edge case: non-existent family URL)
- [x] T022 Update Vite config in `frontend/vite.config.ts` to proxy `/api` requests to backend (`http://localhost:8001`) for development
- [x] T023 Update `docker-compose.yaml` and `docker-compose.override.yaml` to pass Google OAuth environment variables and mount SQLite data volume

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — Parent Registration via Google (Priority: P1) MVP

**Goal**: A parent can register via Google OAuth, choose a family bank URL slug, and access their dashboard

**Independent Test**: Complete Google sign-in flow, choose slug, verify redirect to dashboard with family URL created

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T024 [P] [US1] Write contract tests for `GET /api/auth/google/login`, `GET /api/auth/google/callback`, `POST /api/families`, `GET /api/families/check-slug` in `backend/tests/contract/auth_test.go` — test redirect response, state cookie, callback with valid/invalid state, family creation with valid/invalid/duplicate slugs, slug availability check with suggestions
- [ ] T025 [P] [US1] Write integration test for parent registration journey in `backend/tests/integration/parent_registration_test.go` — simulate full OAuth flow (mock Google), choose slug, verify parent + family created in DB, verify session cookie set, verify duplicate Google ID returns existing account

### Implementation for User Story 1

- [x] T026 [P] [US1] Implement family store operations (Create, GetBySlug, SlugExists, SuggestSlugs) in `backend/internal/store/family.go` with slug validation (3-30 chars, lowercase alphanumeric + hyphens, regex `^[a-z0-9][a-z0-9-]*[a-z0-9]$`) per FR-003, FR-004
- [x] T027 [P] [US1] Write unit tests for family store in `backend/internal/store/family_test.go` (create family, duplicate slug rejection, slug validation, slug suggestions)
- [x] T028 [P] [US1] Implement parent store operations (Create, GetByGoogleID, GetByID) in `backend/internal/store/parent.go`
- [x] T029 [P] [US1] Write unit tests for parent store in `backend/internal/store/parent_test.go` (create parent, find by Google ID, duplicate Google ID rejection)
- [x] T030 [US1] Implement slug validation helper in `backend/internal/family/validation.go` (ValidateSlug function checking format and length per FR-003)
- [x] T031 [US1] Implement Google OAuth handlers in `backend/internal/auth/google.go`: `HandleGoogleLogin` (generate state, set cookie, redirect to Google), `HandleGoogleCallback` (validate state, exchange code, fetch userinfo, create-or-find parent, create session, redirect to /dashboard or /setup)
- [x] T032 [US1] Implement family handlers in `backend/internal/family/handlers.go`: `HandleCreateFamily` (POST /api/families — validate slug, create family, link to parent), `HandleCheckSlug` (GET /api/families/check-slug — return availability + suggestions)
- [x] T033 [US1] Register US1 routes in `backend/main.go`: `GET /api/auth/google/login`, `GET /api/auth/google/callback`, `POST /api/families` (auth required), `GET /api/families/check-slug` (auth required)
- [x] T034 [P] [US1] Implement `HomePage` in `frontend/src/pages/HomePage.tsx` with "Sign in with Google" button that navigates to `/api/auth/google/login`
- [x] T035 [P] [US1] Implement `GoogleCallback` page in `frontend/src/pages/GoogleCallback.tsx` that handles OAuth redirect (reads auth state, redirects to dashboard or setup)
- [x] T036 [US1] Implement `SlugPicker` component in `frontend/src/components/SlugPicker.tsx` with slug input field, real-time availability checking via `GET /api/families/check-slug`, validation feedback, and slug suggestions display
- [x] T037 [US1] Implement setup page (slug selection) in `frontend/src/pages/SetupPage.tsx` using `SlugPicker` component, calls `POST /api/families` on submit, redirects to dashboard on success
- [x] T038 [US1] Implement basic `ParentDashboard` in `frontend/src/pages/ParentDashboard.tsx` showing parent display name, family slug/URL, and placeholder for child list (completed in US3)

**Checkpoint**: Parent registration via Google fully functional. Parent can sign in, choose slug, see dashboard.

---

## Phase 4: User Story 2 — Parent Login via Google (Priority: P1)

**Goal**: A registered parent can return, log in via Google, access dashboard, and log out

**Independent Test**: Log in with previously registered Google account, verify dashboard access, log out, verify redirect to home

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T039 [P] [US2] Write contract tests for `POST /api/auth/logout` and `GET /api/auth/me` in `backend/tests/contract/auth_test.go` (append to existing file) — test logout clears session, /me returns parent info, /me returns 401 without session
- [ ] T040 [P] [US2] Write integration test for parent login journey in `backend/tests/integration/parent_login_test.go` — registered parent logs in (mock Google), gets session, accesses /me, logs out, verify session invalidated, verify unregistered Google account redirects to setup

### Implementation for User Story 2

- [x] T041 [US2] Implement `HandleGetMe` in `backend/internal/auth/handlers.go` (GET /api/auth/me — read session context, return parent or child identity with family slug)
- [x] T042 [US2] Implement `HandleLogout` in `backend/internal/auth/handlers.go` (POST /api/auth/logout — delete session from DB, clear cookie, log auth event)
- [x] T043 [US2] Register US2 routes in `backend/main.go`: `POST /api/auth/logout` (auth required), `GET /api/auth/me` (auth required)
- [x] T044 [US2] Update `Layout` component in `frontend/src/components/Layout.tsx` to call `POST /api/auth/logout` on logout button click and redirect to home page
- [x] T045 [US2] Update `ProtectedRoute` in `frontend/src/components/ProtectedRoute.tsx` to use `GET /api/auth/me` response to distinguish parent vs child and redirect unregistered parents to setup

**Checkpoint**: Parent login/logout fully functional. Combined with US1, complete parent access cycle works.

---

## Phase 5: User Story 3 — Parent Creates Child Account (Priority: P1)

**Goal**: A logged-in parent can add a child with a first name and password, seeing the child's login credentials

**Independent Test**: Parent creates child, sees child in list with login URL and credentials displayed

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T046 [P] [US3] Write contract tests for `POST /api/children` and `GET /api/children` in `backend/tests/contract/child_test.go` — test create child with valid data (201), password too short (400), duplicate name in family (409), list children returns correct data, auth required (401)
- [ ] T047 [P] [US3] Write integration test for child account creation journey in `backend/tests/integration/child_account_test.go` — parent creates child, verify child in DB with hashed password (not plaintext per SC-005), verify unique name enforcement within family, verify child appears in GET /api/children list

### Implementation for User Story 3

- [x] T048 [P] [US3] Implement child store operations (Create, GetByFamilyAndName, ListByFamily, GetByID) in `backend/internal/store/child.go` with bcrypt password hashing (cost 12) per FR-008, unique constraint on (family_id, first_name) per FR-007
- [x] T049 [P] [US3] Write unit tests for child store in `backend/internal/store/child_test.go` (create child with hashed password, verify bcrypt hash not plaintext, find by family+name, duplicate name rejection, list by family)
- [x] T050 [US3] Implement child management handlers in `backend/internal/family/handlers.go` (append): `HandleCreateChild` (POST /api/children — validate password min 6 chars per FR-006, validate unique name per FR-007, hash password, create child, return login URL), `HandleListChildren` (GET /api/children — list all children in parent's family)
- [x] T051 [US3] Register US3 routes in `backend/main.go`: `POST /api/children` (parent auth required), `GET /api/children` (parent auth required)
- [x] T052 [P] [US3] Implement `AddChildForm` component in `frontend/src/components/AddChildForm.tsx` with first name input, password input, validation feedback (min 6 chars), submit calls `POST /api/children`, displays login URL and credentials on success
- [x] T053 [P] [US3] Implement `ChildList` component in `frontend/src/components/ChildList.tsx` that calls `GET /api/children` and displays list of children with name, locked status, and created date
- [x] T054 [US3] Update `ParentDashboard` in `frontend/src/pages/ParentDashboard.tsx` to include `AddChildForm` and `ChildList` components

**Checkpoint**: Child account creation fully functional. Parent can create children and see them listed.

---

## Phase 6: User Story 4 — Child Login via Family Bank URL (Priority: P1)

**Goal**: A child navigates to their family URL, logs in with name and password, sees their dashboard

**Independent Test**: Navigate to family URL, log in as child, see child dashboard, log out, test wrong credentials and account lockout

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T055 [P] [US4] Write contract tests for `POST /api/auth/child/login` and `GET /api/families/:slug` in `backend/tests/contract/auth_test.go` (append) — test child login success (200 + session cookie), invalid credentials (401 with friendly message), account locked (403), family not found (404), family exists check (200)
- [ ] T056 [P] [US4] Write integration test for child login journey in `backend/tests/integration/child_login_test.go` — create family+parent+child, child logs in via family slug, gets session, accesses /me as child, logs out, test 5 failed attempts triggers lockout (FR-010), verify parent notified of lockout (FR-011), verify auth events logged per FR-018

### Implementation for User Story 4

- [x] T057 [US4] Implement child authentication operations in `backend/internal/store/child.go` (append): `IncrementFailedAttempts`, `LockAccount`, `ResetFailedAttempts`, `IsLocked` — account lockout after 5 failures per FR-010
- [x] T058 [US4] Write unit tests for child lockout logic in `backend/internal/store/child_test.go` (append) — test increment attempts, lock at 5, reset on success, is_locked flag
- [x] T059 [US4] Implement child login handler in `backend/internal/auth/child.go`: `HandleChildLogin` (POST /api/auth/child/login — find family by slug, find child by family+name, check locked status, compare bcrypt password, create 24-hour session, handle lockout at 5 failures, log auth events)
- [x] T060 [US4] Implement family lookup handler in `backend/internal/family/handlers.go` (append): `HandleGetFamily` (GET /api/families/:slug — return exists true/false, public endpoint)
- [x] T061 [US4] Register US4 routes in `backend/main.go`: `POST /api/auth/child/login` (no auth), `GET /api/families/:slug` (no auth)
- [x] T062 [US4] Implement `FamilyLogin` page in `frontend/src/pages/FamilyLogin.tsx` — reads `:familySlug` from URL params, calls `GET /api/families/:slug` to verify family exists (show 404 if not), displays first name + password inputs, submits to `POST /api/auth/child/login`, shows friendly error messages for invalid credentials and locked accounts, redirects to child dashboard on success
- [x] T063 [US4] Implement `ChildDashboard` page in `frontend/src/pages/ChildDashboard.tsx` — displays child's first name, family name, and placeholder account balance, with logout button

**Checkpoint**: Child login fully functional. Complete parent-to-child flow works (register, create child, child logs in).

---

## Phase 7: User Story 5 — Parent Manages Child Credentials (Priority: P2)

**Goal**: A parent can reset a child's password and update their display name

**Independent Test**: Parent resets password, child logs in with new password (old fails), parent updates name, name reflected in app

### Tests for User Story 5

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T064 [P] [US5] Write contract tests for `PUT /api/children/:id/password` and `PUT /api/children/:id/name` in `backend/tests/contract/child_test.go` (append) — test password reset success (200 + unlocks account), password too short (400), forbidden (403 - wrong family), child not found (404), name update success (200), duplicate name (409)
- [ ] T065 [P] [US5] Write integration test for credential management journey in `backend/tests/integration/child_account_test.go` (append) — parent resets locked child's password, verify account unlocked + failed attempts reset, child logs in with new password, old password fails, parent updates name, child logs in with new name

### Implementation for User Story 5

- [x] T066 [US5] Implement child update operations in `backend/internal/store/child.go` (append): `UpdatePassword` (hash new password, reset is_locked + failed_attempts), `UpdateName` (validate uniqueness within family, update first_name + updated_at)
- [x] T067 [US5] Write unit tests for child update operations in `backend/internal/store/child_test.go` (append) — test password update rehashes, unlock on reset, name update, duplicate name rejection
- [x] T068 [US5] Implement credential management handlers in `backend/internal/family/handlers.go` (append): `HandleResetPassword` (PUT /api/children/:id/password — validate parent owns family, validate password min 6, hash + update, log event), `HandleUpdateName` (PUT /api/children/:id/name — validate parent owns family, validate unique name, update, log event)
- [x] T069 [US5] Register US5 routes in `backend/main.go`: `PUT /api/children/:id/password` (parent auth required), `PUT /api/children/:id/name` (parent auth required)
- [x] T070 [US5] Implement `ManageChild` component in `frontend/src/components/ManageChild.tsx` — displays child info, "Reset Password" button with new password input, "Update Name" input, calls respective API endpoints, shows success/error feedback
- [x] T071 [US5] Update `ParentDashboard` in `frontend/src/pages/ParentDashboard.tsx` to show `ManageChild` component when a child is selected from `ChildList`

**Checkpoint**: Credential management fully functional. Parents can manage all child account details.

---

## Phase 8: User Story 6 — Session Persistence (Priority: P2)

**Goal**: Logged-in users remain authenticated across browser restarts within their session TTL

**Independent Test**: Log in, close browser, reopen, verify still authenticated. Wait for expiry, verify redirected to login.

### Tests for User Story 6

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T072 [P] [US6] Write contract tests for session persistence in `backend/tests/contract/auth_test.go` (append) — test /me with valid non-expired parent session (7 days), test /me with valid non-expired child session (24 hours), test /me with expired session returns 401, verify session cookie has correct Max-Age (7 days for parent, 24 hours for child)
- [ ] T073 [P] [US6] Write integration test for session expiry in `backend/tests/integration/parent_login_test.go` (append) — create session with short TTL, verify /me works before expiry, verify /me returns 401 after expiry, verify expired session cleaned up

### Implementation for User Story 6

- [x] T074 [US6] Implement session cookie Max-Age differentiation in `backend/internal/auth/session.go`: parent sessions set `Max-Age=604800` (7 days per FR-012), child sessions set `Max-Age=86400` (24 hours per FR-013), verify `HttpOnly`, `Secure` (configurable for dev), `SameSite=Lax` attributes
- [x] T075 [US6] Implement expired session cleanup goroutine in `backend/internal/store/session.go` (append): `StartCleanupLoop` that runs `DeleteExpired` periodically (every hour), started from `main.go`
- [x] T076 [US6] Update auth middleware in `backend/internal/middleware/auth.go` to lazily delete expired sessions on access (delete row + clear cookie + return 401) in addition to the periodic cleanup

**Checkpoint**: Session persistence fully functional. All user stories complete.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T077 [P] Add input sanitization for all user-provided strings (slug, child name) in `backend/internal/family/validation.go` — trim whitespace, reject HTML/script content per constitution Security-First principle
- [x] T078 [P] Add rate limiting for child login endpoint to supplement account lockout in `backend/internal/middleware/auth.go` (or new file `backend/internal/middleware/ratelimit.go`)
- [x] T079 Verify all auth events are logged by reviewing each handler calls `LogEvent` — cross-reference with FR-018 event types (login_success, login_failure, logout, account_created, account_locked, password_reset, name_updated) in `backend/internal/store/auth_event.go`
- [ ] T080 Run `quickstart.md` validation scenarios manually against running application, document any failures
- [x] T081 [P] Update `backend/Dockerfile` to ensure SQLite data volume mount works correctly in production build
- [x] T082 [P] Update `frontend/nginx.conf` to handle SPA routing for all new pages (/:familySlug, /dashboard, /setup, /auth/callback, /child/dashboard)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational (Phase 2) — no story dependencies
- **US2 (Phase 4)**: Depends on US1 (Phase 3) — uses same OAuth endpoints, needs parent+family to exist for login testing
- **US3 (Phase 5)**: Depends on US1 (Phase 3) — needs parent+family to exist for child creation
- **US4 (Phase 6)**: Depends on US3 (Phase 5) — needs child accounts to exist for login testing
- **US5 (Phase 7)**: Depends on US4 (Phase 6) — needs child login working to verify password reset
- **US6 (Phase 8)**: Depends on US2 (Phase 4) + US4 (Phase 6) — needs both parent and child sessions working
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

```
Phase 1 (Setup)
  └──> Phase 2 (Foundational)
         └──> Phase 3 (US1: Parent Registration) ── MVP
                ├──> Phase 4 (US2: Parent Login)
                │      └──> Phase 8 (US6: Session Persistence) ─┐
                └──> Phase 5 (US3: Create Child)                 │
                       └──> Phase 6 (US4: Child Login) ──────────┘
                              └──> Phase 7 (US5: Manage Credentials)
                                                                 │
                                          Phase 9 (Polish) <─────┘
```

### Within Each User Story

1. Tests MUST be written and FAIL before implementation (TDD)
2. Store operations before handlers
3. Backend before frontend
4. Handlers before route registration
5. Components before page integration

### Parallel Opportunities

- **Phase 1**: T003, T004, T005, T006 can all run in parallel after T001+T002
- **Phase 2**: T009, T011, T013, T014, T015, T016, T18, T019, T020, T021 can run in parallel
- **Phase 3**: T024+T025 (tests) in parallel; T026+T027+T028+T029 (stores) in parallel; T034+T035 (frontend) in parallel
- **Phase 5**: T046+T047 (tests) in parallel; T048+T049 (stores) in parallel; T052+T053 (frontend) in parallel
- **Phase 6**: T055+T056 (tests) in parallel
- **Phase 7**: T064+T065 (tests) in parallel
- **Phase 8**: T072+T073 (tests) in parallel
- **Phase 9**: T077, T078, T081, T082 all in parallel

---

## Parallel Example: User Story 1

```bash
# Launch tests first (in parallel):
Task: "Contract tests for auth + family endpoints in backend/tests/contract/auth_test.go"
Task: "Integration test for registration journey in backend/tests/integration/parent_registration_test.go"

# After tests written and failing, launch store implementations (in parallel):
Task: "Family store operations in backend/internal/store/family.go"
Task: "Family store unit tests in backend/internal/store/family_test.go"
Task: "Parent store operations in backend/internal/store/parent.go"
Task: "Parent store unit tests in backend/internal/store/parent_test.go"

# After stores complete, implement handlers (sequential):
Task: "Slug validation in backend/internal/family/validation.go"
Task: "Google OAuth handlers in backend/internal/auth/google.go"
Task: "Family handlers in backend/internal/family/handlers.go"
Task: "Register routes in backend/main.go"

# Frontend (can parallel with backend after API is defined):
Task: "HomePage in frontend/src/pages/HomePage.tsx"
Task: "GoogleCallback in frontend/src/pages/GoogleCallback.tsx"
Task: "SlugPicker component in frontend/src/components/SlugPicker.tsx"
Task: "SetupPage in frontend/src/pages/SetupPage.tsx"
Task: "ParentDashboard in frontend/src/pages/ParentDashboard.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (Parent Registration)
4. **STOP and VALIDATE**: Test registration flow end-to-end
5. Deploy/demo if ready — parents can register and see their dashboard

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. US1 (Parent Registration) → Test independently → **MVP!**
3. US2 (Parent Login) → Test independently → Parents can return
4. US3 (Create Child) → Test independently → Families can grow
5. US4 (Child Login) → Test independently → Core educational value delivered
6. US5 (Manage Credentials) → Test independently → Self-service management
7. US6 (Session Persistence) → Test independently → Polished UX
8. Polish → Security hardening, validation, documentation

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Constitution requires TDD: all tests written and failing BEFORE implementation
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- bcrypt cost factor 12 for all password hashing
- Session cookies: `HttpOnly`, `Secure`, `SameSite=Lax`
