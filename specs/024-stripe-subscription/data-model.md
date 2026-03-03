# Data Model: Stripe Subscription Integration

**Feature**: 024-stripe-subscription
**Date**: 2026-02-26

## Entity Changes

### Family (modified)

The existing `families` table is extended with subscription-related columns.

| Field | Type | Nullable | Default | Description |
|-------|------|----------|---------|-------------|
| id | SERIAL | no | auto | Primary key (existing) |
| slug | TEXT | no | — | Family URL slug (existing) |
| timezone | TEXT | no | 'America/New_York' | Family timezone (existing) |
| created_at | TIMESTAMPTZ | no | NOW() | Creation timestamp (existing) |
| **account_type** | TEXT | no | 'free' | 'free' or 'plus' — the active access tier |
| **stripe_customer_id** | TEXT | yes | NULL | Stripe Customer ID, set on first checkout |
| **stripe_subscription_id** | TEXT | yes | NULL | Stripe Subscription ID for active/cancelled subscription |
| **subscription_status** | TEXT | yes | NULL | Mirrors Stripe status: 'active', 'past_due', 'canceled', NULL |
| **subscription_current_period_end** | TIMESTAMPTZ | yes | NULL | When the current billing period ends |
| **subscription_cancel_at_period_end** | BOOLEAN | no | FALSE | Whether subscription cancels at period end |

**Constraints**:
- `stripe_customer_id` is UNIQUE (one Stripe customer per family)
- `stripe_subscription_id` is UNIQUE (one active subscription per family)
- `account_type` CHECK constraint: must be 'free' or 'plus'

**Indexes**:
- Unique index on `stripe_customer_id` (for webhook lookups)
- Unique index on `stripe_subscription_id` (for webhook lookups)

### Stripe Webhook Events (new)

Tracks processed webhook events for idempotency.

| Field | Type | Nullable | Default | Description |
|-------|------|----------|---------|-------------|
| stripe_event_id | TEXT | no | — | Stripe event ID (primary key) |
| event_type | TEXT | no | — | Event type (e.g., 'checkout.session.completed') |
| processed_at | TIMESTAMPTZ | no | NOW() | When the event was processed |

## State Transitions

### Account Type

```
free ──[checkout.session.completed]──► plus
plus ──[subscription deleted / period end]──► free
```

### Subscription Status

```
NULL ──[checkout.session.completed]──► active
active ──[payment failed]──► past_due
past_due ──[payment succeeded]──► active
active ──[cancel requested]──► canceled (cancel_at_period_end = true)
canceled ──[period ends]──► NULL (account_type → free)
past_due ──[all retries fail]──► NULL (account_type → free)
```

Note: "canceled" in Stripe means "will cancel at period end" when `cancel_at_period_end` is true. The subscription is still active until the period ends. When the period ends, Stripe sends `customer.subscription.deleted` and we set status to NULL and account_type to 'free'.

## Migration

**File**: `backend/migrations/006_stripe_subscription.up.sql`

```sql
-- Extend families table with subscription fields
ALTER TABLE families ADD COLUMN account_type TEXT NOT NULL DEFAULT 'free';
ALTER TABLE families ADD CONSTRAINT families_account_type_check CHECK (account_type IN ('free', 'plus'));
ALTER TABLE families ADD COLUMN stripe_customer_id TEXT UNIQUE;
ALTER TABLE families ADD COLUMN stripe_subscription_id TEXT UNIQUE;
ALTER TABLE families ADD COLUMN subscription_status TEXT;
ALTER TABLE families ADD COLUMN subscription_current_period_end TIMESTAMPTZ;
ALTER TABLE families ADD COLUMN subscription_cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE;

-- Webhook idempotency table
CREATE TABLE stripe_webhook_events (
    stripe_event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**File**: `backend/migrations/006_stripe_subscription.down.sql`

```sql
DROP TABLE IF EXISTS stripe_webhook_events;
ALTER TABLE families DROP COLUMN IF EXISTS subscription_cancel_at_period_end;
ALTER TABLE families DROP COLUMN IF EXISTS subscription_current_period_end;
ALTER TABLE families DROP COLUMN IF EXISTS subscription_status;
ALTER TABLE families DROP COLUMN IF EXISTS stripe_subscription_id;
ALTER TABLE families DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE families DROP CONSTRAINT IF EXISTS families_account_type_check;
ALTER TABLE families DROP COLUMN IF EXISTS account_type;
```

## Relationships

```
Family 1 ──── 0..1 Stripe Customer (via stripe_customer_id)
Family 1 ──── 0..1 Stripe Subscription (via stripe_subscription_id)
Stripe Webhook Event ──── 0..1 Family (lookup via subscription/customer ID in event payload)
```
