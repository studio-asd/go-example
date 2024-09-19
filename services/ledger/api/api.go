package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/bufbuild/protovalidate-go"

	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
	"github.com/albertwidi/pkg/postgres"
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

type API struct{}

func New() *API {
	return &API{}
}

// Transact moves money from accounts to accounts within the transaction scope.
func (a *API) Transact(ctx context.Context, req *ledgerv1.TransactRequest, fn func(context.Context, *postgres.Postgres)) (*ledgerv1.TransactResponse, error) {
	if err := validator.Validate(req); err != nil {
		var validationErr *protovalidate.ValidationError
		if errors.As(err, &validationErr) {
			for _, violation := range validationErr.ToProto().Violations {
				fmt.Println(violation.ConstraintId)
				fmt.Println(violation.GetMessage())
			}
			return nil, err
		}
		return nil, err
	}
	return nil, nil
}
