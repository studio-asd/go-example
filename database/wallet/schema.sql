-- drop tables.
DROP TABLE IF EXISTS wallet_accounts;

DROP TABLE IF EXISTS wallet_transactions;

DROP TABLE IF EXISTS wallet_transfers;

-- create tables.
CREATE TABLE IF NOT EXISTS wallet_users (
    user_id varchar NOT NULL,
    user_type varchar NOT NULL,
    user_status int NOT NULL,
    -- intermediary_wallet_id is an intermediary wallet for state transition for all wallets inside a given user. The intermediary
    -- wallet is useful for state transition because its lock contention is per user-basis before we are transitioning the funds
    -- to a global system wallet like withdrawal.
    intermediary_wallet_id varchar NOT NULL,
    -- chargeback_wallet_id is a wallet to chargeback the user for any case of reversal. For example when payment reversal happens and
    -- the user doesn't  have enough money to be taken, we will charge the money via the chargeback wallet to ensure user can only
    -- use their wallet if they pay the amount of the chargeback.
    chargeback_wallet_id varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_accounts (
    wallet_id varchar PRIMARY KEY,
    ledger_account_id varchar NOT NULL,
    user_id varchar NOT NULL,
    wallet_status int NOT NULL,
    -- wallet_owner defines who is the owner of the wallet, there are two type of owner
    -- the 1st one is 'system', and the 2nd one is 'user'. Only wallet with the owner of
    -- 'system' can goes below 0.
    --
    -- The walelt_owner usually combined with the wallet_type as each owner can have a specialized
    -- wallet type for their needs.
    wallet_owner int NOT NULL,
    -- wallet_type defines the types of the wallet based on the owner.
    wallet_type int NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    transaction_id varchar PRIMARY KEY,
    transaction_type int NOT NULL,
    transaction_status int NOT NULL,
    idempotency_key varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_transfers (
    transaction_id varchar PRIMARY KEY,
    from_wallet_id varchar NOT NULL,
    to_wallet_id varchar NOT NULL,
    amount numeric NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_deposits (
    transaction_id varchar PRIMARY KEY,
    -- deposit_wallet_id is where the money is coming from to user.
    deposit_wallet_id varchar NOT NULL,
    user_wallet_id varchar NOT NULL,
    amount numeric NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_withdrawals (
    transaction_id varchar PRIMARY KEY,
    -- withdrawal_wallet_id is where the money is coming to from user.
    withdrawal_wallet_id varchar NOT NULL,
    -- user_wallet_id is where the money is coming from.
    user_wallet_id varchar NOT NULL,
    -- user_intermediary_wallet_id is user as an intermediary wallet when withdrawal happens. The withdrawal
    -- need a status transition because the transfer to the outside of the system is not instant. And most likely
    -- we need the payment gateway to confirm it first with the bank. Because there is a transition, there will be
    -- also a chance that the transaction will failed and the money need to go back to its user.
    -- Transferring money immediately to the withdrwal_wallet works but it is less convenience as money can go back
    -- and forth depends on the status. Since withdrawal_wallet is a global wallet and there are a lot of transactions
    -- going on, its gonna be more convenient if all transactions that goes to the withdrawal_wallet is only the successful one.
    user_intermediary_wallet_id varchar NOT NULL,
    amount numeric NOT NULL,
    withdrawal_fee numeric NOT NULL,
    final_amount numeric NOT NULL,
    withdrawal_status int NOT NULL,
    withdrawal_channel int NOT NULL,
    withdrawal_via_pg boolean NOT NULL,
    withdrawal_pg_vendor varchar,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_bank_withdrawals (
    transaction_id varchar NOT NULL,
    bank_name varchar NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_ewallet_withdrawals (
    transaction_id varchar NOT NULL,
    ewallet_name varchar NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_reversal (
    transaction_id varchar NOT NULL,
    reversed_transaction_id varchar NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_chargebacks (
    transaction_id varchar NOT NULL,
    chargeback_type int NOT NULL,
    amount numeric NOT NULL,
    reason TEXT NOT NULL,
    created_at timestamptz NOT NULL
);
