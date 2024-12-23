package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/albertwidi/pkg/postgres"

	"github.com/albertwidi/go-example/internal/await"
)

type TransactFunc func(context.Context, *postgres.Postgres) error

// TransactExec is a helper to craete a transaction scope across modules.
type TransactExec struct {
	ctx context.Context
	pg  *postgres.Postgres
	iso sql.IsolationLevel
}

func NewTransactExec(ctx context.Context, pg *postgres.Postgres, iso sql.IsolationLevel) *TransactExec {
	return &TransactExec{
		ctx: ctx,
		pg:  pg,
		iso: iso,
	}
}

// Do triggers the transaction with function1 and function2 in parameters. The function2 also have additional timeout parameter because
// function1 will be used by the function/transaction inside the imported module.
func (txe *TransactExec) Do(fn1 TransactFunc, fn2 TransactFunc, fn2Timeout time.Duration) error {
	ok, _ := txe.pg.InTransaction()
	if ok {
		return errors.New("transactExec/Do: postgresql connection is already in transaction")
	}
	if fn2 == nil {
		return txe.pg.Transact(txe.ctx, txe.iso, func(ctx context.Context, pg *postgres.Postgres) error {
			return fn1(ctx, pg)
		})
	}

	return txe.pg.Transact(txe.ctx, txe.iso, func(ctx context.Context, pg *postgres.Postgres) error {
		if err := await.Do(ctx, fn2Timeout, func(doCtx context.Context) error {
			return fn2(doCtx, pg)
		}); err != nil {
			return err
		}
		return fn1(ctx, pg)
	})
}
