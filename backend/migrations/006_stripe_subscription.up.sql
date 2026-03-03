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
