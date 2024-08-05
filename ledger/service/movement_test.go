package service

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/internal/currency"
	"github.com/albertwidi/go-example/ledger"
	ledgerpg "github.com/albertwidi/go-example/ledger/postgres"
	"github.com/albertwidi/pkg/postgres"
)

func TestMovementEntriesToLedgerEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		movementID string
		balances   map[string]ledgerpg.GetAccountsBalanceRow
		entries    []MovementEntry
		expect     ledger.MovementLedgerEntries
		err        error
	}{
		{
			name:       "one account to another account",
			movementID: "one",
			balances: map[string]ledgerpg.GetAccountsBalanceRow{
				"one": {
					AccountID:     "one",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					LastLedgerID:  "one_one",
					CurrencyID:    1,
				},
				"two": {
					AccountID:     "two",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_two",
					CurrencyID:    1,
				},
			},
			entries: []MovementEntry{
				{
					FromAccountID: "one",
					ToAccountID:   "two",
					Amount:        decimal.NewFromInt(100),
				},
			},
			expect: ledger.MovementLedgerEntries{
				LedgerEntries: []ledger.LedgerEntry{
					{
						MovementID:       "one",
						AccountID:        "one",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(-100),
					},
					{
						MovementID:       "one",
						AccountID:        "two",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(100),
					},
				},
				AccountsSummary: map[string]ledger.AccountMovementSummary{
					"one": {
						BalanceChanges: decimal.NewFromInt(-100),
						LastLedgerID:   "one_one",
						EndingBalance:  decimal.Zero,
					},
					"two": {
						BalanceChanges: decimal.NewFromInt(100),
						LastLedgerID:   "one_two",
						EndingBalance:  decimal.NewFromInt(100),
					},
				},
				Accounts: []string{
					"one",
					"two",
				},
			},
			err: nil,
		},
		{
			name:       "one account to another accounts",
			movementID: "one",
			balances: map[string]ledgerpg.GetAccountsBalanceRow{
				"one": {
					AccountID:     "one",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.NewFromInt(200),
					LastLedgerID:  "one_one",
					CurrencyID:    1,
				},
				"two": {
					AccountID:     "two",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_two",
					CurrencyID:    1,
				},
				"three": {
					AccountID:     "three",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_three",
					CurrencyID:    1,
				},
			},
			entries: []MovementEntry{
				{
					FromAccountID: "one",
					ToAccountID:   "two",
					Amount:        decimal.NewFromInt(100),
				},
				{
					FromAccountID: "one",
					ToAccountID:   "three",
					Amount:        decimal.NewFromInt(100),
				},
			},
			expect: ledger.MovementLedgerEntries{
				LedgerEntries: []ledger.LedgerEntry{
					{
						MovementID:       "one",
						AccountID:        "one",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(-100),
					},
					{
						MovementID:       "one",
						AccountID:        "two",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(100),
					},
					{
						MovementID:       "one",
						AccountID:        "one",
						MovementSequence: 2,
						Amount:           decimal.NewFromInt(-100),
					},
					{
						MovementID:       "one",
						AccountID:        "three",
						MovementSequence: 2,
						Amount:           decimal.NewFromInt(100),
					},
				},
				AccountsSummary: map[string]ledger.AccountMovementSummary{
					"one": {
						BalanceChanges: decimal.NewFromInt(-200),
						LastLedgerID:   "one_one",
						EndingBalance:  decimal.NewFromInt(0),
					},
					"two": {
						BalanceChanges: decimal.NewFromInt(100),
						LastLedgerID:   "one_two",
						EndingBalance:  decimal.NewFromInt(100),
					},
					"three": {
						BalanceChanges: decimal.NewFromInt(100),
						LastLedgerID:   "one_three",
						EndingBalance:  decimal.NewFromInt(100),
					},
				},
				Accounts: []string{
					"one",
					"two",
					"three",
				},
			},
			err: nil,
		},
		{
			name:       "many accounts to many accounts",
			movementID: "one",
			balances: map[string]ledgerpg.GetAccountsBalanceRow{
				"one": {
					AccountID:     "one",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.NewFromInt(200),
					LastLedgerID:  "one_one",
					CurrencyID:    1,
				},
				"two": {
					AccountID:     "two",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_two",
					CurrencyID:    1,
				},
				"three": {
					AccountID:     "three",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_three",
					CurrencyID:    1,
				},
				"four": {
					AccountID:     "four",
					AccountType:   ledger.AccountTypeUser,
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_four",
					CurrencyID:    1,
				},
			},
			entries: []MovementEntry{
				{
					FromAccountID: "one",
					ToAccountID:   "two",
					Amount:        decimal.NewFromInt(100),
				},
				{
					FromAccountID: "one",
					ToAccountID:   "three",
					Amount:        decimal.NewFromInt(100),
				},
				{
					FromAccountID: "three",
					ToAccountID:   "four",
					Amount:        decimal.NewFromInt(100),
				},
				{
					FromAccountID: "two",
					ToAccountID:   "three",
					Amount:        decimal.NewFromInt(100),
				},
			},
			expect: ledger.MovementLedgerEntries{
				LedgerEntries: []ledger.LedgerEntry{
					{
						MovementID:       "one",
						AccountID:        "one",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(-100),
					},
					{
						MovementID:       "one",
						AccountID:        "two",
						MovementSequence: 1,
						Amount:           decimal.NewFromInt(100),
					},
					{
						MovementID:       "one",
						AccountID:        "one",
						MovementSequence: 2,
						Amount:           decimal.NewFromInt(-100),
					},
					{
						MovementID:       "one",
						AccountID:        "three",
						MovementSequence: 2,
						Amount:           decimal.NewFromInt(100),
					},
					{
						MovementID:       "one",
						AccountID:        "three",
						MovementSequence: 3,
						Amount:           decimal.NewFromInt(-100),
					},
					{
						MovementID:       "one",
						AccountID:        "four",
						MovementSequence: 3,
						Amount:           decimal.NewFromInt(100),
					},
					{
						MovementID:       "one",
						AccountID:        "two",
						MovementSequence: 4,
						Amount:           decimal.NewFromInt(-100),
					},
					{
						MovementID:       "one",
						AccountID:        "three",
						MovementSequence: 4,
						Amount:           decimal.NewFromInt(100),
					},
				},
				AccountsSummary: map[string]ledger.AccountMovementSummary{
					"one": {
						BalanceChanges: decimal.NewFromInt(-200),
						LastLedgerID:   "one_one",
						EndingBalance:  decimal.Zero,
					},
					"two": {
						BalanceChanges: decimal.NewFromInt(0),
						LastLedgerID:   "one_two",
						EndingBalance:  decimal.Zero,
					},
					"three": {
						BalanceChanges: decimal.NewFromInt(100),
						LastLedgerID:   "one_three",
						EndingBalance:  decimal.NewFromInt(100),
					},
					"four": {
						BalanceChanges: decimal.NewFromInt(100),
						LastLedgerID:   "one_four",
						EndingBalance:  decimal.NewFromInt(100),
					},
				},
				Accounts: []string{
					"one",
					"two",
					"three",
					"four",
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		// Pre-build the ledger_id as we will generate the id on the fly.
		for idx, e := range test.expect.LedgerEntries {
			e.LedgerID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(e.MovementID+":"+e.AccountID+":"+strconv.Itoa(e.MovementSequence))).String()
			test.expect.LedgerEntries[idx] = e
			// Set the ledger id to the account summary as we need it for comparison.
			as := test.expect.AccountsSummary[e.AccountID]
			as.NextLedgerID = e.LedgerID
			test.expect.AccountsSummary[e.AccountID] = as
		}

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			le, err := movementEntriesToLedgerEntries(test.movementID, test.balances, test.entries...)
			if err != test.err {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(ledger.LedgerEntry{}, "CreatedAt", "Timestamp"),
				// We ignore the account summary here because it is a map, and the order of the map is not deterministic.
				cmpopts.IgnoreFields(ledger.MovementLedgerEntries{}, "AccountsSummary"),
			}
			if diff := cmp.Diff(test.expect, le, opts...); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}

			for accountID, accSum := range test.expect.AccountsSummary {
				got, ok := le.AccountsSummary[accountID]
				if !ok {
					t.Fatalf("account id %s not found in the summary", accountID)
				}
				if diff := cmp.Diff(accSum, got); diff != "" {
					t.Fatalf("(-want/+got)\n%s", diff)
				}
			}
		})
	}
}

func TestEligibleForMovement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		check checkEligible
		err   error
	}{
		{
			name: "same account",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "a",
				FromAccountType: ledger.AccountTypeUser,
				ToAccountType:   ledger.AccountTypeUser,
			},
			err: ledger.ErrCannotMoveToSelf,
		},
		{
			name: "different account, mismatch currency",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "b",
				FromAccountType: ledger.AccountTypeUser,
				ToAccountType:   ledger.AccountTypeUser,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 2,
				},
			},
			err: ledger.ErrMismatchCurrencies,
		},
		{
			name: "empty account type, from",
			check: checkEligible{
				FromAccountID: "a",
				ToAccountID:   "b",
				ToAccountType: ledger.AccountTypeUser,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrForbiddenAccountTypeTransfer,
		},
		{
			name: "empty account type, to",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "b",
				FromAccountType: ledger.AccountTypeUser,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrForbiddenAccountTypeTransfer,
		},
		{
			name: "empty account type, both",
			check: checkEligible{
				FromAccountID: "a",
				ToAccountID:   "b",
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrForbiddenAccountTypeTransfer,
		},
		{
			name: "forbidden account type, deposit to withdrawal",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "b",
				FromAccountType: ledger.AccountTypeDeposit,
				ToAccountType:   ledger.AccountTypeWithdrawal,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrForbiddenAccountTypeTransfer,
		},
		{
			name: "forbidden account type, user to deposit",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "b",
				FromAccountType: ledger.AccountTypeUser,
				ToAccountType:   ledger.AccountTypeDeposit,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrForbiddenAccountTypeTransfer,
		},
		{
			name: "forbidden account type, withdrawal to deposit",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "b",
				FromAccountType: ledger.AccountTypeWithdrawal,
				ToAccountType:   ledger.AccountTypeDeposit,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrForbiddenAccountTypeTransfer,
		},
		{
			name: "forbidden account type, withdrawal to user",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "b",
				FromAccountType: ledger.AccountTypeWithdrawal,
				ToAccountType:   ledger.AccountTypeUser,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrForbiddenAccountTypeTransfer,
		},
		{
			name: "ok",
			check: checkEligible{
				FromAccountID:   "a",
				ToAccountID:     "b",
				FromAccountType: ledger.AccountTypeUser,
				ToAccountType:   ledger.AccountTypeUser,
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := checkEligibleForMovement(test.check); !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
		})
	}
}

func TestTransact(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	tq, err := testHelper.ForkPostgresSchema(context.Background(), testQueries, "public")
	if err != nil {
		t.Fatal(err)
	}
	tl := New(tq)

	newTableQuery := "CREATE TABLE IF NOT EXISTS trasact_test(id int PRIMARY KEY);"
	err = tq.Do(context.Background(), func(ctx context.Context, pg *postgres.Postgres) error {
		_, err := pg.Exec(ctx, newTableQuery)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("transact_success", func(t *testing.T) {
		t.Parallel()

		fn := func(ctx context.Context, pg *postgres.Postgres) error {
			insertQuery := "INSERT INTO transact_test VALUES(1);"
			_, err := pg.Exec(ctx, insertQuery)
			if err != nil {
				return err
			}
			return nil
		}

		tl.Transact(context.Background(), CreateTransaction{
			UniqueID: uuid.NewString(),
			Entries:  []MovementEntry{},
		}, fn)
	})

	t.Run("transact_failed", func(t *testing.T) {
		t.Parallel()
	})
}
