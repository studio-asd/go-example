package service

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/albertwidi/pkg/postgres"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/internal/currency"
	"github.com/albertwidi/go-example/ledger"
	ledgerpg "github.com/albertwidi/go-example/ledger/postgres"
)

type CreateAccount struct {
	ID            string
	ParentID      string
	AccountType   string
	Currency      *currency.Currency
	AllowNegative bool
	Status        string
	// WithFunds creates an account with starting funds. This parameter can only be done inside test, and will be ignored
	// if testing.Testing() is false.
	WithFunds struct {
		Amount decimal.Decimal
	}
}

// onceCreateAccForTest will be used to ensure we are only creates the account for all currencies once for testing purpose.
var onceCreateAccForTest sync.Once

func (l *Ledger) CreateAccount(ctx context.Context, req CreateAccount) error {
	cla := ledgerpg.CreateLedgerAccount{
		AccountID:       req.ID,
		ParentAccountID: req.ParentID,
		AccountStatus:   req.Status,
		AllowNegative:   req.AllowNegative,
		Currency:        req.Currency,
	}

	// If it is a test and with funds parameter is not empty, then we should create an account with some funds inside it. The funds will still
	// be given from a ledger account inside the system.
	if !req.WithFunds.Amount.IsZero() && testing.Testing() {
		onceCreateAccForTest.Do(func() {
			clas := make([]ledgerpg.CreateLedgerAccount, len(currency.Currencies.List()))
			for _, curr := range currency.Currencies.List() {
				clas = append(clas, ledgerpg.CreateLedgerAccount{
					// This means the account_id will always be 'test_deposit_{CURRENCY_NAME}'.
					AccountID:       fmt.Sprintf("test_deposit_%s", curr.Name),
					ParentAccountID: "",
					AccountType:     ledger.AccountTypeDeposit,
					AllowNegative:   true,
					Currency:        curr,
				})
			}
			if err := l.q.CreateLedgerAccounts(ctx, clas...); err != nil {
				panic(err)
			}
		})

		return l.Transact(ctx, CreateTransaction{}, func(ctx context.Context, pg *postgres.Postgres) error {
			q := ledgerpg.New(pg)
			if err := q.CreateLedgerAccount(ctx, cla); err != nil {
				return err
			}
			return nil
		})
	}
	return l.q.CreateLedgerAccount(ctx, cla)
}
