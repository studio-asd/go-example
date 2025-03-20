package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/studio-asd/pkg/postgres"

	"github.com/studio-asd/go-example/internal/currency"
)

type CreateLedgerAccount struct {
	AccountID       string
	Name            string
	Description     string
	ParentAccountID string
	AllowNegative   bool
	Currency        *currency.Currency
	CreatedAt       time.Time
	// balance can only be set internally for testing purpose.
	balance decimal.Decimal
}

func (q *Queries) CreateLedgerAccounts(ctx context.Context, c ...CreateLedgerAccount) error {
	fn := func(ctx context.Context, q *Queries) error {
		var err error
		for _, cla := range c {
			err := q.CreateLedgerAccount(ctx, cla)
			if err != nil {
				return err
			}
		}
		return err
	}
	return q.WithMetrics(ctx, "createLedgerAccounts", func(ctx context.Context, q *Queries) error {
		return q.ensureInTransact(ctx, sql.LevelReadCommitted, fn)
	})
}

func (q *Queries) CreateLedgerAccount(ctx context.Context, c CreateLedgerAccount) error {
	// We should not allow balance to be set if it used outside of the testing scope.
	if !c.balance.IsZero() && !testing.Testing() {
		return errors.New("balance can only be set inside testing")
	}
	var parentAccountID sql.NullString
	if c.ParentAccountID != "" {
		parentAccountID = sql.NullString{
			String: c.ParentAccountID,
			Valid:  true,
		}
	}

	fn := func(ctx context.Context, qr *Queries) error {
		if err := qr.CreateAccount(ctx, CreateAccountParams{
			AccountID:       c.AccountID,
			Name:            c.Name,
			Description:     c.Description,
			ParentAccountID: parentAccountID,
			CurrencyID:      c.Currency.ID,
			CreatedAt:       c.CreatedAt,
		}); err != nil {
			return err
		}
		if err := qr.CreateAccountBalance(ctx, CreateAccountBalanceParams{
			AccountID:       c.AccountID,
			ParentAccountID: parentAccountID,
			AllowNegative:   c.AllowNegative,
			Balance:         c.balance,
			CurrencyID:      c.Currency.ID,
			// LastLedgerID is always empty at first.
			LastLedgerID: "",
			CreatedAt:    c.CreatedAt,
		}); err != nil {
			return err
		}
		return nil
	}
	return q.WithMetrics(ctx, "createLedgerAccount", func(ctx context.Context, q *Queries) error {
		return q.ensureInTransact(ctx, sql.LevelReadCommitted, fn)
	})
}

// GetAccountsBalanceMappByAccID returns the account balance of accounts using map and account_id as its key. The function can be used to quickly look into
// the account information(O(1)) rather than looking from the entire accounts range(O(n)).
func (q *Queries) GetAccountsBalanceMappedByAccID(ctx context.Context, accounts ...string) (map[string]GetAccountsBalanceRow, error) {
	accountsBalance := make(map[string]GetAccountsBalanceRow)
	err := q.WithMetrics(ctx, "getAccountsBalanceMappedbyAccID", func(ctx context.Context, q *Queries) error {
		return q.db.RunQuery(ctx, getAccountsBalance, func(rows *postgres.RowsCompat) error {
			var i GetAccountsBalanceRow
			if err := rows.Scan(
				&i.AccountID,
				&i.AllowNegative,
				&i.Balance,
				&i.CurrencyID,
				&i.LastLedgerID,
				&i.LastMovementID,
				&i.CreatedAt,
				&i.UpdatedAt,
			); err != nil {
				return err
			}
			if err := rows.Err(); err != nil {
				return err
			}
			accountsBalance[i.AccountID] = i
			return nil
		})
	})
	return accountsBalance, err
}

func (q *Queries) GetAccountsBalanceWithChildMappedByAccID(ctx context.Context, accounts ...string) (map[string]GetAccountsBalanceRow, error) {
	accountsBalance := make(map[string]GetAccountsBalanceRow)
	if err := q.db.RunQuery(ctx, getAccountsBalancesWithChild, func(rows *postgres.RowsCompat) error {
		var (
			i GetAccountsBalanceRow
			// Below variables are ignored/omitted from the result.
			iMainAccountBalance  decimal.Decimal
			iChildAccountBalance decimal.Decimal
		)
		if err := rows.Scan(
			&i.AccountID,
			&i.AllowNegative,
			&i.Balance,
			&iMainAccountBalance,
			&iChildAccountBalance,
			&i.LastLedgerID,
			&i.LastMovementID,
			&i.CurrencyID,
			&i.CreatedAt,
		); err != nil {
			return err
		}
		if err := rows.Err(); err != nil {
			return err
		}
		accountsBalance[i.AccountID] = i
		return nil
	}, accounts); err != nil {
		return nil, err
	}
	return accountsBalance, nil
}
