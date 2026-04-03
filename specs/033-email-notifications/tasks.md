# Tasks: Email Notifications

**Input**: Design documents from `/specs/033-email-notifications/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md, quickstart.md

**Tests**: Included per constitution (Test-First Development principle). Tests MUST be written and verified to FAIL before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Database migration and model updates shared by all stories

- [x] T001 [P] Create migration file `backend/migrations/014_notification_preferences.up.sql` — ALTER TABLE parents ADD COLUMN notify_withdrawal_requests BOOLEAN NOT NULL DEFAULT TRUE, notify_chore_completions BOOLEAN NOT NULL DEFAULT TRUE, notify_decisions BOOLEAN NOT NULL DEFAULT TRUE
- [x] T002 [P] Create migration file `backend/migrations/014_notification_preferences.down.sql` — ALTER TABLE parents DROP COLUMN notify_withdrawal_requests, notify_chore_completions, notify_decisions
- [x] T003 Add three GORM fields to Parent model in `backend/models/parent.go`: NotifyWithdrawalRequests (bool, default true), NotifyChoreCompletions (bool, default true), NotifyDecisions (bool, default true)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core notification infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational

- [x] T004 [P] Write test for `GetByFamilyID` method in `backend/repositories/parent_repo_test.go` — create family with 2 parents, verify both returned; verify empty slice for family with no parents
- [x] T005 [P] Write test for `UpdateNotificationPrefs` method in `backend/repositories/parent_repo_test.go` — update individual prefs, verify partial update only changes specified fields
- [x] T006 [P] Write test for HMAC unsubscribe token generation and validation in `backend/internal/notification/unsubscribe_test.go` — valid token round-trips, tampered token rejected, token contains correct parent ID
- [x] T007 [P] Write test for email composition helpers in `backend/internal/notification/service_test.go` — verify withdrawal request email subject/body includes child name, amount, reason, bank name, unsubscribe URL; verify chore completion email includes chore name and reward

### Implementation for Foundational

- [x] T008 [P] Add `GetByFamilyID(familyID int64) ([]models.Parent, error)` method to `backend/repositories/parent_repo.go` — GORM query: `db.Where("family_id = ?", familyID).Find(&parents)`
- [x] T009 [P] Add `UpdateNotificationPrefs(parentID int64, prefs map[string]bool) error` method to `backend/repositories/parent_repo.go` — GORM update only provided fields
- [x] T010 Create notification service struct and constructor in `backend/internal/notification/service.go` — holds brevoClient, parentRepo, familyRepo, jwtSecret; has `SendEmail(ctx, to, subject, body)` method wrapping Brevo `SendTransacEmail`
- [x] T011 [P] Implement HMAC unsubscribe token generation (`GenerateUnsubscribeToken(parentID, secret)`) and validation (`ValidateUnsubscribeToken(token, secret)`) in `backend/internal/notification/unsubscribe.go`
- [x] T012 Implement email template composition functions in `backend/internal/notification/service.go` — `composeWithdrawalRequestEmail`, `composeChoreCompletionEmail` (single), `composeChoreCompletionBatchEmail`, `composeDecisionEmail`; each returns (subject, body) using plain text templates from contracts/api.md

**Checkpoint**: Foundation ready — notification service can send emails, tokens can be generated, parent preferences can be queried/updated

---

## Phase 3: User Story 1 — Parent Receives Email When Child Requests a Withdrawal (Priority: P1) 🎯 MVP

**Goal**: When a child submits a withdrawal request, all opted-in parents in the family receive an email with the child's name, amount, and reason.

**Independent Test**: Submit a withdrawal request as a child and verify parent(s) receive an email with correct details.

### Tests for User Story 1

- [x] T013 [P] [US1] Write integration test in `backend/internal/notification/service_test.go` — `TestNotifyWithdrawalRequest`: mock Brevo client, create family with 2 parents (one opted-in, one opted-out), call `NotifyWithdrawalRequest`, verify email sent only to opted-in parent with correct subject/body
- [x] T014 [P] [US1] Write integration test in `backend/internal/withdrawal/handler_test.go` — `TestSubmitRequest_SendsNotification`: submit withdrawal request via handler, verify notification service was called with correct child name, amount, reason, family ID

### Implementation for User Story 1

- [x] T015 [US1] Implement `NotifyWithdrawalRequest(ctx, familyID, childName, amountCents, reason, bankName)` method on notification service in `backend/internal/notification/service.go` — fetch parents by familyID, filter by `NotifyWithdrawalRequests == true`, compose email, send in goroutine per parent (fire-and-forget, log errors)
- [x] T016 [US1] Add notification service as dependency to withdrawal handler in `backend/internal/withdrawal/handler.go` — update Handler struct and NewHandler constructor to accept `*notification.Service`
- [x] T017 [US1] Call `NotifyWithdrawalRequest` at end of `HandleSubmitRequest` in `backend/internal/withdrawal/handler.go` — after successful request creation and before writing response, dispatch notification with child name, amount, reason from the request

**Checkpoint**: Withdrawal request notifications fully functional and testable independently

---

## Phase 4: User Story 2 — Parent Receives Email When Child Completes a Chore (Priority: P1)

**Goal**: When a child completes a chore, all opted-in parents receive an email. Multiple completions within 5 minutes are batched into a single summary email.

**Independent Test**: Complete a chore as a child and verify parent(s) receive an email; complete multiple chores quickly and verify a single batched email.

### Tests for User Story 2

- [x] T018 [P] [US2] Write unit test for batcher in `backend/internal/notification/batcher_test.go` — `TestBatcher_SingleCompletion`: add one chore completion, wait for flush, verify single email sent with chore details
- [x] T019 [P] [US2] Write unit test for batcher in `backend/internal/notification/batcher_test.go` — `TestBatcher_BatchMultipleCompletions`: add 3 chore completions within 5-min window for same family+child, verify single batched email with all 3 chores and total reward
- [x] T020 [P] [US2] Write unit test for batcher in `backend/internal/notification/batcher_test.go` — `TestBatcher_SeparateFamilies`: add completions for different families, verify separate emails per family
- [x] T021 [P] [US2] Write integration test in `backend/internal/chore/handler_test.go` — `TestCompleteChore_SendsNotification`: complete chore via handler, verify batcher received the completion event

### Implementation for User Story 2

- [x] T022 [US2] Implement chore completion batcher in `backend/internal/notification/batcher.go` — struct with mutex-protected map of familyID+childID → []ChoreCompletion; `Add(familyID, childID, childName, choreName, rewardCents)` starts 5-min timer on first add; on flush: compose single or batched email, send to opted-in parents, clear buffer
- [x] T023 [US2] Add `Start()` and `Stop()` methods to batcher in `backend/internal/notification/batcher.go` — Start initializes the batcher; Stop flushes remaining items and cleans up timers
- [x] T024 [US2] Add notification service (with batcher) as dependency to chore handler in `backend/internal/chore/handler.go` — update Handler struct and NewHandler constructor to accept `*notification.Service`
- [x] T025 [US2] Call `service.QueueChoreCompletion` at end of `HandleCompleteChore` in `backend/internal/chore/handler.go` — after successful status change to pending_approval, queue the completion with child name, chore name, reward amount, family ID

**Checkpoint**: Chore completion notifications (with batching) fully functional and testable independently

---

## Phase 5: User Story 3 — Parent Manages Notification Preferences (Priority: P2)

**Goal**: Parents can view and toggle notification preferences per type in the settings page. All types enabled by default. One-click unsubscribe from email disables all.

**Independent Test**: Toggle preferences in settings UI and verify API updates correctly; click unsubscribe link in email and verify all preferences disabled.

### Tests for User Story 3

- [x] T026 [P] [US3] Write contract test for GET /api/settings/notifications in `backend/internal/settings/handlers_test.go` — verify returns 3 boolean fields, defaults to all true for new parent
- [x] T027 [P] [US3] Write contract test for PUT /api/settings/notifications in `backend/internal/settings/handlers_test.go` — verify partial update (send only 1 field), verify all fields returned in response, verify 400 on invalid input
- [x] T028 [P] [US3] Write contract test for GET /api/notifications/unsubscribe in `backend/internal/notification/unsubscribe_test.go` — verify valid token sets all 3 prefs to false, verify invalid token returns 400, verify HTML response

### Implementation for User Story 3

- [x] T029 [US3] Add `HandleGetNotificationPrefs` handler to `backend/internal/settings/handlers.go` — GET /api/settings/notifications; fetch parent by ID, return notify_withdrawal_requests, notify_chore_completions, notify_decisions
- [x] T030 [US3] Add `HandleUpdateNotificationPrefs` handler to `backend/internal/settings/handlers.go` — PUT /api/settings/notifications; decode JSON body, validate boolean fields, call parentRepo.UpdateNotificationPrefs, return updated prefs
- [x] T031 [US3] Update settings Handlers struct in `backend/internal/settings/handlers.go` to accept `parentRepo` dependency (currently only has familyRepo)
- [x] T032 [US3] Implement unsubscribe HTTP handler in `backend/internal/notification/unsubscribe.go` — `HandleUnsubscribe(w, r)`: extract token from query param, validate HMAC, extract parentID, set all 3 prefs to false via parentRepo, return HTML confirmation page
- [x] T033 [US3] Extend GET /api/settings response in `backend/internal/settings/handlers.go` — add `notifications` object to existing HandleGetSettings response per contracts/api.md
- [x] T034 [US3] Add notification preference TypeScript types to `frontend/src/types.ts` — NotificationPreferences interface with 3 boolean fields
- [x] T035 [P] [US3] Add API functions in `frontend/src/api.ts` — `getNotificationPrefs()` and `updateNotificationPrefs(prefs)` calling GET/PUT /api/settings/notifications
- [x] T036 [US3] Create `frontend/src/components/NotificationSettings.tsx` — toggle switches for each notification type using existing UI components (Button/Card pattern); load prefs on mount, save on toggle with optimistic update
- [x] T037 [US3] Add "Notifications" category to CATEGORIES array and render NotificationSettings component in `frontend/src/pages/SettingsPage.tsx` — add Bell icon from lucide-react, wire to new component when category is active

**Checkpoint**: Parents can manage notification preferences via settings UI and one-click unsubscribe

---

## Phase 6: User Story 4 — Parent Receives Email When Another Parent Makes a Decision (Priority: P3)

**Goal**: When a parent approves/denies a chore or withdrawal request, other parents in the family receive an informational notification email.

**Independent Test**: In a two-parent family, have one parent approve a request and verify the other parent receives an email.

### Tests for User Story 4

- [x] T038 [P] [US4] Write integration test in `backend/internal/notification/service_test.go` — `TestNotifyDecision_ExcludesActingParent`: 2-parent family, parent A approves, verify only parent B notified (not parent A)
- [x] T039 [P] [US4] Write integration test in `backend/internal/notification/service_test.go` — `TestNotifyDecision_SingleParent`: 1-parent family, parent approves, verify no email sent
- [x] T040 [P] [US4] Write integration test in `backend/internal/notification/service_test.go` — `TestNotifyDecision_RespectsPrefs`: parent B has notify_decisions=false, verify no email sent to B

### Implementation for User Story 4

- [x] T041 [US4] Implement `NotifyDecision(ctx, familyID, actingParentID, actingParentName, childName, requestType, action, amount, choreName, denialReason, bankName)` on notification service in `backend/internal/notification/service.go` — fetch parents by familyID, exclude actingParentID, filter by NotifyDecisions == true, compose decision email, send in goroutine
- [x] T042 [US4] Call `NotifyDecision` at end of `HandleApprove` and `HandleReject` in `backend/internal/chore/handler.go` — after successful approval/rejection, dispatch decision notification with parent name, child name, chore name, action
- [x] T043 [US4] Call `NotifyDecision` at end of `HandleApprove` and `HandleDeny` in `backend/internal/withdrawal/handler.go` — after successful approval/denial, dispatch decision notification with parent name, child name, amount, action, denial reason

**Checkpoint**: All 4 user stories functional — decision notifications complete the notification suite

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Wiring, integration, and final validation

- [x] T044 Wire notification service in `backend/main.go` — instantiate notification.Service with brevoClient, parentRepo, familyRepo, jwtSecret; pass to chore and withdrawal handler constructors; register GET /api/settings/notifications, PUT /api/settings/notifications, GET /api/notifications/unsubscribe routes
- [x] T045 Initialize and start chore batcher in `backend/main.go` — call batcher.Start() after service creation; defer batcher.Stop() for graceful shutdown
- [x] T046 Update TRUNCATE list in test helpers to include new notification preference columns if needed in `backend/internal/testutil/` or `backend/repositories/` test helpers
- [x] T047 Run full backend test suite: `cd backend && go test -p 1 ./...`
- [x] T048 Run frontend type check and build: `cd frontend && npx tsc --noEmit && npm run build`
- [x] T049 Run frontend lint: `cd frontend && npm run lint`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (migration + model) — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Phase 2 — can start immediately after foundational
- **User Story 2 (Phase 4)**: Depends on Phase 2 — can run in parallel with US1
- **User Story 3 (Phase 5)**: Depends on Phase 2 — can run in parallel with US1/US2
- **User Story 4 (Phase 6)**: Depends on Phase 2 + US1 handler changes (T016) — start after US1
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Independent after foundational — MVP milestone
- **US2 (P1)**: Independent after foundational — can parallel with US1
- **US3 (P2)**: Independent after foundational — can parallel with US1/US2
- **US4 (P3)**: Depends on US1's handler dependency injection (T016) since it adds notification calls to the same handlers — start after US1

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Service methods before handler integration
- Backend before frontend (US3)
- Core implementation before handler wiring

### Parallel Opportunities

- T001 + T002 (migration files) can run in parallel
- T004 + T005 + T006 + T007 (foundational tests) can all run in parallel
- T008 + T009 + T011 (repo methods + unsubscribe tokens) can run in parallel
- US1, US2, US3 test phases can run in parallel (different files)
- T034 + T035 (frontend types + API functions) can run in parallel
- T038 + T039 + T040 (US4 tests) can run in parallel

---

## Parallel Example: Foundational Phase

```text
# Parallel batch 1 — all foundational tests (different test files):
T004: GetByFamilyID test in parent_repo_test.go
T005: UpdateNotificationPrefs test in parent_repo_test.go
T006: Unsubscribe token test in unsubscribe_test.go
T007: Email composition test in service_test.go

# Parallel batch 2 — independent implementations:
T008: GetByFamilyID method in parent_repo.go
T009: UpdateNotificationPrefs method in parent_repo.go
T011: HMAC token utils in unsubscribe.go

# Sequential — depends on T008, T009:
T010: Notification service struct in service.go
T012: Email template functions in service.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (migration + model)
2. Complete Phase 2: Foundational (repo methods, notification service, token utils)
3. Complete Phase 3: User Story 1 (withdrawal request notifications)
4. **STOP and VALIDATE**: Submit a withdrawal request and verify parent receives email
5. Wire in main.go (T044) minimally for US1 and deploy

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. Add US1 → Withdrawal notifications working → **Deploy (MVP!)**
3. Add US2 → Chore notifications with batching → Deploy
4. Add US3 → Preference management UI → Deploy
5. Add US4 → Decision notifications → Deploy
6. Polish → Final validation → Deploy

### Recommended Sequence (Solo Developer)

Phase 1 → Phase 2 → Phase 3 (US1) → Phase 4 (US2) → Phase 5 (US3) → Phase 6 (US4) → Phase 7

---

## Notes

- Constitution requires TDD — all test tasks must pass red/green/refactor cycle
- Fire-and-forget: notification goroutines must not panic — use recover() in goroutine wrapper
- Batcher: use sync.Mutex for thread safety; flush on Stop() for graceful shutdown
- Unsubscribe tokens: use existing JWT_SECRET for HMAC-SHA256 signing
- Frontend: reuse existing Card/Button/Select UI components; follow SettingsPage category pattern
- Email sender: always "noreply@bankofdad.xyz" / "Bank of Dad" (matches contact handler)
