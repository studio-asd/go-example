-- drop tables.
DROP TABLE IF EXISTS wallet_accounts;

DROP TABLE IF EXISTS wallet_transactions;

DROP TABLE IF EXISTS wallet_transfers;

-- create tables.
CREATE TABLE IF NOT EXISTS wallet_users (
    user_id VARCHAR NOT NULL,
    user_status INT NOT NULL,
    intermediary_wallet_id VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS wallet_accounts (
    wallet_id VARCHAR PRIMARY KEY,
    ledger_account_id VARCHAR NOT NULL,
    user_id VARCHAR NOT NULL,
    wallet_status INT NOT NULL,
    -- wallet_owner defines who is the owner of the wallet, there are two type of owner
    -- the 1st one is 'system', and the 2nd one is 'user'. Only wallet with the owner of
    -- 'system' can goes below 0.
    --
    -- The walelt_owner usually combined with the wallet_type as each owner can have a specialized
    -- wallet type for their needs.
    wallet_owner INT NOT NULL,
    -- wallet_type defines the types of the wallet based on the owner.
    wallet_type INT NOT NULL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    transaction_id VARCHAR PRIMARY KEY,
    transaction_type INT NOT NULL,
    transaction_status INT NOT NULL,
    created_at TIMESTAMPTZNOT NULL,
    updated_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS wallet_transfers (
    transaction_id VARCHAR PRIMARY KEY,
    from_wallet_id VARCHAR NOT NULL,
    to_wallet_id VARCHAR NOT NULL,
    amount NUMERIC NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_deposits (
    transaction_id VARCHAR PRIMARY KEY,
    -- deposit_wallet_id is where the money is coming from to user.
    deposit_wallet_id VARCHAR NOT NULL,
    user_wallet_id NOT NULL,
    amount NUMERIC NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS wallet_withdrawals (
    transaction_id VARCHAR PRIMARY KEY,
    -- withdrawal_wallet_id is where the money is coming to from user.
    withdrawal_wallet_id VARCHAR NOT NULL,
    -- user_wallet_id is where the money is coming from.
    user_wallet_id VARCHAR NOT NULL,
    -- user_intermediary_wallet_id is user as an intermediary wallet when withdrawal happens. The withdrawal
    -- need a status transition because the transfer to the outside of the system is not instant. And most likely
    -- we need the payment gateway to confirm it first with the bank. Because there is a transition, there will be
    -- also a chance that the transaction will failed and the money need to go back to its user.
    -- Transferring money immediately to the withdrwal_wallet works but it is less convenience as money can go back
    -- and forth depends on the status. Since withdrawal_wallet is a global wallet and there are a lot of transactions
    -- going on, its gonna be more convenient if all transactions that goes to the withdrawal_wallet is only the successful one.
    user_intermediary_wallet_id VARCHAR NOT NULL,
    amount NUMERIC NOT NULL,
    withdrawal_fee NUMERIC NOT NULL,
    final_amount NUMERIC NOT NULL,
    withdrawal_status INT NOT NULL,
    withdrawal_channel INT NOT NULL,
    withdrawal_via_pg BOOLEAN NOT NULL,
    withdrawal_pg_vendor VARCHAR,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS wallet_bank_withdrawals (
    transaction_id VARCHAR NOT NULL,
    bank_name VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_ewallet_withdrawals (
    transaction_id VARCHAR NOT NULL,
    ewallet_name VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
