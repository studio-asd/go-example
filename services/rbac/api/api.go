package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/studio-asd/pkg/srun"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/studio-asd/go-example/internal/protovalidate"
	rbacv1 "github.com/studio-asd/go-example/proto/api/rbac/v1"
	rbacpg "github.com/studio-asd/go-example/services/rbac/internal/postgres"
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
			&rbacv1.CreateSecurityPermissionRequest{},
			&rbacv1.CreateSecurityRoleRequest{},
		),
	)
	if err != nil {
		panic(err)
	}
}

type API struct {
	queries *rbacpg.Queries
	logger  *slog.Logger
}

func New() {

}

func (a *API) Name() string {
	return "rbac-api"
}

func (a *API) Init(ctx srun.Context) error {
	a.logger = ctx.Logger
	return nil
}

func (a *API) CreatePermissions(ctx context.Context, req []*rbacv1.CreateSecurityPermissionRequest) ([]*rbacv1.CreateSecurityPermissionResponse, error) {
	createdAt := time.Now()

	insertParams := make([]rbacpg.SecurityPermission, len(req))
	for idx, r := range req {
		if err := validator.Validate(r); err != nil {
			return nil, err
		}
		insertParams[idx] = rbacpg.SecurityPermission{
			PermissionExternalID: uuid.NewString(),
			PermissionName:       r.PermissionName,
			PermissionType:       int32(r.PermissionType),
			PermissionKey:        r.PermissionKey,
			PermissionValue:      r.PermissionValue,
			CreatedAt:            createdAt,
		}
	}
	_, err := a.queries.CreatePermissions(ctx, insertParams)
	if err != nil {
		return nil, err
	}

	responses := make([]*rbacv1.CreateSecurityPermissionResponse, len(req))
	createdAtPb := timestamppb.New(createdAt)
	for idx, param := range insertParams {
		responses[idx] = &rbacv1.CreateSecurityPermissionResponse{
			PermissionId:   param.PermissionExternalID,
			PermissionName: param.PermissionName,
			PermissionType: req[idx].PermissionType,
			CreatedAt:      createdAtPb,
		}
	}
	return responses, nil
}

func (a *API) CreateRole(ctx context.Context, req *rbacv1.CreateSecurityRoleRequest) (*rbacv1.CreateSecurityRoleResponse, error) {
	return nil, nil
}
