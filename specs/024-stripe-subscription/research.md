# Research: Stripe Subscription Integration

**Feature**: 024-stripe-subscription
**Date**: 2026-02-26

## Decision 1: Stripe Go SDK Version and Client Pattern

**Decision**: Use `github.com/stripe/stripe-go/v82` with the legacy sub-package pattern (global `stripe.Key`).

**Rationale**: The v82 sub-package pattern (`checkout/session`, `billingportal/session`, `webhook`) is the most documented and stable approach. While v84 introduces a new instance-based client (`stripe.NewClient()`), the sub-package pattern is well-tested and simpler for our use case. We use v82 since that's the most widely battle-tested version with the sub-package approach.

**Alternatives considered**:
- v84 instance-based client (`stripe.NewClient()`): Future-proof but newer, less community documentation. Can migrate later if needed.
- Embedded Stripe.js / Stripe Elements: Requires PCI-DSS SAQ A-EP. Stripe Checkout (hosted) keeps us at SAQ A — no card data touches our servers.

## Decision 2: Checkout vs Embedded Payment Form

**Decision**: Use Stripe Checkout (hosted payment page), not embedded Stripe Elements.

**Rationale**: Stripe Checkout is a full-page redirect to Stripe's hosted payment page. This completely eliminates PCI compliance scope for our backend — no card numbers, no tokenization code, no custom payment UI to maintain. The frontend simply redirects to a URL returned by our backend.

**Alternatives considered**:
- Stripe Elements (embedded form): More control over UI but increases PCI scope and frontend complexity. Not worth it for a simple $2/month subscription.
- Custom payment form: Maximum complexity and security risk. Not appropriate.

## Decision 3: Subscription Management via Customer Portal

**Decision**: Use Stripe's hosted Customer Portal for subscription management (cancel, change plan, update payment method).

**Rationale**: The Customer Portal is a Stripe-hosted page that handles cancellation, plan switching, payment method updates, and invoice history. This eliminates the need to build custom management UI, reduces our attack surface for payment-related actions, and automatically stays current with Stripe's UI improvements. Our subscription settings page shows status and provides a link to the portal.

**Alternatives considered**:
- Custom management UI with direct Stripe API calls: Much more code to write and maintain, more error handling, more security surface area. The portal does everything we need.
- No management at all (cancel by contacting support): Poor user experience, violates consumer protection norms.

## Decision 4: Data Model — Columns on Families vs Separate Table

**Decision**: Add subscription columns directly to the `families` table rather than creating a separate `subscriptions` table.

**Rationale**: The relationship is strictly 1:1 (one subscription per family). Adding columns avoids a JOIN on every subscription status check and keeps queries simple. The spec mentions a Subscription entity conceptually, but the implementation is simpler as columns on families. If we ever need subscription history (e.g., tracking past subscriptions), we can add a history table later.

**Alternatives considered**:
- Separate `subscriptions` table: Cleaner entity separation but adds JOIN complexity for a 1:1 relationship. YAGNI — we don't need subscription history yet.
- Separate table with materialized `account_type` on families: Worst of both worlds — two places to update.

## Decision 5: Stripe Customer Creation Timing

**Decision**: Create the Stripe Customer lazily, at checkout time, not when the parent signs up.

**Rationale**: Most parents may never subscribe. Eagerly creating Stripe Customers for all parents pollutes the Stripe Dashboard and creates orphaned customer records. The Checkout Session can auto-create a customer with `CustomerCreation: "always"`, and we store the returned customer ID when the `checkout.session.completed` webhook fires.

**Alternatives considered**:
- Eager creation at signup: Simpler checkout flow (Customer already exists) but creates many unused Stripe records.
- Create just before checkout: More control over the Customer record but adds an extra API call. Not needed since Checkout handles it.

## Decision 6: Webhook Events to Listen For

**Decision**: Listen for these 4 events (minimal robust set):

| Event | Purpose |
|-------|---------|
| `checkout.session.completed` | Activate subscription, store Stripe customer/subscription IDs |
| `customer.subscription.updated` | Sync status changes (active → past_due, plan changes, renewals) |
| `customer.subscription.deleted` | Revoke access, revert to free |
| `invoice.payment_failed` | Log payment failure, subscription goes past_due via subscription.updated |

**Rationale**: This covers the full lifecycle. `invoice.paid` is redundant with `customer.subscription.updated` (which fires on renewal too). We keep `invoice.payment_failed` for logging/alerting but the actual status change comes through `subscription.updated`.

**Alternatives considered**:
- Minimal (only checkout.session.completed + subscription.deleted): Misses dunning/past_due state transitions.
- Maximum (all invoice + subscription events): Over-processing; most events duplicate information.

## Decision 7: Webhook Idempotency

**Decision**: Use a `stripe_webhook_events` table to track processed event IDs.

**Rationale**: Stripe retries failed webhook deliveries for up to 72 hours. Without idempotency, duplicate events could cause double-processing. A simple table with `stripe_event_id` as primary key and `ON CONFLICT DO NOTHING` makes duplicate processing harmless.

## Decision 8: Account Type Update Flow

**Decision**: Account type is updated exclusively via webhooks, never from client-side redirect.

**Rationale**: The Stripe Checkout success redirect happens before Stripe sends the webhook, and the redirect alone doesn't confirm payment. A user could manipulate the redirect URL. The webhook carries a cryptographically verified event. The frontend polls or re-fetches subscription status after redirect to pick up the webhook-driven update.

**Flow**:
1. Parent clicks "Upgrade" → backend creates Checkout Session → redirect to Stripe
2. Parent completes payment → Stripe redirects to success URL
3. Frontend shows "Processing..." and polls `GET /api/subscription`
4. Stripe sends `checkout.session.completed` webhook → backend updates `account_type` to 'plus'
5. Frontend poll detects the change → shows "Plus" status

## Decision 9: Stripe Price Configuration

**Decision**: Use Stripe Price lookup keys (`plus_monthly`, `plus_annual`) rather than hardcoded Price IDs.

**Rationale**: Price IDs are opaque strings that differ between test and production Stripe accounts. Lookup keys are human-readable, can be updated in the Stripe Dashboard without code changes, and make the code self-documenting. Prices are created in the Stripe Dashboard, not in code.

**Alternatives considered**:
- Hardcoded Price IDs in env vars: Works but fragile and less readable.
- Prices created via API: Over-engineering for 2 fixed prices.
