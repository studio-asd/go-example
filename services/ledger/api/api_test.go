package api

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/albertwidi/pkg/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

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
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_Account{
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
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_Account{
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
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_Account{
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
				Accounts: []*ledgerv1.CreateLedgerAccountsRequest_Account{
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
	// The test is about having two different accounts. One account have 100, and another account is 0.
	// Then we will try to move all the 100 to the second account using goroutines, then at the end of the test
	// we will check whether the first account gone below 0.
	t.Run("negative not allowed", func(t *testing.T) {
		th, err := testHelper.ForkPostgresSchema(context.Background(), testAPI.queries, "public")
		if err != nil {
			t.Fatal(err)
		}
		a := New(th.Queries())

		accountsResp, err := a.CreateAccounts(context.Background(), &ledgerv1.CreateLedgerAccountsRequest{
			Accounts: []*ledgerv1.CreateLedgerAccountsRequest_Account{
				{
					AllowNegative: false,
					CurrencyId:    1,
				},
				{
					AllowNegative: false,
					CurrencyId:    1,
				},
				// The third account is only to deposit 100 to the first account.
				{
					AllowNegative: true,
					CurrencyId:    1,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		firstAccount := accountsResp.Accounts[0]
		secondAccount := accountsResp.Accounts[1]
		depositAccount := accountsResp.Accounts[2]

		_, err = a.Transact(context.Background(), &ledgerv1.TransactRequest{
			IdempotencyKey: "one",
			MovementEntries: []*ledgerv1.MovementEntry{
				{
					FromAccountId: depositAccount.GetAccountId(),
					ToAccountId:   firstAccount.GetAccountId(),
					Amount:        "100",
				},
			},
		}, nil)
		if err != nil {
			t.Fatal(err)
		}

		var errMu sync.Mutex
		var errArr []error
		wg := sync.WaitGroup{}
		// Spawn 20 goroutines and do a transfer of 10 for each goroutine. With 20 goroutines, we should have 10 success transactions
		// and 10 failed transactions with total of 100 of money movements.
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := a.Transact(context.Background(), &ledgerv1.TransactRequest{
					IdempotencyKey: uuid.NewString(),
					MovementEntries: []*ledgerv1.MovementEntry{
						{
							FromAccountId: firstAccount.GetAccountId(),
							ToAccountId:   secondAccount.GetAccountId(),
							Amount:        "10",
						},
					},
				}, nil)
				if err != nil {
					errMu.Lock()
					errArr = append(errArr, err)
					errMu.Unlock()
				}
			}()
		}
		wg.Wait()

		if len(errArr) != 10 {
			t.Fatalf("expecting error length of %d but got %d", 10, len(errArr))
		}

		resp, err := a.GetAccountsBalance(context.Background(), &ledgerv1.GetAccountsBalanceRequest{
			AccountIds: []string{firstAccount.GetAccountId(), secondAccount.GetAccountId()},
		})
		if err != nil {
			t.Fatal(err)
		}
		// Check the first account, the amount of balance should be zero(0).
		if resp.Balances[0].GetAccountId() != firstAccount.GetAccountId() {
			t.Fatal("wrong account id for the first acccount")
		}
		if resp.Balances[0].GetBalance() != decimal.Zero.String() {
			t.Fatalf("first account balance is not zero, got %s", resp.Balances[0].String())
		}
		// Check the second account, the amount of balance should be a hundred(100).
		if resp.Balances[1].GetAccountId() != secondAccount.GetAccountId() {
			t.Fatal("wrong account id for the first acccount")
		}
		if resp.Balances[1].GetBalance() != "100" {
			t.Fatalf("second account balance is not zero, got %s", resp.Balances[1].String())
		}
	})
}
