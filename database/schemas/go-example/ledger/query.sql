-- name: CreateAccount :exec
INSERT INTO ledger.accounts(
	account_id,
	name,
	description,
	parent_account_id,
	currency_id,
	created_at
) VALUES($1,$2,$3,$4,$5,$6);

-- name: CreateAccountBalance :exec
INSERT INTO ledger.accounts_balance(
	account_id,
	parent_account_id,
	allow_negative,
	balance,
	last_ledger_id,
	last_movement_id,
	currency_id,
	created_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8);

-- name: GetAccounts :many
SELECT *
FROM ledger.accounts
WHERE account_id = ANY($1::varchar[])
ORDER BY created_at;

-- name: GetAccountsBalance :many
SELECT ab.account_id,
    ab.parent_account_id,
	ab.allow_negative,
	ab.balance,
	ab.currency_id,
	ab.last_ledger_id,
	ab.last_movement_id,
	ab.created_at,
	ab.updated_at
FROM ledger.accounts_balance ab,
	ledger.accounts ac
WHERE ab.account_id = ANY($1::varchar[])
	AND ab.account_id = ac.account_id;

-- name: GetMovementByIdempotencyKey :one
SELECT movement_id,
    idempotency_key,
    created_at,
    updated_at
FROM ledger.movements
WHERE idempotency_key = $1;

-- name: CreateMovement :exec
INSERT INTO ledger.movements(
	movement_id,
	idempotency_key,
	created_at,
	updated_at
) VALUES($1,$2,$3,$4);

-- name: GetMovement :one
SELECT * FROM ledger.movements
WHERE movement_id = $1;

-- name: GetAccountsLedgerByMovementID :many
SELECT ledger_id,
	movement_id,
	movement_sequence,
	account_id,
	amount,
	previous_ledger_id,
	client_id,
	created_at,
	client_id
FROM ledger.accounts_ledger
WHERE movement_id = $1
ORDER BY created_at;

-- name: GetAccountsBalanceWithChild :one
WITH sum_main AS (
    SELECT account_id,
        allow_negative,
        balance,
        last_ledger_id,
        last_movement_id,
        currency_id,
        created_at
    FROM ledger.accounts_balance
    WHERE account_id = $1
),
child_accounts AS (
    SELECT parent_account_id as account_id,
        SUM(balance) as balance
    FROM ledger.accounts_balance
    WHERE parent_account_id = $1
    GROUP BY parent_account_id
)
SELECT
    main_acc.account_id,
    main_acc.allow_negative,
    main_acc.balance + child_acc.balance total_account_balance,
    main_acc.balance main_account_balance,
    child_acc.balance child_account_balance,
    main_acc.last_ledger_id,
    main_acc.last_movement_id,
    main_acc.currency_id,
    main_acc.created_at
FROM sum_main main_acc,
    child_accounts child_acc
WHERE main_acc.account_id = child_acc.account_id;

-- name: GetAccountsBalancesWithChild :many
WITH sum_main AS (
    SELECT account_id,
        allow_negative,
        balance,
        last_ledger_id,
        last_movement_id,
        currency_id,
        created_at
    FROM ledger.accounts_balance
    WHERE account_id = ANY($1::varchar[])
),
child_accounts AS (
    SELECT parent_account_id as account_id,
        SUM(balance) as balance
    FROM ledger.accounts_balance
    WHERE parent_account_id = ANY($1::varchar[])
    GROUP BY parent_account_id
)
SELECT
    main_acc.account_id,
    main_acc.allow_negative,
    main_acc.balance + child_acc.balance total_account_balance,
    main_acc.balance main_account_balance,
    child_acc.balance child_account_balance,
    main_acc.last_ledger_id,
    main_acc.last_movement_id,
    main_acc.currency_id,
    main_acc.created_at
FROM sum_main main_acc,
    child_accounts child_acc
WHERE main_acc.account_id = child_acc.account_id;

-- name: GetAccountsBalancesWithChildForUpdate :many
WITH sum_main AS (
    SELECT account_id,
        allow_negative,
        balance,
        last_ledger_id,
        last_movement_id,
        currency_id,
        created_at
    FROM ledger.accounts_balance
    WHERE account_id = ANY($1::varchar[])
),
child_accounts AS (
    SELECT parent_account_id as account_id,
        SUM(balance) as balance
    FROM ledger.accounts_balance
    WHERE parent_account_id = ANY($1::varchar[])
    GROUP BY parent_account_id
)
SELECT
    main_acc.account_id,
    main_acc.allow_negative,
    main_acc.balance + child_acc.balance total_account_balance,
    main_acc.balance main_account_balance,
    child_acc.balance child_account_balance,
    main_acc.last_ledger_id,
    main_acc.last_movement_id,
    main_acc.currency_id,
    main_acc.created_at
FROM sum_main main_acc,
    child_accounts child_acc
WHERE main_acc.account_id = child_acc.account_id
FOR UPDATE;
