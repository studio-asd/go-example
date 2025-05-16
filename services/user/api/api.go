package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/studio-asd/pkg/postgres"
	"github.com/studio-asd/pkg/srun"
	"google.golang.org/protobuf/types/known/timestamppb"

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
			&userv1.LoginEmailPassword{},
			&userv1.AuthorizationRequest{},
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
	password, err := encryptUserPassword(req.Password, randSalt())
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	userUUID := uuid.NewString()
	_, err = a.queries.RegisterUserWithPassword(ctx, userpg.RegisterUserWithPassword{
		UUID:               uuid.NewString(),
		Email:              req.GetEmail(),
		Password:           string(password),
		PasswordSecretKey:  "user_password",
		PasswordSecretType: int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
		CreatedAt:          createdAt,
	})
	if err != nil {
		return nil, err
	}
	return &userv1.RegisterUserResponse{
		UserId:    userUUID,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

func (a *API) LoginRequest(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}
	switch req.GetLogin().(type) {
	case *userv1.LoginRequest_LoginPassword:
		req.ProtoReflect().Descriptor().FullName()
		return a.loginPassword(ctx, req.GetLoginPassword())
	default:
		return nil, nil
	}
}
