package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/albertwidi/pkg/postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/internal/currency"
	"github.com/albertwidi/go-example/ledger"
)

func TestCreateLedgerAccounts(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name                  string
		createAccounts        []CreateLedgerAccount
		expectAccounts        []Account
		expectAccountsBalance []GetAccountsBalanceRow
		// isolatedSchema fork the schema and creates a new schema for the test.
		isolatedSchema bool
		err            error
	}{
		{
			name: "many accounts",
			createAccounts: []CreateLedgerAccount{
				{
					AccountID:       "one",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   string(AccountStatusActive),
					AllowNegative:   true,
					Currency:        currency.IDR,
					CreatedAt:       now.Add(time.Second),
				},
				{
					AccountID:       "two",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   string(AccountStatusActive),
					AllowNegative:   true,
					Currency:        currency.USD,
					CreatedAt:       now.Add(time.Second * 2),
				},
			},
			expectAccounts: []Account{
				{
					AccountID:       "one",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   AccountStatusActive,
					CurrencyID:      1,
					CreatedAt:       now.Add(time.Second),
				},
				{
					AccountID:       "two",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   AccountStatusActive,
					CurrencyID:      2,
					CreatedAt:       now.Add(time.Second * 2),
				},
			},
			expectAccountsBalance: []GetAccountsBalanceRow{
				{
					AccountID:     "one",
					AccountStatus: AccountStatusActive,
					AccountType:   ledger.AccountTypeDeposit,
					CurrencyID:    1,
					AllowNegative: true,
					Balance:       decimal.Zero,
					LastLedgerID:  "",
					CreatedAt:     now.Add(time.Second * 1),
				},
				{
					AccountID:     "two",
					AccountStatus: AccountStatusActive,
					AccountType:   ledger.AccountTypeDeposit,
					CurrencyID:    2,
					AllowNegative: true,
					Balance:       decimal.Zero,
					LastLedgerID:  "",
					CreatedAt:     now.Add(time.Second * 2),
				},
			},
			err: nil,
		},
		{
			name: "one account",
			createAccounts: []CreateLedgerAccount{
				{
					AccountID:       "one_one",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   string(AccountStatusActive),
					AllowNegative:   true,
					Currency:        currency.IDR,
					CreatedAt:       now.Add(time.Second * 3),
				},
			},
			expectAccounts: []Account{
				{
					AccountID:       "one_one",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   AccountStatusActive,
					CurrencyID:      1,
					CreatedAt:       now.Add(time.Second * 3),
				},
			},
			expectAccountsBalance: []GetAccountsBalanceRow{
				{
					AccountID:     "one_one",
					AccountStatus: AccountStatusActive,
					AccountType:   ledger.AccountTypeDeposit,
					CurrencyID:    1,
					AllowNegative: true,
					Balance:       decimal.Zero,
					LastLedgerID:  "",
					CreatedAt:     now.Add(time.Second * 3),
				},
			},
			err: nil,
		},
		{
			name:           "isolated, conflict account id",
			isolatedSchema: true,
			createAccounts: []CreateLedgerAccount{
				{
					AccountID:       "one",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   string(AccountStatusActive),
					AllowNegative:   true,
					Currency:        currency.IDR,
					CreatedAt:       now.Add(time.Second),
				},
				{
					AccountID:       "one",
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AccountStatus:   string(AccountStatusActive),
					AllowNegative:   true,
					Currency:        currency.USD,
					CreatedAt:       now.Add(time.Second * 2),
				},
			},
			err: postgres.ErrUniqueViolation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var err error
			tq := testHelper.Queries()
			if test.isolatedSchema {
				tq, err = testHelper.ForkPostgresSchema(context.Background(), testHelper.Queries(), "public")
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := tq.CreateLedgerAccounts(context.Background(), test.createAccounts...); !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}

			var accounts []string
			for _, ca := range test.createAccounts {
				accounts = append(accounts, ca.AccountID)
			}

			gotAcc, err := tq.GetAccounts(context.Background(), accounts)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectAccounts, gotAcc); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}

			gotAccb, err := tq.GetAccountsBalance(context.Background(), accounts)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectAccountsBalance, gotAccb); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}
		})
	}
}
