package services

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"

	"github.com/studio-asd/pkg/resources"

	ledgerv1 "github.com/studio-asd/go-example/proto/api/ledger/v1"
	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	ledgerapi "github.com/studio-asd/go-example/services/ledger/api"
	userapi "github.com/studio-asd/go-example/services/user/api"
)

type Services struct {
	ledger *ledgerapi.API
	user   *userapi.API
	auth   *serviceAuth
}

func New(ledger *ledgerapi.API, user *userapi.API) *Services {
	return &Services{
		ledger: ledger,
		user:   user,
		auth: &serviceAuth{
			noAuthPatterns: map[string]string{
				// For debugging.
				"/v1/user/info": http.MethodGet,
			},
		},
	}
}

func (s *Services) RegisterAPIServices(grpcServer *resources.GRPCServerObject) error {
	// gRPC Gateway.
	err := grpcServer.RegisterGatewayService(func(gateway *resources.GRPCGatewayObject) error {
		gateway.RegisterMetadataHandler(metadataForwarder)
		gateway.RegisterMiddleware(s.middlewares()...)
		gateway.RegisterServiceHandler(func(mux *runtime.ServeMux) error {
			if err := ledgerv1.RegisterLedgerServiceHandlerServer(context.Background(), mux, s.ledger.GRPC()); err != nil {
				return err
			}
			if err := userv1.RegisterUserServiceHandlerServer(context.Background(), mux, s.user.GRPC()); err != nil {
				return err
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Services) middlewares() []runtime.Middleware {
	return []runtime.Middleware{
		s.auth.middleware,
	}
}

// metadataForwarder is always executed after middleware and before the actual handler is being executed. So we can always rely on this function
// to forward the metadata to the context.
func metadataForwarder(ctx context.Context, r *http.Request) metadata.MD {
	headers := map[string]string{
		"User-Agent":      r.Header.Get("User-Agent"),
		"Authorization":   r.Header.Get("Authorization"),
		"X-Forwarded-For": r.Header.Get("X-Forwarded-For"),
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(headers)
		return md
	}
	for k, v := range headers {
		md.Append(k, v)
	}
	return md
}
