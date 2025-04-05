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

	secretSalt := value.SecretSalt.String
	prefixSalt := secretSalt[0:len(secretSalt)]
	suffixSalt := secretSalt[len(secretSalt):]
	// The raw passwrod is generated through hashing the password with a salt and constructed in a specific way.
	// raw_password := prefixSalt + value.SecretValue + suffixSalt
	rawPassword := prefixSalt + req.Password + suffixSalt
	// Re-generate the password from the user parameter.
	genPassword, err := encryptUserPassword(rawPassword)
	if err != nil {
		return nil, err
	}
	// Compare the generated password with the stored password
	if genPassword != password(value.SecretValue) {
		return nil, usersvc.ErrInvalidPassword
	}

	sessionToken, err := a.createLoginSession(ctx)
	if err != nil {
		return nil, err
	}

	return &userv1.LoginResponse{
		Token: sessionToken,
	}, nil
}
