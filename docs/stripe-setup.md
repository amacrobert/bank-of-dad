# Stripe Setup Guide

Bank of Dad uses Stripe Checkout for subscriptions and Stripe Customer Portal for subscription management. No card data touches our servers.

## Prerequisites

- A [Stripe account](https://dashboard.stripe.com)
- [Stripe CLI](https://stripe.com/docs/stripe-cli) for local development: `brew install stripe/stripe-cli/stripe`

## 1. Create Product and Prices

In the Stripe Dashboard (**Test Mode** for local/staging, **Live Mode** for production):

1. Go to **Products** > **Add product**
2. Name: **Bank of Dad Plus**
3. Add two recurring prices:

| Price | Amount | Interval | Lookup Key |
|-------|--------|----------|------------|
| Monthly | $1.00  | month | `plus_monthly` |
| Annual | $10.00 | year | `plus_annual` |

Setting the **lookup key** is critical — the backend resolves prices by lookup key, not by price ID.

## 2. Configure Customer Portal

1. Go to **Settings** > **Billing** > **Customer portal**
2. Enable:
   - Cancel subscription
   - Switch plans (between monthly and annual)
   - Update payment method
3. Under **Products**, add both prices (`plus_monthly`, `plus_annual`) so subscribers can switch between them
4. Set the default return URL:
   - Local: `http://localhost:8000/settings/subscription`
   - Production: `https://bankofdad.xyz/settings/subscription`

## 3. Environment Variables

The backend requires two Stripe environment variables:

```bash
STRIPE_SECRET_KEY=sk_test_...        # API key from Developers > API keys
STRIPE_WEBHOOK_SECRET=whsec_...      # Webhook signing secret (see sections below)
```

## 4. Local Development Setup

### Start the webhook forwarder

The Stripe CLI forwards webhook events from Stripe's servers to your local backend.

```bash
# Terminal 1: Start the backend
cd backend && go run .

# Terminal 2: Log in to Stripe CLI (first time only)
stripe login

# Terminal 3: Forward webhooks to your local server
stripe listen --forward-to localhost:8001/api/stripe/webhook
```

The `stripe listen` command prints a webhook signing secret on startup:

```
> Ready! Your webhook signing secret is whsec_abc123... (^C to quit)
```

Use that value as `STRIPE_WEBHOOK_SECRET` in your local environment.

### Test the checkout flow

1. Navigate to `http://localhost:8000/settings/subscription`
2. Click **Monthly -- $2/mo** or **Annual -- $20/yr**
3. On the Stripe Checkout page, use the test card:
   - Number: `4242 4242 4242 4242`
   - Expiry: any future date
   - CVC: any 3 digits
4. After completing checkout, you'll be redirected back and the page will poll until the webhook confirms the upgrade

### Useful Stripe CLI commands

```bash
# Trigger a specific event manually
stripe trigger checkout.session.completed

# Replay recent events to your local endpoint
stripe events resend evt_xxx --webhook-endpoint we_xxx

# View recent events
stripe events list --limit 5
```

## 5. Production Setup (api.bankofdad.xyz)

### Create a webhook endpoint

1. Switch to **Live Mode** in the Stripe Dashboard
2. Go to **Developers** > **Webhooks** > **Add endpoint**
3. Endpoint URL: `https://api.bankofdad.xyz/api/stripe/webhook`
4. Select these events:
   - `checkout.session.completed`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_failed`
5. Click **Add endpoint**
6. Copy the **Signing secret** (`whsec_...`) from the endpoint detail page

### Set production environment variables

```bash
STRIPE_SECRET_KEY=sk_live_...       # Live API key from Developers > API keys
STRIPE_WEBHOOK_SECRET=whsec_...     # Signing secret from the webhook endpoint above
```

### Update Customer Portal return URL

In **Settings** > **Billing** > **Customer portal**, set the return URL to:

```
https://bankofdad.xyz/settings/subscription
```

## Verification Checklist

After setup, confirm each step works:

1. **GET /api/subscription** returns `{"account_type": "free", ...}` for an existing parent
2. **POST /api/subscription/checkout** with `{"price_lookup_key": "plus_monthly"}` returns a `checkout_url`
3. Completing checkout triggers a `checkout.session.completed` webhook that upgrades the family to Plus
4. **GET /api/subscription** now returns `{"account_type": "plus", "subscription_status": "active", ...}`
5. **POST /api/subscription/portal** returns a `portal_url` for the Plus family
6. Cancelling in the portal triggers `customer.subscription.updated` with `cancel_at_period_end: true`

## Troubleshooting

**Webhook signature verification fails (400)**
- Local: make sure `STRIPE_WEBHOOK_SECRET` matches the `whsec_...` value printed by `stripe listen`
- Production: make sure it matches the signing secret from the webhook endpoint page, not the API key

**Price not found for lookup key**
- Confirm the prices in the Stripe Dashboard have lookup keys `plus_monthly` and `plus_annual` (not just display names)
- Make sure you're in the right mode (Test vs Live) matching your `STRIPE_SECRET_KEY`

**Webhook events not arriving**
- Local: confirm `stripe listen --forward-to localhost:8001/api/stripe/webhook` is running
- Production: check the webhook endpoint's **Recent deliveries** tab in the Stripe Dashboard for errors

**Subscription not activating after checkout**
- The upgrade happens via webhook, not the redirect. If the webhook hasn't arrived yet, the frontend polls for up to 30 seconds
- Check backend logs for `checkout.session.completed` processing errors
