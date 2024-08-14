package service

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/albertwidi/go-example/internal/currency"
	"github.com/albertwidi/go-example/ledger"
)

func TestCreateAccount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	th, err := testHelper.ForkPostgresSchema(context.Background(), testHelper.Queries(), "public")
	if err != nil {
		t.Fatal(err)
	}
	ls := New(th.Queries())

	tests := []struct {
		name             string
		ca               CreateAccount
		expectAcc        Account
		expectAccBalance AccountBalance
	}{
		{
			name: "account with no funds",
			ca: CreateAccount{
				ID:            "one",
				Currency:      currency.IDR,
				Status:        ledger.AccountStatusActive,
				AllowNegative: false,
				AccountType:   ledger.AccountTypeUser,
			},
			expectAcc: Account{
				ID:          "one",
				ParentID:    "",
				Currency:    currency.IDR,
				Status:      ledger.AccountStatusActive,
				AccountType: ledger.AccountTypeUser,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ls.CreateAccount(
				context.Background(),
				test.ca,
			); err != nil {
				t.Fatal(err)
			}

			accs, err := ls.GetAccounts(context.Background(), test.ca.ID)
			if err != nil {
				t.Fatal(err)
			}
			if len(accs) > 1 {
				t.Fatalf("expecting one account %s", test.ca.ID)
			}
			if diff := cmp.Diff(test.expectAcc, accs[0]); diff != "" {
				t.Fatalf("Account (-want/+got):\n%s", diff)
			}
		})
	}
}
