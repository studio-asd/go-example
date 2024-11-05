-- drop tables.
DROP TABLE IF EXISTS movements;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS accounts_balance;
DROP TABLE IF EXISTS accounts_ledger;

-- types.

DROP TYPE IF EXISTS account_status CASCADE;
CREATE TYPE account_status AS ENUM('active', 'inactive');

-- tables and index.

-- accounts is used to store all user accounts.
CREATE TABLE IF NOT EXISTS accounts(
	"account_id" varchar PRIMARY KEY,
	"parent_account_id" varchar NOT NULL,
	"account_status" account_status NOT NULL,
	"currency_id" int NOT NULL,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp
);

-- movements is used to store all movement records.
CREATE TABLE IF NOT EXISTS movements(
	"movement_id" varchar PRIMARY KEY,
    "idempotency_key" varchar UNIQUE NOT NULL,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp
);

-- accounts_balance is used to store the latest state of user's balance. This table will be used for user
-- balance fast retrieval and for locking the user balance for movement.
CREATE TABLE IF NOT EXISTS accounts_balance(
	"account_id" varchar PRIMARY KEY,
	"currency_id" int NOT NULL,
	-- allow_negative allows some accounts to have negative balance. For example, for the funding
	-- account we might allow the account to have negative balance.
	"allow_negative" boolean NOT NULL,
	"balance" numeric NOT NULL,
	"last_movement_id" varchar NOT NULL,
	"last_ledger_id" varchar NOT NULL,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp
);

-- accounts_balance_history is a historical balance changes for an account based on movement_id. The historical
-- balance is per movement_id and not per ledger_id because we pre-calculates everything inside the system. And because
-- we are building the ledger optimistically, the historical amount can changed by the time we calculate because
-- we are not locking the balance up-front(as this will be expensive). So it makes more sense to create the history
-- based on movement_id because we will do that in bulk rather than ledger_id.
--
-- This table can be used for various things like retrieving opening and ending balance of an account at a given time.
CREATE TABLE IF NOT EXISTS accounts_balance_history(
    "movement_id" varchar PRIMARY KEY,
    -- ledger_id is the id of where the balance is being summarized, the SUM of the balance in the accounts_ledger should be
    -- the same if we sumarize everything up to this ledger_id.
    "ledger_id" varchar NOT NULL,
    "account_id" varchar NOT NULL,
    "balance" numeric NOT NULL,
    "previous_balance" numeric NOT NULL,
    "previous_movement_id" varchar NOT NULL,
    "previous_ledger_id" varchar NOT NULL,
    "created_at" timestamp NOT NULL
);

-- accounts_ledger is used to store all ledger changes for a specific account. A single transaction
-- can possibly affecting multiple acounts in the ledger.
--
-- Row in this table is IMMUTABLE and should NOT be updated.
CREATE TABLE IF NOT EXISTS accounts_ledger(
    -- the ledger id is the primary key of the accounts_ledger. Even though we have unique constraint, but the client
    -- can always refer themselves to this unique id when it comes to reconciliation.
	--
	-- Why varchar? Because its hard to pre-determine the id if it is a generated_id by the PostgreSQL. As some records
	-- will be pre-generated inside the program.
    "ledger_id" varchar PRIMARY KEY,
	"movement_id" varchar NOT NULL,
	"account_id" varchar NOT NULL,
	-- movement_seuqnce is the ordered sequence of movement inside a movement_id.
	"movement_sequence" int NOT NULL,
	"currency_id" int NOT NULL,
	"amount" numeric NOT NULL,
	-- previous_ledger_id will be used to track the sequence of the ledger entries of a user.
	"previous_ledger_id" varchar NOT NULL,
	"created_at" timestamp NOT NULL,
	-- client_id is an identifier that the client can use in case they want to link their ids to per-ledger-row. With this, there are
	-- many cases they can use with the ledger.
	--
	-- For example, the client want to use the ledger for transfer. The client might want to have a separate transfer table that have its own
	-- id, use that id when creating the transaction to the ledger.
	"client_id" varchar
);

-- accounts_ledger index will be used for several cases:
-- 1. We want to retrieve all transactions within a movement_id. Possibly sorted by timestamp.
-- 2. We want to retrieve all transactions within an account_id. Possibly sorted by timestamp.
DROP INDEX IF EXISTS idx_accounts_ledger_account_id;
DROP INDEX IF EXISTS idx_accounts_ledger_movement_id;
CREATE INDEX IF NOT EXISTS idx_accounts_ledger_movement_id ON accounts_ledger("movement_id");
CREATE INDEX IF NOT EXISTS idx_accounts_ledger_account_id ON accounts_ledger("account_id");
