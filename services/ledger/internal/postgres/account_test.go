package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/studio-asd/pkg/postgres"

	"github.com/studio-asd/go-example/internal/currency"
)

func TestCreateLedgerAccounts(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	tests := []struct {
		name                  string
		createAccounts        []CreateLedgerAccount
		expectAccounts        []LedgerAccount
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
					AllowNegative:   true,
					Currency:        currency.IDR,
					CreatedAt:       now.Add(time.Second),
				},
				{
					AccountID:       "two",
					ParentAccountID: "",
					AllowNegative:   true,
					Currency:        currency.USD,
					CreatedAt:       now.Add(time.Second * 2),
				},
			},
			expectAccounts: []LedgerAccount{
				{
					AccountID:  "one",
					CurrencyID: 1,
					CreatedAt:  now.Add(time.Second),
				},
				{
					AccountID:  "two",
					CurrencyID: 2,
					CreatedAt:  now.Add(time.Second * 2),
				},
			},
			expectAccountsBalance: []GetAccountsBalanceRow{
				{
					AccountID:     "one",
					CurrencyID:    1,
					AllowNegative: true,
					Balance:       decimal.Zero,
					LastLedgerID:  "",
					CreatedAt:     now.Add(time.Second * 1),
				},
				{
					AccountID:     "two",
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
					AccountID:     "one_one",
					AllowNegative: true,
					Currency:      currency.IDR,
					CreatedAt:     now.Add(time.Second * 3),
				},
			},
			expectAccounts: []LedgerAccount{
				{
					AccountID:  "one_one",
					CurrencyID: 1,
					CreatedAt:  now.Add(time.Second * 3),
				},
			},
			expectAccountsBalance: []GetAccountsBalanceRow{
				{
					AccountID:     "one_one",
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
			name:           "conflict account id",
			isolatedSchema: true,
			createAccounts: []CreateLedgerAccount{
				{
					AccountID:       "one",
					ParentAccountID: "",
					AllowNegative:   true,
					Currency:        currency.IDR,
					CreatedAt:       now.Add(time.Second),
				},
				{
					AccountID:       "one",
					ParentAccountID: "",
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
			th, err := testHelper.ForkPostgresSchema(context.Background(), testQueries.Postgres(), "ledger")
			if err != nil {
				t.Fatal(err)
			}
			tq := New(th.Postgres())

			err = tq.CreateLedgerAccounts(context.Background(), test.createAccounts...)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
			if err != nil {
				return
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

func TestGetAccountsBalanceWithChildMappByAccID(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	tests := []struct {
		name           string
		createAccounts []CreateLedgerAccount
		findAccountIDs []string
		expect         map[string]GetAccountsBalanceRow
	}{
		{
			name: "one account",
			createAccounts: []CreateLedgerAccount{
				{
					AccountID:       "one",
					ParentAccountID: "",
					Currency:        currency.IDR,
					AllowNegative:   false,
					CreatedAt:       createdAt,
					balance:         decimal.NewFromInt(100),
				},
				{
					AccountID:       "two",
					ParentAccountID: "one",
					Currency:        currency.IDR,
					AllowNegative:   false,
					CreatedAt:       createdAt,
					balance:         decimal.NewFromInt(100),
				},
				{
					AccountID:       "three",
					ParentAccountID: "one",
					Currency:        currency.IDR,
					AllowNegative:   false,
					CreatedAt:       createdAt,
					balance:         decimal.NewFromInt(300),
				},
				{
					AccountID:       "four",
					ParentAccountID: "one",
					Currency:        currency.IDR,
					AllowNegative:   false,
					CreatedAt:       createdAt,
					balance:         decimal.NewFromInt(-500),
				},
			},
			findAccountIDs: []string{"one"},
			expect: map[string]GetAccountsBalanceRow{
				"one": {
					AccountID:     "one",
					AllowNegative: false,
					Balance:       decimal.NewFromInt(0),
					CurrencyID:    currency.IDR.ID,
					CreatedAt:     createdAt,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			th, err := testHelper.ForkPostgresSchema(context.Background(), testHelper.Postgres(), "ledger")
			if err != nil {
				t.Fatal(err)
			}
			q := New(th.Postgres())
			t.Log(th.DefaultSearchPath())

			if err := q.CreateLedgerAccounts(context.Background(), test.createAccounts...); err != nil {
				t.Fatal(err)
			}

			balances, err := q.GetAccountsBalanceWithChildMappedByAccID(context.Background(), test.findAccountIDs...)
			if err != nil {
				t.Fatal(err)
			}

			for accID, expect := range test.expect {
				got, ok := balances[accID]
				if !ok {
					t.Fatalf("account id %s not found", accID)
				}

				if diff := cmp.Diff(expect, got); diff != "" {
					t.Fatalf("(-want/+got)\n%s", diff)
				}
			}
		})
	}
}
