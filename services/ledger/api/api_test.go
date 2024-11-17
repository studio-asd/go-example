package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/albertwidi/pkg/postgres"

	"github.com/albertwidi/go-example/internal/currency"
	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
	"github.com/albertwidi/go-example/services/ledger"
	ledgerpg "github.com/albertwidi/go-example/services/ledger/internal/postgres"
)

func TestCreateAccounts(t *testing.T) {
	t.Parallel()

	th, err := testHelper.ForkPostgresSchema(context.Background(), testAPI.queries, "public")
	if err != nil {
		t.Fatal(err)
	}
	api := New(th.Queries())

	// Setup the test, we will create multiple accounts with and without parent account as the basis.
	// We will use queries directly as it doesn't have any checks, so its easy to create the data.
	if err := api.queries.CreateLedgerAccounts(
		context.Background(),
		[]ledgerpg.CreateLedgerAccount{
			{
				AccountID:     "no_parent",
				AccountStatus: ledgerpg.AccountStatusActive,
				AllowNegative: false,
				Currency:      currency.IDR,
				CreatedAt:     time.Now(),
			},
			{
				AccountID:       "with_parent",
				ParentAccountID: "no_parent",
				AccountStatus:   ledgerpg.AccountStatusActive,
				AllowNegative:   false,
				Currency:        currency.IDR,
				CreatedAt:       time.Now(),
			},
			{
				AccountID:     "no_parent_inactive",
				AccountStatus: ledgerpg.AccountStatusInactive,
				AllowNegative: false,
				Currency:      currency.IDR,
				CreatedAt:     time.Now(),
			},
		}...,
	); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		request *ledgerv1.CreateLedgerAccountsRequest
		err     error
	}{
		{
			name: "with no parent",
			request: &ledgerv1.CreateLedgerAccountsRequest{
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_CreateAccount{
					{
						AllowNegative: true,
						CurrencyId:    1,
					},
					{
						AllowNegative: false,
						CurrencyId:    2,
					},
				},
			},
			err: nil,
		},
		{
			name: "with parent",
			request: &ledgerv1.CreateLedgerAccountsRequest{
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_CreateAccount{
					{
						ParentAccountId: "no_parent",
						AllowNegative:   true,
						CurrencyId:      1,
					},
					{
						ParentAccountId: "no_parent",
						AllowNegative:   false,
						CurrencyId:      2,
					},
				},
			},
			err: nil,
		},
		{
			name: "with parent, parent has parent",
			request: &ledgerv1.CreateLedgerAccountsRequest{
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_CreateAccount{
					{
						ParentAccountId: "with_parent",
						AllowNegative:   true,
						CurrencyId:      1,
					},
					{
						ParentAccountId: "no_parent",
						AllowNegative:   false,
						CurrencyId:      2,
					},
				},
			},
			err: ledger.ErrAccountHasParent,
		},
		{
			name: "with parent, account inactive",
			request: &ledgerv1.CreateLedgerAccountsRequest{
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_CreateAccount{
					{
						ParentAccountId: "no_parent_inactive",
						AllowNegative:   true,
						CurrencyId:      1,
					},
					{
						ParentAccountId: "no_parent",
						AllowNegative:   false,
						CurrencyId:      2,
					},
				},
			},
			err: ledger.ErrAccountInactive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := api.CreateAccounts(context.Background(), test.request)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
			if err != nil {
				return
			}
			if len(resp.Accounts) == 0 {
				t.Fatal("got zero accounts for response")
			}
			for _, acc := range resp.Accounts {
				if acc.AccountId == "" {
					t.Fatal("response account_id is empty")
				}
			}
		})
	}
}

func TestTransact(t *testing.T) {
	t.Skip()
	t.Parallel()

	tq, err := testHelper.ForkPostgresSchema(context.Background(), testQueries, "public")
	if err != nil {
		t.Fatal(err)
	}
	tl := New(tq.Queries())

	newTableQuery := "CREATE TABLE IF NOT EXISTS trasact_test(id int PRIMARY KEY);"
	err = tq.Queries().Do(context.Background(), func(ctx context.Context, pg *postgres.Postgres) error {
		_, err := pg.Exec(ctx, newTableQuery)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("transact_foregin_func_success", func(t *testing.T) {
		t.Parallel()

		fn := func(ctx context.Context, pg *postgres.Postgres) error {
			insertQuery := "INSERT INTO transact_test VALUES(1);"
			_, err := pg.Exec(ctx, insertQuery)
			if err != nil {
				return err
			}
			return nil
		}

		tl.Transact(context.Background(), &ledgerv1.TransactRequest{
			IdempotencyKey: "",
		}, fn)
	})

	t.Run("transact_foreign_function_failed", func(t *testing.T) {
		t.Parallel()
	})
}

// TestTransactCorrectness tests the correctness and end result of the concurrent transactions.
// The idea of this test to list all the transactions upfront and check whether the ending balance is correct.
func TestTransactCorrectness(t *testing.T) {

}
