package api

import (
	"context"
	"sync"
	"testing"

	"github.com/albertwidi/pkg/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
)

func TestTransact(t *testing.T) {
	t.Skip()
	t.Parallel()

	tq, err := testHelper.ForkPostgresSchema(context.Background(), testQueries)
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
		th, err := testHelper.ForkPostgresSchema(context.Background(), testAPI.queries)
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
