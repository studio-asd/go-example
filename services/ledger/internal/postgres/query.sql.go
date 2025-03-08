// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

const createAccount = `-- name: CreateAccount :exec
INSERT INTO ledger.accounts(
	account_id,
	parent_account_id,
	currency_id,
	created_at
) VALUES($1,$2,$3,$4)
`

type CreateAccountParams struct {
	AccountID       string
	ParentAccountID string
	CurrencyID      int32
	CreatedAt       time.Time
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) error {
	_, err := q.db.Exec(ctx, createAccount,
		arg.AccountID,
		arg.ParentAccountID,
		arg.CurrencyID,
		arg.CreatedAt,
	)
	return err
}

const createAccountBalance = `-- name: CreateAccountBalance :exec
INSERT INTO ledger.accounts_balance(
	account_id,
	parent_account_id,
	allow_negative,
	balance,
	last_ledger_id,
	last_movement_id,
	currency_id,
	created_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8)
`

type CreateAccountBalanceParams struct {
	AccountID       string
	ParentAccountID sql.NullString
	AllowNegative   bool
	Balance         decimal.Decimal
	LastLedgerID    string
	LastMovementID  string
	CurrencyID      int32
	CreatedAt       time.Time
}

func (q *Queries) CreateAccountBalance(ctx context.Context, arg CreateAccountBalanceParams) error {
	_, err := q.db.Exec(ctx, createAccountBalance,
		arg.AccountID,
		arg.ParentAccountID,
		arg.AllowNegative,
		arg.Balance,
		arg.LastLedgerID,
		arg.LastMovementID,
		arg.CurrencyID,
		arg.CreatedAt,
	)
	return err
}

const createMovement = `-- name: CreateMovement :exec
INSERT INTO ledger.movements(
	movement_id,
	idempotency_key,
	created_at,
	updated_at
) VALUES($1,$2,$3,$4)
`

type CreateMovementParams struct {
	MovementID     string
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      sql.NullTime
}

func (q *Queries) CreateMovement(ctx context.Context, arg CreateMovementParams) error {
	_, err := q.db.Exec(ctx, createMovement,
		arg.MovementID,
		arg.IdempotencyKey,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getAccounts = `-- name: GetAccounts :many
SELECT account_id, parent_account_id, currency_id, created_at, updated_at
FROM ledger.accounts
WHERE account_id = ANY($1::varchar[])
ORDER BY created_at
`

func (q *Queries) GetAccounts(ctx context.Context, dollar_1 []string) ([]LedgerAccount, error) {
	rows, err := q.db.Query(ctx, getAccounts, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LedgerAccount
	for rows.Next() {
		var i LedgerAccount
		if err := rows.Scan(
			&i.AccountID,
			&i.ParentAccountID,
			&i.CurrencyID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAccountsBalance = `-- name: GetAccountsBalance :many
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
	AND ab.account_id = ac.account_id
`

type GetAccountsBalanceRow struct {
	AccountID       string
	ParentAccountID sql.NullString
	AllowNegative   bool
	Balance         decimal.Decimal
	CurrencyID      int32
	LastLedgerID    string
	LastMovementID  string
	CreatedAt       time.Time
	UpdatedAt       sql.NullTime
}

func (q *Queries) GetAccountsBalance(ctx context.Context, dollar_1 []string) ([]GetAccountsBalanceRow, error) {
	rows, err := q.db.Query(ctx, getAccountsBalance, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAccountsBalanceRow
	for rows.Next() {
		var i GetAccountsBalanceRow
		if err := rows.Scan(
			&i.AccountID,
			&i.ParentAccountID,
			&i.AllowNegative,
			&i.Balance,
			&i.CurrencyID,
			&i.LastLedgerID,
			&i.LastMovementID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAccountsBalanceWithChild = `-- name: GetAccountsBalanceWithChild :one
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
WHERE main_acc.account_id = child_acc.account_id
`

type GetAccountsBalanceWithChildRow struct {
	AccountID           string
	AllowNegative       bool
	TotalAccountBalance decimal.Decimal
	MainAccountBalance  decimal.Decimal
	ChildAccountBalance decimal.Decimal
	LastLedgerID        string
	LastMovementID      string
	CurrencyID          int32
	CreatedAt           time.Time
}

func (q *Queries) GetAccountsBalanceWithChild(ctx context.Context, dollar_1 sql.NullString) (GetAccountsBalanceWithChildRow, error) {
	row := q.db.QueryRow(ctx, getAccountsBalanceWithChild, dollar_1)
	var i GetAccountsBalanceWithChildRow
	err := row.Scan(
		&i.AccountID,
		&i.AllowNegative,
		&i.TotalAccountBalance,
		&i.MainAccountBalance,
		&i.ChildAccountBalance,
		&i.LastLedgerID,
		&i.LastMovementID,
		&i.CurrencyID,
		&i.CreatedAt,
	)
	return i, err
}

const getAccountsBalancesWithChild = `-- name: GetAccountsBalancesWithChild :many
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
`

type GetAccountsBalancesWithChildRow struct {
	AccountID           string
	AllowNegative       bool
	TotalAccountBalance decimal.Decimal
	MainAccountBalance  decimal.Decimal
	ChildAccountBalance decimal.Decimal
	LastLedgerID        string
	LastMovementID      string
	CurrencyID          int32
	CreatedAt           time.Time
}

func (q *Queries) GetAccountsBalancesWithChild(ctx context.Context, dollar_1 []string) ([]GetAccountsBalancesWithChildRow, error) {
	rows, err := q.db.Query(ctx, getAccountsBalancesWithChild, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAccountsBalancesWithChildRow
	for rows.Next() {
		var i GetAccountsBalancesWithChildRow
		if err := rows.Scan(
			&i.AccountID,
			&i.AllowNegative,
			&i.TotalAccountBalance,
			&i.MainAccountBalance,
			&i.ChildAccountBalance,
			&i.LastLedgerID,
			&i.LastMovementID,
			&i.CurrencyID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAccountsBalancesWithChildForUpdate = `-- name: GetAccountsBalancesWithChildForUpdate :many
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
FOR UPDATE
`

type GetAccountsBalancesWithChildForUpdateRow struct {
	AccountID           string
	AllowNegative       bool
	TotalAccountBalance decimal.Decimal
	MainAccountBalance  decimal.Decimal
	ChildAccountBalance decimal.Decimal
	LastLedgerID        string
	LastMovementID      string
	CurrencyID          int32
	CreatedAt           time.Time
}

func (q *Queries) GetAccountsBalancesWithChildForUpdate(ctx context.Context, dollar_1 []string) ([]GetAccountsBalancesWithChildForUpdateRow, error) {
	rows, err := q.db.Query(ctx, getAccountsBalancesWithChildForUpdate, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAccountsBalancesWithChildForUpdateRow
	for rows.Next() {
		var i GetAccountsBalancesWithChildForUpdateRow
		if err := rows.Scan(
			&i.AccountID,
			&i.AllowNegative,
			&i.TotalAccountBalance,
			&i.MainAccountBalance,
			&i.ChildAccountBalance,
			&i.LastLedgerID,
			&i.LastMovementID,
			&i.CurrencyID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAccountsLedgerByMovementID = `-- name: GetAccountsLedgerByMovementID :many
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
ORDER BY created_at
`

type GetAccountsLedgerByMovementIDRow struct {
	LedgerID         string
	MovementID       string
	MovementSequence int32
	AccountID        string
	Amount           decimal.Decimal
	PreviousLedgerID string
	ClientID         sql.NullString
	CreatedAt        time.Time
	ClientID_2       sql.NullString
}

func (q *Queries) GetAccountsLedgerByMovementID(ctx context.Context, movementID string) ([]GetAccountsLedgerByMovementIDRow, error) {
	rows, err := q.db.Query(ctx, getAccountsLedgerByMovementID, movementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAccountsLedgerByMovementIDRow
	for rows.Next() {
		var i GetAccountsLedgerByMovementIDRow
		if err := rows.Scan(
			&i.LedgerID,
			&i.MovementID,
			&i.MovementSequence,
			&i.AccountID,
			&i.Amount,
			&i.PreviousLedgerID,
			&i.ClientID,
			&i.CreatedAt,
			&i.ClientID_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMovement = `-- name: GetMovement :one
SELECT movement_id, idempotency_key, created_at, updated_at, reversed_at, reversal_movement_id FROM ledger.movements
WHERE movement_id = $1
`

func (q *Queries) GetMovement(ctx context.Context, movementID string) (LedgerMovement, error) {
	row := q.db.QueryRow(ctx, getMovement, movementID)
	var i LedgerMovement
	err := row.Scan(
		&i.MovementID,
		&i.IdempotencyKey,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ReversedAt,
		&i.ReversalMovementID,
	)
	return i, err
}

const getMovementByIdempotencyKey = `-- name: GetMovementByIdempotencyKey :one
SELECT movement_id,
    idempotency_key,
    created_at,
    updated_at
FROM ledger.movements
WHERE idempotency_key = $1
`

type GetMovementByIdempotencyKeyRow struct {
	MovementID     string
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      sql.NullTime
}

func (q *Queries) GetMovementByIdempotencyKey(ctx context.Context, idempotencyKey string) (GetMovementByIdempotencyKeyRow, error) {
	row := q.db.QueryRow(ctx, getMovementByIdempotencyKey, idempotencyKey)
	var i GetMovementByIdempotencyKeyRow
	err := row.Scan(
		&i.MovementID,
		&i.IdempotencyKey,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
