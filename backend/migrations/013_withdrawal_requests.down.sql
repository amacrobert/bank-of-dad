-- Drop withdrawal_requests table and indexes
DROP TABLE IF EXISTS withdrawal_requests;

-- Revert transaction_type check constraint
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_transaction_type_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_transaction_type_check
    CHECK (transaction_type IN ('deposit', 'withdrawal', 'allowance', 'interest', 'chore'));
