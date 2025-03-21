package api

import (
	"context"

	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type GRPC struct {
	userv1.UnimplementedUserServiceServer
	api *API
}

func newGRPC(api *API) *GRPC {
	return &GRPC{
		api: api,
	}
}

func (g *GRPC) Register(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}
	return nil, nil
}

func (a *GRPC) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}
	return nil, nil
}

func (a *GRPC) Info(ctx context.Context, e *emptypb.Empty) (*userv1.InfoResponse, error) {
	return &userv1.InfoResponse{
		UserId: "user",
	}, nil
}
