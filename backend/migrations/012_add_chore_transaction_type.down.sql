-- Revert: remove 'chore' from the allowed transaction_type values
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_transaction_type_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_transaction_type_check
    CHECK (transaction_type IN ('deposit', 'withdrawal', 'allowance', 'interest'));
