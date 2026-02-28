# Implementation Plan: Stripe Subscription Integration

**Branch**: `024-stripe-subscription` | **Date**: 2026-02-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/024-stripe-subscription/spec.md`

## Summary

Add optional Plus subscription ($2/month or $20/year) to Bank of Dad via Stripe. Parents can upgrade from the new `/settings/subscription` page, which redirects to Stripe Checkout for payment. Subscription management (cancel, plan switch, payment method update) is handled via Stripe's hosted Customer Portal. All subscription state changes are driven by Stripe webhooks. The family's `account_type` column tracks the current tier (free/plus). No features are gated behind Plus at this time.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `github.com/stripe/stripe-go/v82` (new), pgx/v5 (existing), Vite + Tailwind CSS 4 (existing)
**Storage**: PostgreSQL 17 — extend `families` table, add `stripe_webhook_events` table
**Testing**: `go test -p 1 ./...` (backend), `npx tsc --noEmit && npm run build` (frontend)
**Target Platform**: Web application (server + SPA)
**Project Type**: Web service with SPA frontend
**Performance Goals**: Webhook processing < 1s, subscription status page loads in < 500ms
**Constraints**: No card data touches our servers (Stripe Checkout handles PCI), webhook signature verification required
**Scale/Scope**: Single-family subscription, 4 new API endpoints, 1 new frontend component, 1 DB migration

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development — PASS

- All new backend endpoints will have contract tests
- Webhook handler will have unit tests with mock Stripe events
- Store methods will have integration tests against test DB
- Frontend component will be validated via type check + build

### II. Security-First Design — PASS

- Webhook endpoint validates `Stripe-Signature` header (no auth bypass)
- All subscription management endpoints require parent auth (`requireParent` middleware)
- No card data touches our servers (Stripe Checkout is fully hosted)
- Stripe secret key and webhook secret stored as environment variables, never in code
- Idempotent webhook processing prevents replay attacks

### III. Simplicity — PASS

- Uses Stripe Checkout (hosted page) instead of building a custom payment form
- Uses Stripe Customer Portal instead of building custom subscription management UI
- Subscription data stored as columns on existing `families` table (1:1 relationship, no separate table)
- Stripe Customer created lazily at checkout time, not for all parents
- 4 webhook events (minimal robust set), not the full Stripe event catalog
- New dependency (`stripe-go`) is the official Stripe SDK — justified by core requirement

### Post-Phase 1 Re-check — PASS

- Data model adds 6 columns to `families` + 1 small table — minimal schema change
- No new abstractions or patterns introduced — follows existing store/handler conventions
- API contracts are straightforward CRUD-style endpoints

## Project Structure

### Documentation (this feature)

```text
specs/024-stripe-subscription/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: Stripe integration research
├── data-model.md        # Phase 1: Schema changes
├── quickstart.md        # Phase 1: Setup guide
├── contracts/
│   └── api.md           # Phase 1: API endpoint contracts
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── main.go                          # Add subscription routes + handler wiring
├── internal/
│   ├── config/config.go             # Add STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET env vars
│   ├── store/
│   │   ├── family.go                # Extend Family struct + subscription query methods
│   │   └── webhook_event.go         # New: stripe webhook event idempotency store
│   └── subscription/
│       └── handlers.go              # New: checkout, portal, status, webhook handlers
├── migrations/
│   ├── 006_stripe_subscription.up.sql    # New: ALTER families + CREATE stripe_webhook_events
│   └── 006_stripe_subscription.down.sql  # New: rollback
└── go.mod                           # Add stripe-go dependency

frontend/
├── src/
│   ├── api.ts                       # Add getSubscription, createCheckout, createPortal functions
│   ├── types.ts                     # Add SubscriptionResponse interface
│   ├── components/
│   │   └── SubscriptionSettings.tsx # New: subscription status + upgrade/manage UI
│   └── pages/
│       └── SettingsPage.tsx          # Add "subscription" category to CATEGORIES array
```

**Structure Decision**: Follows existing web application structure. Backend gets a new `internal/subscription/` package (same pattern as `internal/settings/`, `internal/family/`). Frontend adds one new component following the `AccountSettings.tsx` pattern. Store methods for subscription data go on the existing `FamilyStore` since the data lives on the `families` table, plus a new `WebhookEventStore` for idempotency.

## Complexity Tracking

No constitution violations. All gates pass.
