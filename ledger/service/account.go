package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/albertwidi/pkg/postgres"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/internal/currency"
	"github.com/albertwidi/go-example/ledger"
	ledgerpg "github.com/albertwidi/go-example/ledger/internal/postgres"
)

type Account struct {
	ID          string
	ParentID    string
	AccountType string
	Currency    *currency.Currency
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AccountBalance struct {
	ID            string
	AccountType   string
	AllowNegative bool
	LastLedgerID  string
	currency      *currency.Currency
	Balance       decimal.Decimal
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

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
		AccountType:     req.AccountType,
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

// GetAccountByIDs returns the account list by ids. The function can be used to retrieve one(1) or many(n) accounts
// but it doesn't guarantee the availability of the account when retrieving more than one(1) accounts.
// If the availability of an account is a must when retrieving more than one(1) accounts, the client need to handle it themselves.
func (l *Ledger) GetAccounts(ctx context.Context, ids ...string) ([]Account, error) {
	accs, err := l.q.GetAccounts(ctx, ids)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ledger.ErrAccountNotFound
		}
		return nil, err
	}
	accounts := make([]Account, len(accs))
	for idx, acc := range accs {
		// Retrieve the currency as we store the currency in its id in the database.
		c, err := currency.Currencies.GetByID(acc.CurrencyID)
		if err != nil {
			return nil, err
		}
		accounts[idx] = Account{
			ID:          acc.AccountID,
			ParentID:    acc.ParentAccountID,
			Currency:    c,
			AccountType: string(acc.AccountType),
			Status:      string(acc.AccountStatus),
			CreatedAt:   acc.CreatedAt,
			UpdatedAt:   acc.UpdatedAt.Time,
		}
	}
	return accounts, nil
}

func (l *Ledger) GetAccountsBalance(ctx context.Context, ids ...string) ([]AccountBalance, error) {
	accsBal, err := l.q.GetAccountsBalance(ctx, ids)
	if err != nil {
		return nil, err
	}

	accountsBalance := make([]AccountBalance, len(accsBal))
	for idx, accBal := range accsBal {
		accountsBalance[idx] = AccountBalance{
			ID: accBal.AccountID,
		}
	}
	return accountsBalance, nil
}
