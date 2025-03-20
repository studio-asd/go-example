package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/studio-asd/pkg/resources"

	ledgerv1 "github.com/studio-asd/go-example/proto/api/ledger/v1"
	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	ledgerapi "github.com/studio-asd/go-example/services/ledger/api"
	userapi "github.com/studio-asd/go-example/services/user/api"
)

type Services struct {
	ledger *ledgerapi.API
	user   *userapi.API
	// noAuthMethods stores the methods that don't require authentication. PLEASE be careful on adding more methods
	// here as we need to make sure that the method is really doesn't require authentication.
	//
	// In practice, we should write a description of each method why it doesn't require authentication.
	noAuthMethods map[string]struct{}
}

func New(ledger *ledgerapi.API, user *userapi.API) *Services {
	return &Services{
		ledger: ledger,
		user:   user,
		noAuthMethods: map[string]struct{}{
			userv1.UserService_Register_FullMethodName: {},
			userv1.UserService_Login_FullMethodName:    {},
		},
	}
}

func (s *Services) Register(grpcServer *resources.GRPCServerObject) error {
	err := grpcServer.RegisterGatewayService(func(gateway *resources.GRPCGatewayObject) error {
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
	return err
}

func (s *Services) middlewares() []runtime.Middleware {
	authMD := func(runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			fmt.Println("GOING HERE")
			// Retrieve the gRPC method name instead of the HTTP pattern because currently the grpc-gatway doesn't provide the constant variable of the http pattern.
			// While we can just copy the pattern and method, but it's better to use the method name directly as we already have it.
			method, ok := runtime.RPCMethod(r.Context())
			// This means the request is coming from non-gateway handler, so we can't handle it.
			// We will immediately return internal server error(500) in this case.
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if _, ok := s.noAuthMethods[method]; ok {
				return
			}

			switch r.Header.Get(headerClientType) {
			case clientTypeWeb:
				handleWebAuthentication(w, r)
			case clientTypeMobile:
			}
		}
	}

	return []runtime.Middleware{
		authMD,
	}
}
