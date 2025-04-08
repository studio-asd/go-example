package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"

	"github.com/studio-asd/go-example/internal/currency"
	"github.com/studio-asd/go-example/services/ledger"
	internal "github.com/studio-asd/go-example/services/ledger/internal"
)

// TestMove tests whether the database records are correct when movement happens.
func TestMove(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().UTC()
	timestamp := createdAt.Unix()

	// accountsSetup is a set of accounts for test setup/fixtures.
	// For easiness of the test, all accounts will have 100 for their
	// initial balance setup
	accountsSetup := map[string]decimal.Decimal{
		"1": decimal.NewFromInt(100),
		"2": decimal.NewFromInt(100),
	}

	tests := []struct {
		name                  string
		entries               ledger.MovementLedgerEntries
		expectMovementResult  internal.MovementResult
		expectMovement        Movement
		expectAccountsBalance map[string]GetAccountsBalanceRow
		expectAccountsLedger  []GetAccountsLedgerByMovementIDRow
	}{
		{
			name: "simple entry",
			entries: ledger.MovementLedgerEntries{
				MovementID:     "one",
				IdempotencyKey: "one",
				LedgerEntries: []ledger.LedgerEntry{
					{
						AccountID:        "1",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(-100),
						LedgerID:         "one",
						CreatedAt:        createdAt,
						Timestamp:        timestamp + 1,
					},
					{
						AccountID:        "2",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(100),
						LedgerID:         "two",
						CreatedAt:        createdAt,
						Timestamp:        timestamp + 2,
					},
				},
				AccountsSummary: map[string]ledger.AccountMovementSummary{
					"1": {
						BalanceChanges: decimal.NewFromInt(-100),
						NextLedgerID:   "one",
						EndingBalance:  decimal.Zero,
					},
					"2": {
						BalanceChanges: decimal.NewFromInt(100),
						NextLedgerID:   "two",
						EndingBalance:  decimal.NewFromInt(200),
					},
				},
				Accounts:  []string{"1", "2"},
				CreatedAt: createdAt,
			},
			expectMovementResult: internal.MovementResult{
				MovementID: "one",
				Balances: map[string]internal.MovementEndingBalance{
					"1": {
						AccountID:          "1",
						NextLedgerID:       "one",
						NewBalance:         decimal.NewFromInt(0),
						PreviousBalance:    decimal.NewFromInt(100),
						PreviousLedgerID:   "",
						PreviousMovementID: "",
					},
					"2": {
						AccountID:          "2",
						NextLedgerID:       "two",
						NewBalance:         decimal.NewFromInt(200),
						PreviousBalance:    decimal.NewFromInt(100),
						PreviousLedgerID:   "",
						PreviousMovementID: "",
					},
				},
				Time: createdAt,
			},
			expectMovement: Movement{
				MovementID:     "one",
				IdempotencyKey: "one",
				CreatedAt:      createdAt,
				UpdatedAt:      sql.NullTime{},
			},
			expectAccountsBalance: map[string]GetAccountsBalanceRow{
				"1": {
					AccountID:     "1",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(0),
					CurrencyID:    1,
					LastLedgerID:  "one",
					CreatedAt:     createdAt,
					UpdatedAt:     sql.NullTime{Time: createdAt, Valid: true},
				},
				"2": {
					AccountID:     "2",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(200),
					CurrencyID:    1,
					LastLedgerID:  "two",
					CreatedAt:     createdAt,
					UpdatedAt:     sql.NullTime{Time: createdAt, Valid: true},
				},
			},
			expectAccountsLedger: []GetAccountsLedgerByMovementIDRow{
				{
					LedgerID:         "one",
					MovementID:       "one",
					AccountID:        "1",
					MovementSequence: 1,
					Amount:           decimal.NewFromInt(-100),
					PreviousLedgerID: "one",
					CreatedAt:        createdAt,
					ClientID:         sql.NullString{},
				},
				{
					LedgerID:         "two",
					MovementID:       "one",
					AccountID:        "2",
					MovementSequence: 1,
					Amount:           decimal.NewFromInt(100),
					PreviousLedgerID: "two",
					CreatedAt:        createdAt,
					ClientID:         sql.NullString{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			th, err := testHelper.ForkPostgresSchema(testCtx, testHelper.Postgres(), "ledger")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(th.CloseFunc(t))
			q := New(th.Postgres())

			for accountID, balance := range accountsSetup {
				if err := q.CreateLedgerAccount(testCtx, CreateLedgerAccount{
					AccountID:     accountID,
					AllowNegative: false,
					balance:       balance,
					Currency:      currency.IDR,
					CreatedAt:     createdAt,
				}); err != nil {
					t.Fatal(err)
				}
			}
			var accountsID []string

			for key := range test.expectAccountsBalance {
				accountsID = append(accountsID, key)
			}
			result, err := q.Move(testCtx, test.entries)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectMovementResult, result); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}

			// Check whether the accounts balances are correct..
			balances, err := q.GetAccountsBalanceMappedByAccID(testCtx, accountsID...)
			if err != nil {
				t.Fatal(err)
			}

			for _, accb := range test.expectAccountsBalance {
				b, ok := balances[accb.AccountID]
				if !ok {
					t.Fatalf("account id %s is not exists within balance search\nbalances: %v", accb.AccountID, balances)
				}
				if diff := cmp.Diff(accb, b); diff != "" {
					t.Fatalf("(-want/+got)\n%s", diff)
				}
			}
			// Check whether the ledger entries are correct.
			entries, err := q.GetAccountsLedgerByMovementID(testCtx, test.entries.MovementID)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectAccountsLedger, entries); diff != "" {
				t.Fatalf("accounts_ledger: (-want/+got)\n%s", diff)
			}
		})
	}
}

func TestSelectAccountsBalanceForMovement(t *testing.T) {
	t.Parallel()

	createdAt, err := time.Parse(time.RFC3339, "2024-03-25T00:01:10+07:00")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                   string
		movementID             string
		createdAt              time.Time
		accounts               []CreateAccountBalanceParams
		accountChanges         map[string]ledger.AccountMovementSummary
		expectLastBalanceInfo  map[string]internal.MovementEndingBalance
		expectBulkUpdateParams bulkUpdate
		selectForUpdateErr     error
	}{
		{
			name:       "simple update",
			movementID: "movement_id",
			createdAt:  createdAt,
			accounts: []CreateAccountBalanceParams{
				{
					AccountID:     "one",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					CreatedAt:     time.Now(),
				},
				{
					AccountID:     "two",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					CreatedAt:     time.Now(),
				},
				{
					AccountID:     "three",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					CreatedAt:     time.Now(),
				},
				{
					AccountID:     "four",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					CreatedAt:     time.Now(),
				},
			},
			accountChanges: map[string]ledger.AccountMovementSummary{
				"one": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "one",
					LastLedgerID:   "last",
				},
				"two": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "two",
					LastLedgerID:   "last",
				},
				"three": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "three",
					LastLedgerID:   "last",
				},
				"four": {
					BalanceChanges: decimal.NewFromInt(-100),
					NextLedgerID:   "four",
					LastLedgerID:   "last",
				},
			},
			expectLastBalanceInfo: map[string]internal.MovementEndingBalance{
				"one": {
					AccountID:       "one",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(200),
					NextLedgerID:    "one",
				},
				"two": {
					AccountID:       "two",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(200),
					NextLedgerID:    "two",
				},
				"three": {
					AccountID:       "three",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(200),
					NextLedgerID:    "three",
				},
				"four": {
					AccountID:       "four",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(0),
					NextLedgerID:    "four",
				},
			},
			expectBulkUpdateParams: bulkUpdate{
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
				Values: [][]any{
					{
						"one",
						"two",
						"three",
						"four",
					},
					{
						"200",
						"200",
						"200",
						"0",
					},
					{
						"one",
						"two",
						"three",
						"four",
					},
					{
						createdAt,
						createdAt,
						createdAt,
						createdAt,
					},
				},
			},
			selectForUpdateErr: nil,
		},
		{
			name:       "insufficient balance",
			movementID: "movement_id",
			createdAt:  createdAt,
			accounts: []CreateAccountBalanceParams{
				{
					AccountID:     "one",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(200),
					CreatedAt:     time.Now(),
				},
				{
					AccountID:     "two",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					CreatedAt:     time.Now(),
				},
				{
					AccountID:     "three",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					CreatedAt:     time.Now(),
				},
				{
					AccountID:     "four",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					CreatedAt:     time.Now(),
				},
			},
			accountChanges: map[string]ledger.AccountMovementSummary{
				"one": {
					BalanceChanges: decimal.NewFromInt(-300),
					NextLedgerID:   "one",
					LastLedgerID:   "last",
				},
				"two": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "two",
					LastLedgerID:   "last",
				},
				"three": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "three",
					LastLedgerID:   "last",
				},
				"four": {
					BalanceChanges: decimal.NewFromInt(-100),
					NextLedgerID:   "four",
					LastLedgerID:   "last",
				},
			},
			selectForUpdateErr: ledger.ErrInsufficientBalance,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// List all the accounts id based on the accounts changes.
			var accounts []string
			for acc := range test.accountChanges {
				accounts = append(accounts, acc)
			}
			// Fork a new connection to a new schema so we can test in parallel.
			th, err := testHelper.ForkPostgresSchema(context.Background(), testHelper.Postgres(), "ledger")
			if err != nil {
				t.Fatal(err)
			}
			tq := New(th.Postgres())

			for _, account := range test.accounts {
				if err := tq.CreateAccountBalance(context.Background(), account); err != nil {
					t.Fatal(err)
				}
			}

			var (
				bulkUpdateParams bulkUpdate
				gotInfo          map[string]internal.MovementEndingBalance
			)
			// Test inside the transaction as we are doing SELECT FOR UPDATE.
			gotErr := tq.WithTransact(context.Background(), sql.LevelReadUncommitted, func(ctx context.Context, q *Queries) error {
				bulkUpdateParams, gotInfo, err = selectAccountsBalanceForMovement(ctx, q, test.accountChanges, test.createdAt, accounts)
				return err
			})
			if !errors.Is(gotErr, test.selectForUpdateErr) {
				t.Fatalf("expecting error %v but got %v", test.selectForUpdateErr, gotErr)
			}
			if gotErr != nil {
				return
			}
			for accID, info := range test.expectLastBalanceInfo {
				if diff := cmp.Diff(info, gotInfo[accID]); diff != "" {
					t.Fatalf("%s: (-want/+got)\n%s", accID, diff)
				}
			}
			if diff := cmp.Diff(test.expectBulkUpdateParams, bulkUpdateParams); diff != "" {
				t.Fatalf("bulkUpdateParams: (-want/+got)\n%s", diff)
			}
		})
	}
}
