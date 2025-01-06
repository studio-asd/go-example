package api

import (
	"context"
	"sync"
	"testing"

	"github.com/studio-asd/pkg/postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/testing/protocmp"

	prototesting "github.com/studio-asd/go-example/internal/testing/proto"
	ledgerv1 "github.com/studio-asd/go-example/proto/api/ledger/v1"
	"github.com/studio-asd/go-example/services/ledger"
)

func TestTransact(t *testing.T) {
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

	t.Run("simple_transact", func(t *testing.T) {
		resp := createSimpleTestAccounts(t, testAPI)

		depositAccount := resp.GetAccounts()[2].GetAccountId()
		testAccount := resp.GetAccounts()[0].GetAccountId()

		txResp, err := testAPI.Transact(
			context.Background(),
			&ledgerv1.TransactRequest{
				IdempotencyKey: "test",
				MovementEntries: []*ledgerv1.MovementEntry{
					{
						FromAccount: &ledgerv1.MovementEntry_FromAccount{
							FromAccountId: depositAccount,
						},
						ToAccountId: testAccount,
						Amount:      "100",
						ClientId:    "test_client_id",
					},
				},
			},
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}

		// Check movementId we copy the movement id to the expected proto.Message.
		if txResp.GetMovementId() == "" {
			t.Fatal("movement id is empty")
		}
		if err := txResp.TransactTime.CheckValid(); err != nil {
			t.Fatal(err)
		}
		expect := &ledgerv1.TransactResponse{
			MovementId: txResp.GetMovementId(),
			LedgerEntries: []*ledgerv1.TransactResponse_LedgerEntry{
				{
					LedgerId:         uuid.NewSHA1(uuid.NameSpaceOID, []byte(txResp.GetMovementId()+":"+depositAccount+":"+"1")).String(),
					ClientId:         "test_client_id",
					MovementSequence: 1,
				},
				{
					LedgerId:         uuid.NewSHA1(uuid.NameSpaceOID, []byte(txResp.GetMovementId()+":"+testAccount+":"+"1")).String(),
					ClientId:         "test_client_id",
					MovementSequence: 1,
				},
			},
			EndingBalances: []*ledgerv1.TransactResponse_Balance{
				{
					AccountId:          depositAccount,
					LedgerId:           uuid.NewSHA1(uuid.NameSpaceOID, []byte(txResp.GetMovementId()+":"+depositAccount+":"+"1")).String(),
					PreviousLedgerId:   "",
					PreviousMovementId: "",
					PreviousBalance:    "0",
					NewBalance:         "-100",
				},
				{
					AccountId:          testAccount,
					LedgerId:           uuid.NewSHA1(uuid.NameSpaceOID, []byte(txResp.GetMovementId()+":"+testAccount+":"+"1")).String(),
					PreviousLedgerId:   "",
					PreviousMovementId: "",
					PreviousBalance:    "0",
					NewBalance:         "100",
				},
			},
			TransactTime: txResp.TransactTime,
		}

		sortEndingBalances := protocmp.SortRepeatedFields(&ledgerv1.TransactResponse{}, "ending_balances")
		if diff := cmp.Diff(expect, txResp, sortEndingBalances, protocmp.Transform()); diff != "" {
			t.Fatalf(
				"(-want/+got)\n%s\n\nexpect:\n%s\ngot:\n%s",
				diff,
				prototesting.ToJson(t, expect),
				prototesting.ToJson(t, txResp),
			)
		}
	})

	t.Run("transact_foregin_func_success", func(t *testing.T) {
		t.Skip()
		t.Parallel()

		fn := func(ctx context.Context, pg *postgres.Postgres, info ledger.MovementInfo) error {
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
		t.Skip()
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
		}, nil)
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
					FromAccount: &ledgerv1.MovementEntry_FromAccount{
						FromAccountId: depositAccount.GetAccountId(),
					},
					ToAccountId: firstAccount.GetAccountId(),
					Amount:      "100",
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
							FromAccount: &ledgerv1.MovementEntry_FromAccount{
								FromAccountId: firstAccount.GetAccountId(),
							},
							ToAccountId: secondAccount.GetAccountId(),
							Amount:      "10",
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

func createSimpleTestAccounts(t *testing.T, api *API) *ledgerv1.CreateLedgerAccountsResponse {
	t.Helper()
	accountsResp, err := api.CreateAccounts(context.Background(), &ledgerv1.CreateLedgerAccountsRequest{
		Accounts: []*ledgerv1.CreateLedgerAccountsRequest_Account{
			{
				AllowNegative: false,
				CurrencyId:    1,
			},
			{
				AllowNegative: false,
				CurrencyId:    1,
			},
			// The third account is only for deposit.
			{
				AllowNegative: true,
				CurrencyId:    1,
			},
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return accountsResp
}
