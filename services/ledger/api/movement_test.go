package api

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/studio-asd/go-example/internal/currency"
	ledgerv1 "github.com/studio-asd/go-example/proto/api/ledger/v1"
	"github.com/studio-asd/go-example/services/ledger"
	ledgerpg "github.com/studio-asd/go-example/services/ledger/internal/postgres"
)

func TestCreateLedgerEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		movementID     string
		idempotencyKey string
		balances       map[string]ledgerpg.GetAccountsBalanceRow
		entries        []*ledgerv1.MovementEntry
		expect         ledger.MovementLedgerEntries
		err            error
	}{
		{
			name:           "one account to another account",
			movementID:     "one",
			idempotencyKey: "one",
			balances: map[string]ledgerpg.GetAccountsBalanceRow{
				"one": {
					AccountID:     "one",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(100),
					LastLedgerID:  "one_one",
					CurrencyID:    1,
				},
				"two": {
					AccountID:     "two",
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_two",
					CurrencyID:    1,
				},
			},
			entries: []*ledgerv1.MovementEntry{
				{
					FromAccountId: "one",
					ToAccountId:   "two",
					Amount:        "100",
				},
			},
			expect: ledger.MovementLedgerEntries{
				MovementID:     "one",
				IdempotencyKey: "one",
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
			name:           "one account to another accounts",
			movementID:     "one",
			idempotencyKey: "two",
			balances: map[string]ledgerpg.GetAccountsBalanceRow{
				"one": {
					AccountID:     "one",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(200),
					LastLedgerID:  "one_one",
					CurrencyID:    1,
				},
				"two": {
					AccountID:     "two",
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_two",
					CurrencyID:    1,
				},
				"three": {
					AccountID:     "three",
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_three",
					CurrencyID:    1,
				},
			},
			entries: []*ledgerv1.MovementEntry{
				{
					FromAccountId: "one",
					ToAccountId:   "two",
					Amount:        "100",
				},
				{
					FromAccountId: "one",
					ToAccountId:   "three",
					Amount:        "100",
				},
			},
			expect: ledger.MovementLedgerEntries{
				MovementID:     "one",
				IdempotencyKey: "two",
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
			name:           "many accounts to many accounts",
			movementID:     "one",
			idempotencyKey: "three",
			balances: map[string]ledgerpg.GetAccountsBalanceRow{
				"one": {
					AccountID:     "one",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(200),
					LastLedgerID:  "one_one",
					CurrencyID:    1,
				},
				"two": {
					AccountID:     "two",
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_two",
					CurrencyID:    1,
				},
				"three": {
					AccountID:     "three",
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_three",
					CurrencyID:    1,
				},
				"four": {
					AccountID:     "four",
					AllowNegative: false,
					Balance:       decimal.Zero,
					LastLedgerID:  "one_four",
					CurrencyID:    1,
				},
			},
			entries: []*ledgerv1.MovementEntry{
				{
					FromAccountId: "one",
					ToAccountId:   "two",
					Amount:        "100",
				},
				{
					FromAccountId: "one",
					ToAccountId:   "three",
					Amount:        "100",
				},
				{
					FromAccountId: "one",
					ToAccountId:   "four",
					Amount:        "100",
				},
				{
					FromAccountId: "one",
					ToAccountId:   "three",
					Amount:        "100",
				},
			},
			expect: ledger.MovementLedgerEntries{
				MovementID:     "one",
				IdempotencyKey: "three",
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

			le, err := createLedgerEntries(test.movementID, test.idempotencyKey, test.balances, test.entries...)
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
				FromAccountID: "a",
				ToAccountID:   "a",
			},
			err: ledger.ErrCannotMoveToSelf,
		},
		{
			name: "different account, mismatch currency",
			check: checkEligible{
				FromAccountID: "a",
				ToAccountID:   "b",
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
			name: "from account empty",
			check: checkEligible{
				ToAccountID: "b",
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrAccountSourceOrDestinationEmpty,
		},
		{
			name: "to account empty",
			check: checkEligible{
				FromAccountID: "a",
				FromCurrency: &currency.Currency{
					ID: 1,
				},
				ToCurrency: &currency.Currency{
					ID: 1,
				},
			},
			err: ledger.ErrAccountSourceOrDestinationEmpty,
		},
		{
			name: "ok",
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
