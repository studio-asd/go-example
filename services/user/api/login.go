package api

import (
	"context"
	"errors"

	"github.com/studio-asd/pkg/postgres"

	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	usertypev1 "github.com/studio-asd/go-example/proto/types/user/v1"
	usersvc "github.com/studio-asd/go-example/services/user"
	userpg "github.com/studio-asd/go-example/services/user/internal/postgres"
)

func (a *API) loginPassword(ctx context.Context, req *userv1.LoginEmailPassword) (*userv1.LoginResponse, error) {
	user, err := a.queries.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, usersvc.ErrUserNotFound
		}
	}
	value, err := a.queries.GetUserSecretValue(ctx, userpg.GetUserSecretValueParams{
		UserID:     user.UserID,
		SecretType: int32(usertypev1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
		SecretKey:  secretKeyUserPassword,
	})
	if err != nil {
		return nil, err
	}
	if !value.SecretSalt.Valid {
		return nil, errors.New("secret salt is not valid")
	}

	genPassword, err := encryptUserPassword(req.Password, value.SecretSalt.String)
	if err != nil {
		return nil, err
	}
	// Compare the generated password with the stored password
	if genPassword != password(value.SecretValue) {
		return nil, usersvc.ErrPasswordInvalid
	}

	sessionToken, err := a.createLoginSession(ctx, createLoginSessionRequest{
		userID:   user.UserID,
		userUUID: user.ExternalID,
	})
	if err != nil {
		return nil, err
	}

	return &userv1.LoginResponse{
		Token: sessionToken,
	}, nil
}
