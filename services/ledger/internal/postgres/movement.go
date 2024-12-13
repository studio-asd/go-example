package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/albertwidi/pkg/postgres"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/services/ledger"
	internal "github.com/albertwidi/go-example/services/ledger/internal"
)

// bulkUpdate is a custom type to pass the bulk update parameters around functions. The type is customized
// based on the existing parameters inside of the Postgres package.
type bulkUpdate struct {
	Table   string
	Columns []string
	Types   []string
	Values  [][]any
}

// Move moves balance from one account to another based on the movement entries.
//
// The movement can be vary, depends on the entries being given.
// 1. From one account to another account. One to one.
// 2. From one account to another accounts. One to many.
// 3. From accounts to accounts. Many to many.
//
// As long as the SUM total amount in the entries is 0, then it is a valid movement.
func (q *Queries) Move(ctx context.Context, le ledger.MovementLedgerEntries) (internal.MovementResult, error) {
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
	}
	accountsBalanceHistoryColumns := []string{
		"movement_id",
		"ledger_id",
		"account_id",
		"balance",
		"previous_balance",
		"previous_movement_id",
		"previous_ledger_id",
		"created_at",
	}
	var result internal.MovementResult

	// Below here, we will do everything inside the database transaction. So its important to keep in mind, whatever we are doing here it gotta have to
	// be fast. As we are locking the users balances here, the user won't be able to create new movement if the account is still being locked in the
	// database transaction.
	err := q.ensureInTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *Queries) error {
		bulkUpdateParams, endingBalances, err := selectAccountsBalanceForMovement(ctx, q, le.AccountsSummary, le.CreatedAt, le.Accounts)
		if err != nil {
			return err
		}
		// Create the bulk insert parameters for accounts_balance_history. This is imporatnt as we want to record the histories
		// of the balance based on the movement and not per ledger-record basis.
		bulkInsertBalanceHistoryParams := make([]any, len(accountsBalanceHistoryColumns)*len(endingBalances))
		counter := 0
		for accID, info := range endingBalances {
			beginIndex := len(accountsBalanceHistoryColumns) * counter
			bulkInsertBalanceHistoryParams[beginIndex] = le.MovementID
			bulkInsertBalanceHistoryParams[beginIndex+1] = info.NextLedgerID
			bulkInsertBalanceHistoryParams[beginIndex+2] = accID
			bulkInsertBalanceHistoryParams[beginIndex+3] = info.NewBalance
			bulkInsertBalanceHistoryParams[beginIndex+4] = info.PreviousBalance
			bulkInsertBalanceHistoryParams[beginIndex+5] = info.PreviousMovementID
			bulkInsertBalanceHistoryParams[beginIndex+6] = info.PreviousLedgerID
			bulkInsertBalanceHistoryParams[beginIndex+7] = le.CreatedAt
			counter++
		}

		// Build the bulk insert parameters for ledger. We are doing the bulk insert as we need to insert the ledger for both DEBIT and CREDIT for each
		// movement. The length of the params is (columns * entries) as the parameters are concattenated in a single array.
		bulkInsertLedgerParams := make([]any, len(accountsLedgerColumns)*len(le.LedgerEntries))
		for idx, entry := range le.LedgerEntries {
			// There will be a condition of when the previous ledger_id is empty, because we need to lock the balance row first to get the
			// exact previous identifier. This will always be the case for the first entry of an account movement. When this happen, then we will fill the entry
			// data with the information of when the lock/SELECT FOR UPDATE happened.
			if entry.PreviousLedgerID == "" {
				entry.PreviousLedgerID = endingBalances[entry.AccountID].NextLedgerID
			}
			// Set the client id if the client_id is not null.
			clientID := sql.Null[string]{}
			if entry.ClientID != "" {
				clientID.V = entry.ClientID
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
		}

		// Update the affected users balance. We updated the balance first because it will affects less row than inserting the records to ledger.
		// So if something bad happens, it will be more effecicient to rollback.
		if err := q.db.BulkUpdate(
			ctx,
			bulkUpdateParams.Table,
			bulkUpdateParams.Columns,
			bulkUpdateParams.Types,
			bulkUpdateParams.Values,
		); err != nil {
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
			return fmt.Errorf("failed to insert accounts balance history: %w", err)
		}
		// Insert to the movement.
		if err := q.CreateMovement(ctx, CreateMovementParams{
			MovementID:     le.MovementID,
			IdempotencyKey: le.IdempotencyKey,
			MovementStatus: MovementStatusFinished,
			CreatedAt:      le.CreatedAt,
		}); err != nil {
			fmt.Println(le.MovementID)
			return fmt.Errorf("failed to create movement: %w", err)
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
		// Provide the result information.
		result = internal.MovementResult{
			MovementID: le.MovementID,
			Time:       le.CreatedAt,
			Balances:   endingBalances,
		}
		return nil
	})
	return result, err
}

// selectAccountsBalanceForMovement do SELECT FOR UPDATE to the account_balances and lock specific account_id balance. The function also returns the update statements
// for all accounts so we can also tests whether the update statement is contstructed as we expected or not.
func selectAccountsBalanceForMovement(ctx context.Context, q *Queries, changes map[string]ledger.AccountMovementSummary, createdAt time.Time, accounts []string) (bulkUpdate, map[string]internal.MovementEndingBalance, error) {
	if len(accounts) == 0 {
		return bulkUpdate{}, nil, errors.New("account_id is required to select accounts balance for movement")
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
		return bulkUpdate{}, nil, err
	}
	// Construct the bulk update parameters so we can execute the bulk update in a single transaction.
	bulkUpdate := bulkUpdate{
		Table: "accounts_balance",
		Columns: []string{
			"account_id",
			"balance",
			"last_ledger_id",
			"updated_at",
		},
		Types: []string{
			"VARCHAR",
			"NUMERIC",
			"VARCHAR",
			"TIMESTAMP",
		},
	}
	// The values first array must have the same length of the columns, so we will create it with columns
	// as the base length.
	bulkUpdate.Values = make([][]any, len(bulkUpdate.Columns))

	endingBalances := make(map[string]internal.MovementEndingBalance)
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
		endingBalances[ab.AccountID] = internal.MovementEndingBalance{
			AccountID:          ab.AccountID,
			PreviousBalance:    ab.Balance,
			PreviousLedgerID:   ab.LastLedgerID,
			PreviousMovementID: ab.LastMovementID,
			NewBalance:         newBalance,
			NextLedgerID:       changes[ab.AccountID].NextLedgerID,
		}
		// Append the values to the bulk update parameters as we want to updates all the balances at the same time.
		bulkUpdate.Values[0] = append(bulkUpdate.Values[0], ab.AccountID)
		bulkUpdate.Values[1] = append(bulkUpdate.Values[1], newBalance.String())
		bulkUpdate.Values[2] = append(bulkUpdate.Values[2], changes[ab.AccountID].NextLedgerID)
		bulkUpdate.Values[3] = append(bulkUpdate.Values[3], createdAt)
		return nil
	}, selectForUpdateArgs...)
	if err != nil {
		return bulkUpdate, nil, fmt.Errorf("failed to select for update: %w", err)
	}
	return bulkUpdate, endingBalances, nil
}
