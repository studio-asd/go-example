package api

import (
	"context"
	"errors"
	"time"

	"github.com/studio-asd/pkg/postgres"

	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	"github.com/studio-asd/go-example/services"
	"github.com/studio-asd/go-example/services/user"
)

func (a *API) AuthenticateUser(ctx context.Context) {
}

func (a *API) AuthorizeUser(ctx context.Context, req *userv1.AuthorizationRequest) (*userv1.AuthorizationResponse, error) {
	// Retrieve the token information from the token.
	tokenInfo, err := decodeSessionToken(req.GetSessionToken())
	if err != nil {
		return nil, err
	}

	md, err := services.NewGRPCMetadataRetriever(ctx)
	if err != nil {
		return nil, err
	}

	// Generate the sessionID from the token information.
	sessionID := generateSessionID(sessionIDParams{
		UserID:             tokenInfo.UserID,
		RandomID:           tokenInfo.RandomID,
		CreatedAtTimestamp: tokenInfo.CreataedAtTimestamp,
		UserAgent:          md.UserAgent(),
	})
	// For now we will retrieve everything from the database, there will be a time when we need to cache the session token information.
	session, err := a.queries.GetUserSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, user.ErrUserSessionNotFound
		}
		return nil, err
	}
	// Check whether the user session has expired, we should not allow the user to access the resource if the session has expired.
	if time.Now().After(session.ExpiredAt) {
		return nil, user.ErrSessionExpired
	}

	return &userv1.AuthorizationResponse{
		UserId: tokenInfo.UserID,
	}, nil
}
