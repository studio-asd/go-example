package api

import (
	"context"

	ledgerv1 "github.com/studio-asd/go-example/proto/api/ledger/v1"
)

var _ ledgerv1.LedgerServiceServer = (*GRPC)(nil)

// GRPC is the grpc server implementation of the API. The methods in the struct should only be invoked from the
// rpc framework as interceptor and other parts of the gRPC stacks won't be available via direct method call.
type GRPC struct {
	ledgerv1.UnimplementedLedgerServiceServer
	api *API
}

func newGRPC(api *API) *GRPC {
	return &GRPC{
		api: api,
	}
}

func (g *GRPC) Transact(context.Context, *ledgerv1.TransactRequest) (*ledgerv1.TransactResponse, error) {
	return nil, nil
}
