package api

import (
	"context"

	"github.com/albertwidi/pkg/postgres"

	"github.com/albertwidi/go-example/internal/protovalidate"
	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
	ledgerpg "github.com/albertwidi/go-example/services/ledger/internal/postgres"
)

var validator *protovalidate.Validator

func init() {
	var err error
	validator, err = protovalidate.New(
		protovalidate.WithFailFast(true),
		protovalidate.WithMessages(
			&ledgerv1.TransactRequest{},
		),
	)
	if err != nil {
		panic(err)
	}
}

type API struct {
	queries *ledgerpg.Queries
}

func New(queries *ledgerpg.Queries) *API {
	return &API{
		queries: queries,
	}
}

// Transact moves money from accounts to accounts within the transaction scope.
func (a *API) Transact(ctx context.Context, req *ledgerv1.TransactRequest, fn func(context.Context, *postgres.Postgres)) (*ledgerv1.TransactResponse, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}
	return nil, nil
}
