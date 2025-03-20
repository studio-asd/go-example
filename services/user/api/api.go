package api

import (
	"context"
	"log/slog"

	"github.com/studio-asd/pkg/postgres"
	"github.com/studio-asd/pkg/srun"

	"github.com/studio-asd/go-example/internal/protovalidate"
	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	userpg "github.com/studio-asd/go-example/services/user/internal/postgres"
)

var (
	validator *protovalidate.Validator
	_         srun.ServiceInitAware = (*API)(nil)
)

func init() {
	var err error
	validator, err = protovalidate.New(
		protovalidate.WithFailFast(),
		protovalidate.WithMessages(
			&userv1.RegisterUserRequest{},
			&userv1.LoginRequest{},
		),
	)
	if err != nil {
		panic(err)
	}
}

type API struct {
	queries *userpg.Queries
	logger  *slog.Logger
}

func New(pg *postgres.Postgres) *API {
	return &API{
		queries: userpg.New(pg),
	}
}

func (a *API) Name() string {
	return "user-api"
}

func (a *API) Init(ctx srun.Context) error {
	a.logger = ctx.Logger
	return nil
}

func (a *API) GRPC() *GRPC {
	return newGRPC(a)
}

func (a *API) RegisterUser(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
	return nil, nil
}

func (a *API) LoginRequest(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return nil, nil
}
