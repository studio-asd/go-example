package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
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
	// noAuthMethods stores the http patterns that don't require authentication. PLEASE be careful on adding more methods
	// here as we need to make sure that the method is really doesn't require authentication.
	//
	// In practice, we should write a description of each method why it doesn't require authentication.
	noAuthPatterns map[string]string
}

func New(ledger *ledgerapi.API, user *userapi.API) *Services {
	return &Services{
		ledger: ledger,
		user:   user,
		noAuthPatterns: map[string]string{
			// For debugging.
			"/v1/user/info": http.MethodGet,
		},
	}
}

func (s *Services) RegisterAPIServices(grpcServer *resources.GRPCServerObject) error {
	// gRPC Server.
	grpcServer.RegisterService(func(reg grpc.ServiceRegistrar) {
		ledgerv1.RegisterLedgerServiceServer(reg, s.ledger.GRPC())
		userv1.RegisterUserServiceServer(reg, s.user.GRPC())
	})
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
	authMD := func(hf runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			// Retrieve the http pattern name from the context. Since we are using grpc-gateway, the pattern is not available via r.Pattern
			// and both runtime.HTTPPathPattern and runtime.RPCMethod will reteurn an empty value.
			// Unfortunately the HTTP pattern is not available in grpc gateway generated code so we need to type it by our own.
			ptrn, ok := runtime.HTTPPattern(r.Context())
			// This means the request is coming from non-gateway handler, so we can't handle it.
			// We will immediately return internal server error(500) in this case.
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			httpPathPattern := ptrn.String()

			// Check whether we can go on with the request and whether the request need further authentication or not.
			// If the pattern and method is a match then we should go without authentication.
			httpMethod, ok := s.noAuthPatterns[httpPathPattern]
			if ok && httpMethod == r.Method {
				hf(w, r, pathParams)
				return
			}

			// switch r.Header.Get(headerClientType) {
			// case clientTypeWeb:
			// 	handleWebAuthentication(w, r)
			// case clientTypeMobile:
			// }
		}
	}

	return []runtime.Middleware{
		authMD,
	}
}

func metadataForwarder(ctx context.Context, r *http.Request) metadata.MD {
	ptrn, ok := runtime.HTTPPattern(r.Context())
	// This means the request is coming from non-gateway handler, so we can't handle it.
	// We will immediately return internal server error(500) in this case.
	if ok {
		fmt.Println("OK BOS", ptrn.String())
	}
	return nil
}
