-- drop tables.
DROP TABLE IF EXISTS movements;

DROP TABLE IF EXISTS accounts;

DROP TABLE IF EXISTS accounts_balance;

DROP TABLE IF EXISTS accounts_ledger;

-- types.
DROP TYPE IF EXISTS account_status CASCADE;

CREATE TYPE account_status AS ENUM ('active', 'inactive');

DROP TYPE IF EXISTS movement_status CASCADE;

CREATE TYPE movement_status AS ENUM ('finished', 'reversed');

-- tables and index.
-- accounts is used to store all user accounts.
CREATE TABLE IF NOT EXISTS accounts (
    "account_id" varchar PRIMARY KEY,
    "parent_account_id" varchar NOT NULL,
    "account_status" account_status NOT NULL,
    "currency_id" int NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz
);

-- movements is used to store all movement records.
CREATE TABLE IF NOT EXISTS movements (
    "movement_id" varchar PRIMARY KEY,
    "idempotency_key" varchar UNIQUE NOT NULL,
    "movement_status" movement_status NOT NULL,
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

-- accounts_ledger index will be used for several cases:
-- 1. We want to retrieve all transactions within a movement_id. Possibly sorted by timestamp.
-- 2. We want to retrieve all transactions within an account_id. Possibly sorted by timestamp.
DROP INDEX IF EXISTS idx_accounts_ledger_movement_id;

DROP INDEX IF EXISTS idx_accounts_ledger_account_id;

DROP INDEX IF EXISTS idx_accounts_ledger_client_id;

DROP INDEX IF EXISTS idx_accounts_balance_parent_account_id;

CREATE INDEX IF NOT EXISTS idx_accounts_balance_parent_account_id ON accounts_balance ("parent_account_id");

CREATE INDEX IF NOT EXISTS idx_accounts_ledger_movement_id ON accounts_ledger ("movement_id");

CREATE INDEX IF NOT EXISTS idx_accounts_ledger_account_id ON accounts_ledger ("account_id");

CREATE INDEX IF NOT EXISTS idx_accounts_ledger_client_id ON accounts_ledger ("client_id")
WHERE
    "client_id" IS NOT NULL;
