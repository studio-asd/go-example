DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS movements;
DROP TABLE IF EXISTS accounts_balance;
DROP TABLE IF EXISTS accounts_ledger;

-- accounts_ledger index will be used for several cases:
-- 1. We want to retrieve all transactions within a movement_id. Possibly sorted by timestamp.
-- 2. We want to retrieve all transactions within an account_id. Possibly sorted by timestamp.
DROP INDEX IF EXISTS idx_accounts_ledger_movement_id;
DROP INDEX IF EXISTS idx_accounts_ledger_account_id;
DROP INDEX IF EXISTS idx_accounts_ledger_client_id;
DROP INDEX IF EXISTS idx_accounts_balance_parent_account_id;

DROP TABLE IF EXISTS wallet_accounts;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallet_transfers;