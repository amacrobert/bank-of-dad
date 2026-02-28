DROP TABLE IF EXISTS stripe_webhook_events;
ALTER TABLE families DROP COLUMN IF EXISTS subscription_cancel_at_period_end;
ALTER TABLE families DROP COLUMN IF EXISTS subscription_current_period_end;
ALTER TABLE families DROP COLUMN IF EXISTS subscription_status;
ALTER TABLE families DROP COLUMN IF EXISTS stripe_subscription_id;
ALTER TABLE families DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE families DROP CONSTRAINT IF EXISTS families_account_type_check;
ALTER TABLE families DROP COLUMN IF EXISTS account_type;
