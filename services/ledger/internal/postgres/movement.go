package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/albertwidi/pkg/postgres"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/services/ledger"
)

// Move moves balance from one account to another based on the movement entries.
//
// The movement can be vary, depends on the entries being given.
// 1. From one account to another account. One to one.
// 2. From one account to another accounts. One to many.
// 3. From accounts to accounts. Many to many.
//
// As long as the SUM total amount in the entries is 0, then it is a valid movement.
func (q *Queries) Move(ctx context.Context, le ledger.MovementLedgerEntries) error {
	accountsLedgerColumns := []string{
		"ledger_id",
		"movement_id",
		"account_id",
		"movement_sequence",
		"currency_id",
		"amount",
		"previous_ledger_id",
		"client_id",
		"created_at",
		"timestamp",
	}
	accountsBalanceHistoryColumns := []string{
		"movement_id",
		"account_id",
		"balance",
		"previous_balance",
		"previous_movement_id",
		"created_at",
	}

	// Below here, we will do everything inside the database transaction. So its important to keep in mind, whatever we are doing here it gotta have to
	// be fast. As we are locking the users balances here, the user won't be able to create new movement if the account is still being locked in the
	// database transaction.
	err := q.WithTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *Queries) error {
		updateBalanceQuery, lastBalancesInfo, err := selectAccountsBalanceForMovement(ctx, q, le.AccountsSummary, le.CreatedAt, le.Accounts)
		if err != nil {
			return err
		}
		// Create the bulk insert parameters for accounts_balance_history. This is imporatnt as we want to record the histories
		// of the balance based on the movement and not per ledger-record basis.
		bulkInsertBalanceHistoryParams := make([]any, len(accountsBalanceHistoryColumns)*len(accountsBalanceHistoryColumns))
		counter := 0
		for accID, info := range lastBalancesInfo {
			beginIndex := len(accountsBalanceHistoryColumns) * counter
			bulkInsertBalanceHistoryParams[beginIndex] = le.MovementID
			bulkInsertBalanceHistoryParams[beginIndex+1] = accID
			bulkInsertBalanceHistoryParams[beginIndex+2] = info.NewBalance
			bulkInsertBalanceHistoryParams[beginIndex+3] = info.PreviousBalance
			bulkInsertBalanceHistoryParams[beginIndex+4] = info.PreviousMovementID
			bulkInsertBalanceHistoryParams[beginIndex+5] = le.CreatedAt
		}

		// Build the bulk insert parameters for ledger. We are doing the bulk insert as we need to insert the ledger for both DEBIT and CREDIT for each
		// movement. The length of the params is (columns * entries) as the parameters are concattenated in a single array.
		bulkInsertLedgerParams := make([]any, len(accountsLedgerColumns)*len(le.LedgerEntries))
		for idx, entry := range le.LedgerEntries {
			// There will be a condition of when the previous ledger_id is empty, because we need to lock the balance row first to get the
			// exact previous identifier. This will always be the case for the first entry of an account movement. When this happen, then we will fill the entry
			// data with the information of when the lock/SELECT FOR UPDATE happened.
			if entry.PreviousLedgerID == "" {
				entry.PreviousLedgerID = lastBalancesInfo[entry.AccountID].LastLedgerID
			}
			// Set the client id if the client_id is not null.
			clientID := sql.NullString{}
			if entry.ClientID != "" {
				clientID.String = entry.ClientID
				clientID.Valid = true
			}
			// The beginning of the offset is always (idx * len(accountsLedgerColumns)) because the parameters are concattenated based on the columns length.
			// If the number of columns are increased/decreased, the (offset + x) need to be modified based on the number of the columns/fields.
			offset := idx * len(accountsLedgerColumns)
			bulkInsertLedgerParams[offset] = entry.LedgerID
			bulkInsertLedgerParams[offset+1] = le.MovementID
			bulkInsertLedgerParams[offset+2] = entry.AccountID
			bulkInsertLedgerParams[offset+3] = entry.MovementSequence
			bulkInsertLedgerParams[offset+4] = entry.CurrencyID
			bulkInsertLedgerParams[offset+5] = entry.Amount.String()
			bulkInsertLedgerParams[offset+6] = entry.PreviousLedgerID
			bulkInsertLedgerParams[offset+7] = clientID
			bulkInsertLedgerParams[offset+8] = entry.CreatedAt
			bulkInsertLedgerParams[offset+9] = entry.Timestamp
		}

		// Update the affected users balance. We updated the balance first because it will affects less row than inserting the records to ledger.
		// So if something bad happens, it will be more effecicient to rollback.
		_, err = q.db.Exec(ctx, updateBalanceQuery)
		if err != nil {
			slog.Debug("faield to update balance", slog.String("query", updateBalanceQuery))
			return fmt.Errorf("failed to update balances: %w", err)
		}
		// Insert to the accounts balance history.
		if err := q.db.BulkInsert(
			ctx,
			"accounts_balance_history",
			accountsBalanceHistoryColumns,
			bulkInsertBalanceHistoryParams,
			"",
		); err != nil {
			fmt.Errorf("failed to insert accounts balance history: %w", err)
		}
		// Insert to the accounts ledger.
		if err := q.db.BulkInsert(
			ctx,
			"accounts_ledger",
			accountsLedgerColumns,
			bulkInsertLedgerParams,
			"",
		); err != nil {
			return fmt.Errorf("failed to insert accounts ledger: %w", err)
		}
		return nil
	})
	return err
}

// AccountLastBalanceInfo is the latest information from the account balance for a specific account_id balance.
// The information is needed when moving money from one account to another as we need to know the latest
// state of the account before moving its balance.
type AccountLastBalanceInfo struct {
	AccountID          string
	PreviousBalance    decimal.Decimal
	PreviousLedgerID   string
	PreviousMovementID string
	NewBalance         decimal.Decimal
	LastLedgerID       string
}

// selectAccountsBalanceForMovement do SELECT FOR UPDATE to the account_balances and lock specific account_id balance. The function also returns the update statements
// for all accounts so we can also tests whether the update statement is contstructed as we expected or not.
func selectAccountsBalanceForMovement(ctx context.Context, q *Queries, changes map[string]ledger.AccountMovementSummary, createdAt time.Time, accounts []string) (string, map[string]AccountLastBalanceInfo, error) {
	if len(accounts) == 0 {
		return "", nil, errors.New("account_id is required to select accounts balance for movement")
	}

	selectForUpdate, selectForUpdateArgs, err := squirrel.Select(
		"account_id",
		"allow_negative",
		"balance",
		"last_ledger_id",
		"last_movement_id",
	).
		From("accounts_balance").
		Where(squirrel.Eq{"account_id": accounts}).
		Suffix("FOR UPDATE").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return "", nil, err
	}

	// updateBalanceQuery updates balances of all accounts based on the update information foer each account.
	// The (VALUES %s) will be translated to (balance, movement_id, movement_sequence, account_id, updated_at) of each account.
	// For example: (VALUES (100, '1', 1, '1', '2024-01-01'), (200, '2', 1, '2', '2024-01-01')).
	var updateBalanceValues []string
	updateBalanceQuery := `
UPDATE accounts_balance AS ab SET
	balance = v.balance,
	last_ledger_id = v.last_ledger_id,
	updated_at = v.updated_at
FROM (VALUES %s) AS v(balance, last_ledger_id, account_id, updated_at)
WHERE ab.account_id = v.account_id;
`

	accountsLastInfo := make(map[string]AccountLastBalanceInfo)
	// accounts is the new information of the accounts which we want to change. We will use this to create UPDATE query.
	err = q.db.RunQuery(ctx, selectForUpdate, func(rc *postgres.RowsCompat) error {
		ab := AccountsBalance{}
		if err := rc.Scan(
			&ab.AccountID,
			&ab.AllowNegative,
			&ab.Balance,
			&ab.LastLedgerID,
			&ab.LastMovementID,
		); err != nil {
			return err
		}

		var newBalance decimal.Decimal
		// If the last ledger id is still the same under lock, then we should use the ending balance given, as there are no
		// transactions being recorded concurrently for this account.
		if changes[ab.AccountID].LastLedgerID == ab.LastLedgerID {
			newBalance = changes[ab.AccountID].EndingBalance
		} else {
			// Check if the account have enough balance to be deducted and whether negative balance is allowed for the account.
			newBalance = ab.Balance.Add(changes[ab.AccountID].BalanceChanges)
			if newBalance.IsNegative() && !ab.AllowNegative {
				return ledger.ErrInsufficientBalance
			}
		}
		// Store the accounts information before we change it.
		accountsLastInfo[ab.AccountID] = AccountLastBalanceInfo{
			AccountID:          ab.AccountID,
			PreviousBalance:    ab.Balance,
			PreviousLedgerID:   ab.LastLedgerID,
			PreviousMovementID: ab.LastMovementID,
			NewBalance:         newBalance,
			LastLedgerID:       changes[ab.AccountID].NextLedgerID,
		}

		// Create the VALUES of update balance query. For example:
		// VALUES("100", "movement_id", "1", "account_id", "2024-01-01"),
		// 	("200", "movement_id", "1", "account_id2", "2024-01-01")
		updateBalanceValue := "(" + strings.Join([]string{
			newBalance.String(),
			"'" + changes[ab.AccountID].NextLedgerID + "'",
			"'" + ab.AccountID + "'",
			"to_timestamp('" + createdAt.Format(time.DateTime) + "', 'yyyy-mm-dd hh24:mi:ss')",
		}, ",") + ")"
		updateBalanceValues = append(updateBalanceValues, updateBalanceValue)
		return nil
	}, selectForUpdateArgs...)
	if err != nil {
		return "", nil, fmt.Errorf("failed to select for update: %w", err)
	}
	return fmt.Sprintf(updateBalanceQuery, strings.Join(updateBalanceValues, ",")), accountsLastInfo, nil
}
