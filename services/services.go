package services

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	ledgerv1 "github.com/studio-asd/go-example/proto/api/ledger/v1"
	ledgerapi "github.com/studio-asd/go-example/services/ledger/api"
	"github.com/studio-asd/pkg/resources"
)

type Services struct {
	ledger *ledgerapi.API
}

func New(ledger *ledgerapi.API) *Services {
	return &Services{
		ledger: ledger,
	}
}

func (s *Services) Register(grpcServer *resources.GRPCServerObject) error {
	err := grpcServer.RegisterGatewayService(func(gateway *resources.GRPCGatewayObject) error {
		gateway.RegisterServiceHandler(func(mux *runtime.ServeMux) error {
			if err := ledgerv1.RegisterLedgerServiceHandlerServer(context.Background(), mux, s.ledger.GRPC()); err != nil {
				return err
			}
			return nil
		})
		return nil
	})
	return err
}

func (s *Services) middlewares() []runtime.Middleware {
	authMD := func(runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			pattern, ok := runtime.HTTPPattern(r.Context())
			// This means the request is coming from non-gateway handler, so we can't handle it.
			// We will immediately return internal server error(500) in this case.
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	return []runtime.Middleware{
		authMD,
	}
}
