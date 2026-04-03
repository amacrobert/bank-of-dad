ALTER TABLE parents
    ADD COLUMN notify_withdrawal_requests BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN notify_chore_completions BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN notify_decisions BOOLEAN NOT NULL DEFAULT TRUE;
