package api

import (
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
	"strconv"
	"time"

	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	usertypev1 "github.com/studio-asd/go-example/proto/types/user/v1"
	"github.com/studio-asd/go-example/services"
	usersvc "github.com/studio-asd/go-example/services/user"
	userpg "github.com/studio-asd/go-example/services/user/internal/postgres"
	"github.com/studio-asd/pkg/postgres"
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

func (a *API) createLoginSession(ctx context.Context) (string, error) {
	md, err := services.NewGRPCMetadataRetriever(ctx)
	if err != nil {
		return "", err
	}
	userAgent := md.UserAgent()

	tokenCreatedAt := time.Now()
	// By default we create a session token that expires after an hour.
	tokenExpiredAt := tokenCreatedAt.Add(time.Hour)
	tokenRandomID := strconv.FormatInt(rand.Int64N(10), 10)
	// Create a session token and persist the session.
	sessionToken, err := encodeSessionToken(sessionTokenInfo{
		UserID:              user.ExternalID,
		RandomID:            tokenRandomID,
		CreataedAtTimestamp: tokenCreatedAt.UnixMilli(),
	})
	if err != nil {
		return "", err
	}
	// Create a token id as an identifier for the session.
	// While we don't have user-agent information in the token, we are calculating the user-agent in the id creation
	// to ensure only the specific user-agent can access the session. We don't include it directly to the token for
	// several reasons:
	//
	// 1. We don't want to make the token string longer than necessary.
	// 2. We don't want to include sensitive information in the token so that the client knows that information
	//    is used to identify the user.
	sessionID := generateSessionID(sessionIDParams{
		UserID:             user.ExternalID,
		RandomID:           tokenRandomID,
		CreatedAtTimestamp: tokenCreatedAt.UnixMilli(),
		UserAgent:          md.UserAgent(),
	})
	if err := a.queries.CreateUserSession(ctx, userpg.CreateUserSessionParams{
		SessionID:   sessionID,
		SessionType: int32(usertypev1.UserSessionType_USER_SESSION_TYPE_AUTHENTICATED),
		UserID: sql.NullInt64{
			Int64: user.UserID,
			Valid: true,
		},
		RandomID:             tokenRandomID,
		CreatedFromUserAgent: userAgent,
		CreatedAt:            tokenCreatedAt,
		ExpiredAt:            tokenExpiredAt,
	}); err != nil {
		return "", err
	}

	return sessionToken, nil
}
