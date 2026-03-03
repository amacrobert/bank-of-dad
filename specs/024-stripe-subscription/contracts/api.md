# API Contracts: Stripe Subscription Integration

**Feature**: 024-stripe-subscription
**Date**: 2026-02-26

## Endpoints

### GET /api/subscription

Returns the family's current subscription status. Used by the subscription settings page to display plan info.

**Auth**: Required (parent only)

**Response 200**:
```json
{
  "account_type": "free",
  "subscription_status": null,
  "current_period_end": null,
  "cancel_at_period_end": false
}
```

**Response 200 (subscribed)**:
```json
{
  "account_type": "plus",
  "subscription_status": "active",
  "current_period_end": "2026-03-26T00:00:00Z",
  "cancel_at_period_end": false
}
```

**Response 200 (cancelling)**:
```json
{
  "account_type": "plus",
  "subscription_status": "active",
  "current_period_end": "2026-03-26T00:00:00Z",
  "cancel_at_period_end": true
}
```

---

### POST /api/subscription/checkout

Creates a Stripe Checkout Session for upgrading to Plus. Returns a URL to redirect the user to.

**Auth**: Required (parent only)

**Request body**:
```json
{
  "price_lookup_key": "plus_monthly"
}
```

Valid values: `"plus_monthly"`, `"plus_annual"`

**Response 200**:
```json
{
  "checkout_url": "https://checkout.stripe.com/c/pay/cs_test_..."
}
```

**Response 400** (already subscribed):
```json
{
  "error": "Family already has an active subscription"
}
```

**Response 400** (invalid price key):
```json
{
  "error": "Invalid price lookup key"
}
```

---

### POST /api/subscription/portal

Creates a Stripe Customer Portal session for managing an existing subscription. Returns a URL to redirect the user to.

**Auth**: Required (parent only)

**Response 200**:
```json
{
  "portal_url": "https://billing.stripe.com/p/session/..."
}
```

**Response 400** (no subscription):
```json
{
  "error": "No active subscription to manage"
}
```

---

### POST /api/stripe/webhook

Receives webhook events from Stripe. This endpoint is public (no JWT auth) but validates the Stripe-Signature header.

**Auth**: None (Stripe signature verification)

**Request**: Raw body from Stripe with `Stripe-Signature` header.

**Response 200**: Always returns 200 to acknowledge receipt (even if event is ignored or already processed).

**Response 400**: Invalid signature.

**Handled events**:
- `checkout.session.completed` → Activate subscription, store Stripe IDs
- `customer.subscription.updated` → Sync status, period end, cancel_at_period_end
- `customer.subscription.deleted` → Revert to free
- `invoice.payment_failed` → Log failure (status change handled by subscription.updated)
