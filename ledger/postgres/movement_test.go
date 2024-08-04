package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/ledger"
)

// TestMove tests whether the database records are correct when movement happens.
func TestMove(t *testing.T) {
	t.Parallel()

	createdAt, err := time.Parse(time.RFC3339, "2024-03-25T00:01:10+07:00")
	if err != nil {
		t.Fatal(err)
	}
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
		expectAccountsBalance map[string]AccountsBalance
		expectAccountsLedger  []AccountsLedger
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
					},
					"2": {
						BalanceChanges: decimal.NewFromInt(100),
						NextLedgerID:   "two",
					},
				},
				Accounts:  []string{"1", "2"},
				CreatedAt: createdAt,
			},
			expectAccountsBalance: map[string]AccountsBalance{
				"1": {
					AccountID:     "1",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(0),
					LastLedgerID:  "one",
					CreatedAt:     createdAt,
					UpdatedAt:     sql.NullTime{Time: createdAt, Valid: true},
				},
				"2": {
					AccountID:     "2",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(200),
					LastLedgerID:  "two",
					CreatedAt:     createdAt,
					UpdatedAt:     sql.NullTime{Time: createdAt, Valid: true},
				},
			},
			expectAccountsLedger: []AccountsLedger{
				{
					LedgerID:         "one",
					MovementID:       "one",
					AccountID:        "1",
					MovementSequence: 1,
					Amount:           decimal.NewFromInt(-100),
					PreviousLedgerID: "",
					CreatedAt:        createdAt,
					Timestamp:        timestamp + 1,
					ClientID:         sql.NullString{},
				},
				{
					LedgerID:         "two",
					MovementID:       "one",
					AccountID:        "2",
					MovementSequence: 1,
					Amount:           decimal.NewFromInt(100),
					PreviousLedgerID: "",
					CreatedAt:        createdAt,
					Timestamp:        timestamp + 2,
					ClientID:         sql.NullString{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			query, err := testHelper.ForkPostgresSchema(context.Background(), testQueries, "public")
			if err != nil {
				t.Fatal(err)
			}

			for accountID, balance := range accountsSetup {
				if err := query.CreateAccountBalance(testCtx, CreateAccountBalanceParams{
					AccountID:     accountID,
					AllowNegative: false,
					Balance:       balance,
					CreatedAt:     createdAt,
				}); err != nil {
					t.Fatal(err)
				}
			}
			var accountsID []string

			for key := range test.expectAccountsBalance {
				accountsID = append(accountsID, key)
			}
			if err := query.Move(testCtx, test.entries); err != nil {
				t.Fatal(err)
			}

			// Check whether the accounts balances are correct..
			balances, err := query.GetAccountsBalance(testCtx, accountsID)
			if err != nil {
				t.Fatal(err)
			}
			for _, accb := range test.expectAccountsBalance {
				var found bool
				for _, b := range balances {
					if accb.AccountID == b.AccountID {
						found = true
						if diff := cmp.Diff(accb, b); diff != "" {
							t.Fatalf("(-want/+got)\n%s", diff)
						}
						break
					}
				}
				if !found {
					t.Fatalf("account id %s is not exists within balance search", accb.AccountID)
				}
			}
			// Check whether the ledger entries are correct.
			entries, err := query.GetAccountsLedgerByMovementID(testCtx, test.entries.MovementID)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectAccountsLedger, entries); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
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
		name                     string
		movementID               string
		createdAt                time.Time
		accounts                 []CreateAccountBalanceParams
		accountChanges           map[string]ledger.AccountMovementSummary
		expectLastBalanceInfo    map[string]AccountLastBalanceInfo
		expectUpdateBalanceQuery string
		selectForUpdateErr       error
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
				},
				"two": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "two",
				},
				"three": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "three",
				},
				"four": {
					BalanceChanges: decimal.NewFromInt(-100),
					NextLedgerID:   "four",
				},
			},
			expectLastBalanceInfo: map[string]AccountLastBalanceInfo{
				"one": {
					AccountID:       "one",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(200),
					LastLedgerID:    "one",
				},
				"two": {
					AccountID:       "two",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(200),
					LastLedgerID:    "two",
				},
				"three": {
					AccountID:       "three",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(200),
					LastLedgerID:    "three",
				},
				"four": {
					AccountID:       "four",
					PreviousBalance: decimal.NewFromInt(100),
					NewBalance:      decimal.NewFromInt(0),
					LastLedgerID:    "four",
				},
			},
			expectUpdateBalanceQuery: `
UPDATE accounts_balance AS ab SET
	balance = v.balance,
	last_ledger_id = v.last_ledger_id,
	updated_at = v.updated_at
FROM (VALUES (200,'one','one','2024-03-25 00:01:10'),(200,'two','two','2024-03-25 00:01:10'),(200,'three','three','2024-03-25 00:01:10'),(0,'four','four','2024-03-25 00:01:10')) AS v(balance, last_ledger_id, account_id, updated_at)
WHERE ab.account_id = v.account_id;
`,
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
				},
				"two": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "two",
				},
				"three": {
					BalanceChanges: decimal.NewFromInt(100),
					NextLedgerID:   "three",
				},
				"four": {
					BalanceChanges: decimal.NewFromInt(-100),
					NextLedgerID:   "four",
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
			query, err := testHelper.ForkPostgresSchema(context.Background(), testQueries, "public")
			if err != nil {
				t.Fatal(err)
			}
			for _, account := range test.accounts {
				if err := query.CreateAccountBalance(context.Background(), CreateAccountBalanceParams{
					AccountID:     account.AccountID,
					AllowNegative: account.AllowNegative,
					Balance:       account.Balance,
					CreatedAt:     account.CreatedAt,
				}); err != nil {
					t.Fatal(err)
				}
			}

			var (
				updateBalanceQuery string
				gotInfo            map[string]AccountLastBalanceInfo
			)
			// Test inside the transaction as we are doing SELECT FOR UPDATE.
			gotErr := query.WithTransact(context.Background(), sql.LevelReadUncommitted, func(ctx context.Context, q *Queries) error {
				updateBalanceQuery, gotInfo, err = selectAccountsBalanceForMovement(ctx, q, test.accountChanges, test.createdAt, accounts)
				return err
			})
			if !errors.Is(gotErr, test.selectForUpdateErr) {
				t.Fatalf("expecting error %v but got %v", test.selectForUpdateErr, gotErr)
			}
			for accID, info := range test.expectLastBalanceInfo {
				if diff := cmp.Diff(info, gotInfo[accID]); diff != "" {
					t.Fatalf("%s: (-want/+got)\n%s", accID, diff)
				}
			}
			if diff := cmp.Diff(test.expectUpdateBalanceQuery, updateBalanceQuery); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}
			if test.expectUpdateBalanceQuery != updateBalanceQuery {
				t.Fatalf("expecting query\n%s\nbut got\n%s", test.expectUpdateBalanceQuery, updateBalanceQuery)
			}
		})
	}
}
