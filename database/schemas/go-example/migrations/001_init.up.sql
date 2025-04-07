-- accounts is used to store all user accounts.
CREATE TABLE IF NOT EXISTS accounts (
    "account_id" varchar PRIMARY KEY,
    "name" varchar NOT NULL,
    -- description is a short description for the account. Usually, the account need a name to identify
    -- and a description to explain the purpose of the account.
    "description" text NOT NULL,
    "parent_account_id" varchar,
    "currency_id" int NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz
);

-- movements is used to store all movement records.
CREATE TABLE IF NOT EXISTS movements (
    -- movement_id is UUID_v7.
    "movement_id" varchar PRIMARY KEY,
    "idempotency_key" varchar UNIQUE NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz,
    -- reversed_at is a marker for reversal. If the reversed_at is not null then the movement is reversed.
    "reversed_at" timestamptz,
    -- reversal_movement_id is the id of movement where the reversal is being performed.
    "reversal_movement_id" varchar
);

-- accounts_balance is used to store the latest state of user's balance. This table will be used for user
-- balance fast retrieval and for locking the user balance for movement.
CREATE TABLE IF NOT EXISTS accounts_balance (
    "account_id" varchar PRIMARY KEY,
    "parent_account_id" varchar,
    "currency_id" int NOT NULL,
    -- allow_negative allows some accounts to have negative balance. For example, for the funding
    -- account we might allow the account to have negative balance.
    "allow_negative" boolean NOT NULL,
    "balance" numeric NOT NULL,
    "last_movement_id" varchar NOT NULL,
    "last_ledger_id" varchar NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz
);

-- accounts_balance_history is a historical balance changes for an account based on movement_id. The historical
-- balance is per movement_id and not per ledger_id because we pre-calculates everything inside the system. And because
-- we are building the ledger optimistically, the historical amount can changed by the time we calculate because
-- we are not locking the balance up-front(as this will be expensive). So it makes more sense to create the history
-- based on movement_id because we will do that in bulk rather than ledger_id.
--
-- This table can be used for various things like retrieving opening and ending balance of an account at a given time.
CREATE TABLE IF NOT EXISTS accounts_balance_history (
    "history_id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "movement_id" varchar NOT NULL,
    -- ledger_id is the id of where the balance is being summarized, the SUM of the balance in the accounts_ledger should be
    -- the same if we sumarize everything up to this ledger_id.
    "ledger_id" varchar NOT NULL,
    "account_id" varchar NOT NULL,
    "balance" numeric NOT NULL,
    "previous_balance" numeric NOT NULL,
    "previous_movement_id" varchar NOT NULL,
    "previous_ledger_id" varchar NOT NULL,
    "created_at" timestamptz NOT NULL
);

-- accounts_ledger is used to store all ledger changes for a specific account. A single transaction
-- can possibly affecting multiple acounts in the ledger.
--
-- Row in this table is IMMUTABLE and should NOT be updated.
CREATE TABLE IF NOT EXISTS accounts_ledger (
    "internal_id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- the ledger id is the secondary unique key of the accounts_ledger. Even though we have unique constraint, but the client
    -- can always refer themselves to this unique id when it comes to reconciliation.
    "ledger_id" varchar UNIQUE NOT NULL,
    "movement_id" varchar NOT NULL,
    "account_id" varchar NOT NULL,
    -- movement_seuqnce is the ordered sequence of movement inside a movement_id.
    "movement_sequence" int NOT NULL,
    "currency_id" int NOT NULL,
    "amount" numeric NOT NULL,
    -- previous_ledger_id will be used to track the sequence of the ledger entries of a user.
    "previous_ledger_id" varchar NOT NULL,
    "created_at" timestamptz NOT NULL,
    -- client_id is an identifier that the client can use in case they want to link their ids to per-ledger-row. With this, there are
    -- many cases they can use with the ledger.
    --
    -- For example, the client want to use the ledger for transfer. The client might want to have a separate transfer table that have its own
    -- id, use that id when creating the transaction to the ledger.
    "client_id" varchar,
    -- reversal_of is a ledger_id that being reversed by this ledger_id.
    "reversal_of" varchar
);

CREATE TABLE IF NOT EXISTS reversed_movements (
    "movement_id" varchar NOT NULL,
    "reversal_movement_id" varchar NOT NULL,
    "reversal_reason" text NOT NULL,
    "created_at" timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_accounts_balance_parent_account_id ON accounts_balance ("parent_account_id");

CREATE INDEX IF NOT EXISTS idx_accounts_ledger_movement_id ON accounts_ledger ("movement_id");

CREATE INDEX IF NOT EXISTS idx_accounts_ledger_account_id ON accounts_ledger ("account_id");

CREATE INDEX IF NOT EXISTS idx_accounts_ledger_client_id ON accounts_ledger ("client_id")
WHERE
    "client_id" IS NOT NULL;

CREATE TABLE IF NOT EXISTS wallet_users (
    user_id varchar not null,
    user_type varchar not null,
    user_status int not null,
    -- intermediary_wallet_id is an intermediary wallet for state transition for all wallets inside a given user. the intermediary
    -- wallet is useful for state transition because its lock contention is per user-basis before we are transitioning the funds
    -- to a global system wallet like withdrawal.
    intermediary_wallet_id varchar not null,
    -- chargeback_wallet_id is a wallet to chargeback the user for any case of reversal. for example when payment reversal happens and
    -- the user doesn't  have enough money to be taken, we will charge the money via the chargeback wallet to ensure user can only
    -- use their wallet if they pay the amount of the chargeback.
    chargeback_wallet_id varchar not null,
    created_at timestamptz not null,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_accounts (
    wallet_id varchar primary key,
    ledger_account_id varchar not null,
    user_id varchar not null,
    wallet_status int not null,
    -- wallet_owner defines who is the owner of the wallet, there are two type of owner
    -- the 1st one is 'system', and the 2nd one is 'user'. only wallet with the owner of
    -- 'system' can goes below 0.
    --
    -- the walelt_owner usually combined with the wallet_type as each owner can have a specialized
    -- wallet type for their needs.
    wallet_owner int not null,
    -- wallet_type defines the types of the wallet based on the owner.
    wallet_type int not null,
    created_at timestamptz not null,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    transaction_id varchar primary key,
    transaction_type int not null,
    transaction_status int not null,
    idempotency_key varchar not null,
    created_at timestamptz not null,
    updated_at timestamptz,
    finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_transfers (
    transaction_id varchar primary key,
    from_wallet_id varchar not null,
    to_wallet_id varchar not null,
    amount numeric not null,
    created_at timestamptz not null
);

CREATE TABLE IF NOT EXISTS wallet_deposits (
    transaction_id varchar primary key,
    -- deposit_wallet_id is where the money is coming from to user.
    deposit_wallet_id varchar not null,
    user_wallet_id varchar not null,
    amount numeric not null,
    created_at timestamptz not null,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_withdrawals (
    transaction_id varchar primary key,
    -- withdrawal_wallet_id is where the money is coming to from user.
    withdrawal_wallet_id varchar not null,
    -- user_wallet_id is where the money is coming from.
    user_wallet_id varchar not null,
    -- user_intermediary_wallet_id is user as an intermediary wallet when withdrawal happens. the withdrawal
    -- need a status transition because the transfer to the outside of the system is not instant. and most likely
    -- we need the payment gateway to confirm it first with the bank. because there is a transition, there will be
    -- also a chance that the transaction will failed and the money need to go back to its user.
    -- transferring money immediately to the withdrwal_wallet works but it is less convenience as money can go back
    -- and forth depends on the status. since withdrawal_wallet is a global wallet and there are a lot of transactions
    -- going on, its gonna be more convenient if all transactions that goes to the withdrawal_wallet is only the successful one.
    user_intermediary_wallet_id varchar not null,
    amount numeric not null,
    withdrawal_fee numeric not null,
    final_amount numeric not null,
    withdrawal_status int not null,
    withdrawal_channel int not null,
    withdrawal_via_pg boolean not null,
    withdrawal_pg_vendor varchar,
    created_at timestamptz not null,
    updated_at timestamptz,
    finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS wallet_bank_withdrawals (
    transaction_id varchar not null,
    bank_name varchar not null,
    created_at timestamptz not null
);

CREATE TABLE IF NOT EXISTS wallet_ewallet_withdrawals (
    transaction_id varchar not null,
    ewallet_name varchar not null,
    created_at timestamptz not null
);

CREATE TABLE IF NOT EXISTS wallet_reversal (
    transaction_id varchar not null,
    reversed_transaction_id varchar not null,
    created_at timestamptz not null
);

CREATE TABLE IF NOT EXISTS wallet_chargebacks (
    transaction_id varchar not null,
    chargeback_type int not null,
    amount numeric not null,
    reason text not null,
    created_at timestamptz not null
);
