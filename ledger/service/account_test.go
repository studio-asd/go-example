package service

import (
	"testing"

	"github.com/albertwidi/go-example/internal/currency"
	"github.com/albertwidi/go-example/ledger"
)

func TestCreateAccount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	tests := []struct {
		name string
		ca   CreateAccount
	}{
		{
			name: "account with no funds",
			ca: CreateAccount{
				ID:            "one",
				Currency:      currency.IDR,
				AllowNegative: false,
				AccountType:   ledger.AccountTypeUser,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
		})
	}
}
