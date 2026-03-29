-- Withdrawal requests table
CREATE TABLE withdrawal_requests (
    id BIGSERIAL PRIMARY KEY,
    child_id BIGINT NOT NULL REFERENCES children(id),
    family_id BIGINT NOT NULL REFERENCES families(id),
    amount_cents INTEGER NOT NULL,
    reason VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    denial_reason VARCHAR(500),
    reviewed_by_parent_id BIGINT REFERENCES parents(id),
    reviewed_at TIMESTAMPTZ,
    transaction_id BIGINT REFERENCES transactions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_wr_amount_positive CHECK (amount_cents > 0),
    CONSTRAINT chk_wr_status_valid CHECK (status IN ('pending', 'approved', 'denied', 'cancelled'))
);

-- Indexes
CREATE INDEX idx_withdrawal_requests_child_id ON withdrawal_requests(child_id);
CREATE INDEX idx_withdrawal_requests_family_id_status ON withdrawal_requests(family_id, status);

-- Enforce one pending request per child
CREATE UNIQUE INDEX uq_withdrawal_requests_child_pending ON withdrawal_requests(child_id) WHERE status = 'pending';

-- Add 'withdrawal_request' to the allowed transaction_type values
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_transaction_type_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_transaction_type_check
    CHECK (transaction_type IN ('deposit', 'withdrawal', 'allowance', 'interest', 'chore', 'withdrawal_request'));
