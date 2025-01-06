// Code is generated by helper script. DO NOT EDIT.
// This code is generated to replace the SQLC main codes inside sqlc.go
// SQLC:
//   version    : v1.27.0
//   config     : sqlc.yaml
//   sql_package: pgx/v5
//   database   : go_example

package postgres

import (
	"context"
	"fmt"
	"database/sql"

	"github.com/studio-asd/pkg/postgres"
)

type Queries struct {
	db *postgres.Postgres
}

// New returns a new queries instance of go_example database.
func New(db *postgres.Postgres) *Queries {
	return &Queries{db: db}
}

// WithTransact wraps the queries inside a database transaction. The transaction will be committed if no error returned
// and automatically rolled back when an error occured.
func (q *Queries) WithTransact(ctx context.Context, iso sql.IsolationLevel, fn func(ctx context.Context, q *Queries) error) error {
	return q.db.Transact(ctx, iso, func(ctx context.Context, p *postgres.Postgres) error {
		return fn(ctx, New(p))
	})
}

// ensureInTransact ensures the queries are running inside the transaction scope, if the queries is not running inside the a transaction
// the function will trigger WithTransact method. While the function doesn't guarantee the subsequent function to have the same isolation
// level, but this function will return an error if the expectations and the current isolation level is incompatible.
func (q *Queries) ensureInTransact(ctx context.Context, iso sql.IsolationLevel, fn func(ctx context.Context, q *Queries) error) error {
	inTransaction, isoLevel := q.db.InTransaction()
	if !inTransaction {
		return q.WithTransact(ctx, iso, fn)
	}
	// Don't accept different isolation level between transactions as we will be getting different results.
	if iso != isoLevel {
		return fmt.Errorf("different expectations of isolation level. Got %s but expecting %s", isoLevel, iso)
	}
	return fn(ctx, q)
}

// Do executes queries inside the function fn and allowed other modules to execute queries inside the same transaction scope.
func (q *Queries) Do(ctx context.Context, fn func(ctx context.Context, pg *postgres.Postgres) error ) error {
	return fn(ctx, q.db)
}

// Postgres returns the postgres object.
func (q *Queries) Postgres() *postgres.Postgres {
	return q.db
}
