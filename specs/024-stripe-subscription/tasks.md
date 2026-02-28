# Tasks: Stripe Subscription Integration

**Input**: Design documents from `/specs/024-stripe-subscription/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md

**Tests**: Included per constitution (Test-First Development principle). Contract tests for API endpoints, integration tests for store methods, unit tests for webhook handling.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Add Stripe dependency, configuration, and database schema

- [x] T001 Add `github.com/stripe/stripe-go/v82` dependency to `backend/go.mod` by running `cd backend && go get github.com/stripe/stripe-go/v82`
- [x] T002 [P] Add `StripeSecretKey` and `StripeWebhookSecret` string fields to Config struct and load from `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET` env vars (required, error if empty) in `backend/internal/config/config.go`
- [x] T003 [P] Create migration `backend/migrations/006_stripe_subscription.up.sql` (ALTER families to add account_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, subscription_cancel_at_period_end columns; CREATE stripe_webhook_events table) and `backend/migrations/006_stripe_subscription.down.sql` (reverse) per data-model.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Store layer extensions and handler scaffolding that all user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Extend `Family` struct with `AccountType string`, `StripeCustomerID sql.NullString`, `StripeSubscriptionID sql.NullString`, `SubscriptionStatus sql.NullString`, `SubscriptionCurrentPeriodEnd sql.NullTime`, `SubscriptionCancelAtPeriodEnd bool` fields and update all existing SELECT queries to include the new columns in `backend/internal/store/family.go`
- [x] T005 [P] Add subscription store methods to `FamilyStore` in `backend/internal/store/family.go`: `GetSubscriptionByFamilyID(familyID int64)` returning subscription fields, `GetFamilyByStripeCustomerID(customerID string)`, `GetFamilyByStripeSubscriptionID(subscriptionID string)`, `UpdateSubscriptionFromCheckout(familyID int64, stripeCustomerID, stripeSubscriptionID, status string, periodEnd time.Time)`, `UpdateSubscriptionStatus(stripeSubscriptionID, status string, periodEnd time.Time, cancelAtPeriodEnd bool)`, `ClearSubscription(stripeSubscriptionID string)` that sets account_type='free' and NULLs subscription fields
- [x] T006 [P] Create `WebhookEventStore` in `backend/internal/store/webhook_event.go` with `NewWebhookEventStore(db *sql.DB)`, `IsProcessed(eventID string) (bool, error)` using SELECT EXISTS, and `MarkProcessed(eventID, eventType string) error` using INSERT ON CONFLICT DO NOTHING
- [x] T007 Create subscription `Handlers` struct in `backend/internal/subscription/handlers.go` with `familyStore *store.FamilyStore`, `webhookEventStore *store.WebhookEventStore`, `stripeSecretKey string`, `stripeWebhookSecret string` fields, `NewHandlers(...)` constructor, and `writeJSON` helper. Add stub handler methods: `HandleGetSubscription`, `HandleCreateCheckout`, `HandleCreatePortal`, `HandleStripeWebhook`
- [x] T008 Wire subscription handlers in `backend/main.go`: instantiate `WebhookEventStore` and `subscription.NewHandlers(...)`, register `GET /api/subscription` and `POST /api/subscription/checkout` and `POST /api/subscription/portal` behind `requireParent` middleware, register `POST /api/stripe/webhook` as public route (no auth — uses Stripe signature verification). Pass `cfg.StripeSecretKey` and `cfg.StripeWebhookSecret` to handler constructor
- [x] T009 [P] Add `SubscriptionResponse` interface to `frontend/src/types.ts` with fields: `account_type: string`, `subscription_status: string | null`, `current_period_end: string | null`, `cancel_at_period_end: boolean`
- [x] T010 [P] Add `getSubscription(): Promise<SubscriptionResponse>`, `createCheckoutSession(priceLookupKey: string): Promise<{checkout_url: string}>`, and `createPortalSession(): Promise<{portal_url: string}>` functions to `frontend/src/api.ts` using existing `get`/`post` patterns

**Checkpoint**: Foundation ready — stores, handler scaffolding, routes, and frontend API layer in place

---

## Phase 3: User Story 1 — Parent Upgrades to Plus (Priority: P1) MVP

**Goal**: A parent with a free account can select monthly or annual Plus, complete Stripe Checkout, and have their account upgraded to Plus via webhook

**Independent Test**: Sign in as parent → navigate to /settings/subscription → click upgrade → complete Stripe Checkout → verify account shows Plus

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T011 [P] [US1] Write store integration tests for `GetSubscriptionByFamilyID` and `UpdateSubscriptionFromCheckout` in `backend/internal/store/family_test.go`: test that a new family returns account_type='free' with null subscription fields; test that UpdateSubscriptionFromCheckout sets account_type='plus' and populates all subscription fields; test GetFamilyByStripeCustomerID and GetFamilyByStripeSubscriptionID lookups
- [x] T012 [P] [US1] Write store integration tests for `WebhookEventStore` in `backend/internal/store/webhook_event_test.go`: test IsProcessed returns false for new event; test MarkProcessed then IsProcessed returns true; test duplicate MarkProcessed does not error
- [x] T013 [P] [US1] Write handler tests for `HandleGetSubscription` in `backend/internal/subscription/handlers_test.go`: test returns 200 with account_type='free' for new family; test returns subscription fields for family with active subscription. Use `httptest.NewRecorder` and inject auth context
- [x] T014 [P] [US1] Write handler test for `HandleCreateCheckout` in `backend/internal/subscription/handlers_test.go`: test returns 400 if family already has active subscription; test returns 400 for invalid price_lookup_key. Note: cannot test actual Stripe Checkout session creation in unit test without Stripe test key — test input validation only

### Implementation for User Story 1

- [x] T015 [US1] Implement `GetSubscriptionByFamilyID` on FamilyStore in `backend/internal/store/family.go` — query families table for subscription fields by family ID
- [x] T016 [P] [US1] Implement `IsProcessed` and `MarkProcessed` on WebhookEventStore in `backend/internal/store/webhook_event.go`
- [x] T017 [US1] Implement `UpdateSubscriptionFromCheckout` on FamilyStore in `backend/internal/store/family.go` — UPDATE families SET account_type='plus', stripe_customer_id, stripe_subscription_id, subscription_status='active', subscription_current_period_end WHERE id = $familyID
- [x] T018 [US1] Implement `HandleGetSubscription` in `backend/internal/subscription/handlers.go` — get familyID from auth context, call GetSubscriptionByFamilyID, return JSON per contract
- [x] T019 [US1] Implement `HandleCreateCheckout` in `backend/internal/subscription/handlers.go` — decode request body for `price_lookup_key`, validate against allowed values (plus_monthly, plus_annual), check family doesn't already have active subscription, set `stripe.Key`, look up price by lookup key, create Checkout Session in subscription mode with family's parent email, family ID as ClientReferenceID, success/cancel URLs pointing to frontend /settings/subscription, return checkout URL as JSON
- [x] T020 [US1] Implement `HandleStripeWebhook` in `backend/internal/subscription/handlers.go` — read raw body, verify Stripe-Signature with `webhook.ConstructEvent`, check idempotency via WebhookEventStore, dispatch on event.Type, handle `checkout.session.completed` by extracting Customer ID + Subscription ID + ClientReferenceID (family ID) from session and calling UpdateSubscriptionFromCheckout, mark event processed
- [x] T021 [US1] Create `SubscriptionSettings` component in `frontend/src/components/SubscriptionSettings.tsx` — fetch subscription status on mount via `getSubscription()`, show loading state, display "Free" plan with two upgrade buttons (Monthly $2/mo, Annual $20/yr), on click call `createCheckoutSession()` with appropriate lookup key and `window.location.href` to the returned checkout_url
- [x] T022 [US1] Add "subscription" category to `CATEGORIES` array in `frontend/src/pages/SettingsPage.tsx` with key "subscription", label "Subscription", icon `CreditCard` from lucide-react. Add `{activeCategory === "subscription" && <SubscriptionSettings />}` in the content rendering section
- [x] T023 [US1] Add checkout return handling in `frontend/src/components/SubscriptionSettings.tsx` — detect `?success=true` URL param (set in Stripe Checkout success_url), show "Processing your subscription..." message, poll `getSubscription()` every 2 seconds until account_type becomes 'plus' (max 30 seconds), then show success state

**Checkpoint**: Parents can upgrade from Free to Plus via Stripe Checkout. Webhook activates the subscription.

---

## Phase 4: User Story 2 — Parent Views Subscription Status (Priority: P2)

**Goal**: The subscription page clearly shows the current plan, billing period, renewal date, and available actions

**Independent Test**: Sign in as parent with Plus subscription → navigate to /settings/subscription → verify plan shows "Plus", billing period, and renewal date

### Implementation for User Story 2

- [x] T024 [US2] Extend `SubscriptionSettings` component in `frontend/src/components/SubscriptionSettings.tsx` to show detailed Plus subscriber view: display "Plus" plan badge, subscription status, billing period (derive monthly/annual from current_period_end relative to subscription start), next renewal date formatted in user's locale, and cancel_at_period_end status with "Cancels on [date]" message if applicable
- [x] T025 [US2] Add `account_type` field to the existing `GET /api/auth/me` response by including the family's account_type in the auth handler response in `backend/internal/auth/handlers.go` (or wherever HandleGetMe is defined) — this allows the frontend to show account type in the header/nav without an extra API call

**Checkpoint**: Subscription page shows complete status for both Free and Plus users

---

## Phase 5: User Story 3 — Parent Cancels Plus Subscription (Priority: P3)

**Goal**: A Plus parent can cancel their subscription via Stripe Customer Portal, subscription remains active until period end, then reverts to Free

**Independent Test**: Sign in as Plus parent → click "Manage Subscription" → cancel in Stripe Portal → verify subscription shows "Cancels on [date]" → after period end, account reverts to Free

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T026 [P] [US3] Write handler test for `HandleCreatePortal` in `backend/internal/subscription/handlers_test.go`: test returns 400 if family has no stripe_customer_id
- [x] T027 [P] [US3] Write store integration tests for `UpdateSubscriptionStatus` and `ClearSubscription` in `backend/internal/store/family_test.go`: test UpdateSubscriptionStatus updates status, period_end, cancel_at_period_end; test ClearSubscription resets account_type to 'free' and NULLs all subscription fields
- [x] T028 [P] [US3] Write webhook handler tests for `customer.subscription.updated` and `customer.subscription.deleted` events in `backend/internal/subscription/handlers_test.go`: test subscription.updated syncs status and cancel_at_period_end; test subscription.deleted clears subscription and sets account_type to 'free'

### Implementation for User Story 3

- [x] T029 [US3] Implement `UpdateSubscriptionStatus` and `ClearSubscription` on FamilyStore in `backend/internal/store/family.go` — UpdateSubscriptionStatus: UPDATE families SET subscription_status, subscription_current_period_end, subscription_cancel_at_period_end WHERE stripe_subscription_id; ClearSubscription: UPDATE families SET account_type='free', subscription fields to NULL WHERE stripe_subscription_id
- [x] T030 [US3] Implement `HandleCreatePortal` in `backend/internal/subscription/handlers.go` — get familyID from auth context, look up family's stripe_customer_id, return 400 if not set, create BillingPortal Session with customer ID and return_url pointing to /settings/subscription, return portal URL as JSON
- [x] T031 [US3] Add `customer.subscription.updated` handler in `HandleStripeWebhook` in `backend/internal/subscription/handlers.go` — extract subscription from event, look up family by stripe_subscription_id, call UpdateSubscriptionStatus with sub.Status, sub.CurrentPeriodEnd, sub.CancelAtPeriodEnd
- [x] T032 [US3] Add `customer.subscription.deleted` handler in `HandleStripeWebhook` in `backend/internal/subscription/handlers.go` — extract subscription from event, call ClearSubscription with stripe_subscription_id
- [x] T033 [US3] Add "Manage Subscription" button to `SubscriptionSettings` in `frontend/src/components/SubscriptionSettings.tsx` — shown only when account_type is 'plus', on click calls `createPortalSession()` and redirects to returned portal_url. Show cancellation notice if cancel_at_period_end is true

**Checkpoint**: Full subscription lifecycle works — upgrade, view, cancel, auto-revert

---

## Phase 6: User Story 4 — Parent Switches Billing Period (Priority: P4)

**Goal**: A Plus parent can switch between monthly and annual billing via the Stripe Customer Portal

**Independent Test**: Sign in as Plus parent on monthly → click "Manage Subscription" → switch to annual in Stripe Portal → verify the change is reflected

### Implementation for User Story 4

- [x] T034 [US4] Verify customer.subscription.updated webhook handler (T031) correctly syncs plan changes from the Stripe Customer Portal — no additional backend code needed since the portal handles the UI and the existing webhook handler syncs the result. Add a note in quickstart.md about configuring the portal to allow plan switching between plus_monthly and plus_annual prices

**Checkpoint**: Plan switching works end-to-end via Stripe Customer Portal

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, error handling, and validation

- [x] T035 [P] Add `invoice.payment_failed` webhook handler in `backend/internal/subscription/handlers.go` — log the failure event (family lookup by subscription ID, log warning). Status change is handled by the subsequent customer.subscription.updated event
- [x] T036 [P] Add edge case handling in `HandleStripeWebhook` in `backend/internal/subscription/handlers.go`: log and return 200 for unrecognized event types; log warning and return 200 if family not found for webhook event (don't error — Stripe will retry)
- [x] T037 Run full backend test suite: `cd backend && go test -p 1 ./...`
- [x] T038 Run frontend validation: `cd frontend && npx tsc --noEmit && npm run build`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational — this is the MVP
- **User Story 2 (Phase 4)**: Depends on US1 (needs the SubscriptionSettings component and GET /api/subscription)
- **User Story 3 (Phase 5)**: Depends on US1 (needs webhook infrastructure and SubscriptionSettings)
- **User Story 4 (Phase 6)**: Depends on US3 (needs portal and subscription.updated handler)
- **Polish (Phase 7)**: Depends on US3 completion

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational only — can start as soon as Phase 2 completes
- **User Story 2 (P2)**: Depends on US1 (extends the same frontend component and uses the same GET endpoint)
- **User Story 3 (P3)**: Depends on US1 (extends webhook handler and frontend component)
- **User Story 4 (P4)**: Depends on US3 (relies on portal handler and subscription.updated webhook)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Store methods before handlers
- Handlers before frontend
- Backend before frontend (API must exist before UI calls it)

### Parallel Opportunities

- T002 and T003 can run in parallel (different files)
- T005, T006 can run in parallel (different store files)
- T009, T010 can run in parallel with backend foundational work (different codebase)
- All US1 test tasks (T011-T014) can run in parallel
- T016 can run in parallel with T015/T017 (different store file)
- All US3 test tasks (T026-T028) can run in parallel
- T035, T036 can run in parallel (different concerns in same file but independent)

---

## Parallel Example: User Story 1

```bash
# Launch all tests for US1 together:
Task: "T011 Write store tests for subscription methods in backend/internal/store/family_test.go"
Task: "T012 Write WebhookEventStore tests in backend/internal/store/webhook_event_test.go"
Task: "T013 Write handler tests for GET /api/subscription in backend/internal/subscription/handlers_test.go"
Task: "T014 Write handler tests for POST /api/subscription/checkout in backend/internal/subscription/handlers_test.go"

# After tests written, implement store layer (can parallel across files):
Task: "T015 Implement GetSubscriptionByFamilyID in backend/internal/store/family.go"
Task: "T016 Implement WebhookEventStore methods in backend/internal/store/webhook_event.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (3 tasks)
2. Complete Phase 2: Foundational (7 tasks)
3. Complete Phase 3: User Story 1 (13 tasks — tests + implementation)
4. **STOP and VALIDATE**: Test upgrade flow end-to-end with Stripe CLI
5. Deploy/demo if ready — parents can upgrade to Plus

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. Add User Story 1 → Test upgrade flow → Deploy (MVP — parents can upgrade)
3. Add User Story 2 → Test status display → Deploy (better status visibility)
4. Add User Story 3 → Test cancellation → Deploy (full lifecycle)
5. Add User Story 4 → Test plan switching → Deploy (convenience feature)
6. Polish → Edge cases + validation → Final deploy

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Stripe Checkout and Customer Portal are hosted by Stripe — no PCI scope for our servers
- Webhook endpoint is public (no JWT) but verified via Stripe-Signature header
- Account type changes are webhook-driven only, never from client-side redirects
- Stripe Price lookup keys (plus_monthly, plus_annual) must be configured in Stripe Dashboard before testing
- Use Stripe CLI (`stripe listen --forward-to localhost:8001/api/stripe/webhook`) for local webhook testing
- Test card number: 4242 4242 4242 4242 (any future expiry, any CVC)
