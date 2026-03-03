# Quickstart: Stripe Subscription Integration

**Feature**: 024-stripe-subscription
**Date**: 2026-02-26

## Prerequisites

1. **Stripe account**: Create a Stripe account at https://dashboard.stripe.com
2. **Stripe CLI** (for local webhook testing): `brew install stripe/stripe-cli/stripe`

## Stripe Dashboard Setup

### 1. Create Products and Prices

In the Stripe Dashboard (Test Mode):

1. Go to **Products** → **Add product**
2. Create a product called "Bank of Dad Plus"
3. Add two prices:
   - **Monthly**: $2.00/month, lookup key: `plus_monthly`
   - **Annual**: $20.00/year, lookup key: `plus_annual`

### 2. Configure Customer Portal

1. Go to **Settings** → **Billing** → **Customer portal**
2. Enable these features:
   - Cancel subscription
   - Switch plans (between monthly and annual)
   - Update payment method
3. Add both prices (plus_monthly, plus_annual) to the portal's subscription update options
4. Set the return URL to `http://localhost:8000/settings/subscription`

### 3. Create Webhook Endpoint (Production)

1. Go to **Developers** → **Webhooks** → **Add endpoint**
2. URL: `https://your-domain.com/api/stripe/webhook`
3. Select events:
   - `checkout.session.completed`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_failed`
4. Copy the **Signing secret** (starts with `whsec_`)

## Environment Variables

Add these to your backend environment:

```bash
STRIPE_SECRET_KEY=sk_test_...        # From Stripe Dashboard → Developers → API keys
STRIPE_WEBHOOK_SECRET=whsec_...      # From webhook endpoint setup (or Stripe CLI for local)
STRIPE_PRICE_MONTHLY_LOOKUP=plus_monthly   # Price lookup key for $2/month
STRIPE_PRICE_ANNUAL_LOOKUP=plus_annual     # Price lookup key for $20/year
```

## Local Development

### Running with Stripe CLI (webhook forwarding)

```bash
# Terminal 1: Start the backend
cd backend && go run .

# Terminal 2: Forward Stripe webhooks to local server
stripe listen --forward-to localhost:8001/api/stripe/webhook

# The CLI will print a webhook signing secret (whsec_...) — use this as STRIPE_WEBHOOK_SECRET locally
```

### Testing the checkout flow

```bash
# Trigger a checkout (use Stripe's test card 4242 4242 4242 4242, any future expiry, any CVC)
# Navigate to http://localhost:8000/settings/subscription and click Upgrade
```

## Database Migration

The migration `006_stripe_subscription` runs automatically on backend startup. It adds subscription columns to the `families` table and creates the `stripe_webhook_events` table.

## Verification

After setup, verify:
1. `GET /api/subscription` returns `{"account_type": "free", ...}` for an existing parent
2. `POST /api/subscription/checkout` with `{"price_lookup_key": "plus_monthly"}` returns a checkout URL
3. Completing checkout triggers a webhook that updates the family to Plus
4. `POST /api/subscription/portal` returns a portal URL for a Plus family
