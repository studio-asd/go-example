package api

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

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

type sessionTokenInfo struct {
	// UserID is the user id that we use externally, this means it will be a UUIDv5.
	UserID              string
	RandomID            string
	CreataedAtTimestamp int64
}

func (s sessionTokenInfo) valid() error {
	if s.UserID == "" {
		return errors.New("session_token: user id is empty")
	}
	if s.RandomID == "" {
		return errors.New("session_token: random id is invalid")
	}
	t := time.UnixMilli(s.CreataedAtTimestamp)
	// Check wether the timestamp is valid.
	if t.IsZero() {
		return errors.New("session_token: created at timestamp is invalid")
	}
	// Check whether the timestamp is makes sense, our session is only valid for 1 hour, so it doesn't makes sense
	// to receive the session that was created 6 hours ago.
	if time.Since(t) > time.Hour*6 {
		return errors.New("session_token: created at timestamp is too old")
	}
	return nil
}

func (s sessionTokenInfo) toSessionString() string {
	return s.UserID + ":" + s.RandomID + ":" + strconv.FormatInt(s.CreataedAtTimestamp, 10)
}

// ecodeSessionToken encodes the session token information into a base64 string format. The session token is quite simple as it
// only consists of the user id, random number, and the timestamp when the session was created.
func encodeSessionToken(info sessionTokenInfo) (string, error) {
	if err := info.valid(); err != nil {
		return "", err
	}
	sessionToken := base64.RawStdEncoding.EncodeToString([]byte(info.toSessionString()))
	return sessionToken, nil
}

func decodeSessionToken(token string) (sessionTokenInfo, error) {
	strData, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return sessionTokenInfo{}, err
	}
	data := strings.Split(string(strData), ":")

	timeStamp, err := strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return sessionTokenInfo{}, err
	}

	info := sessionTokenInfo{
		UserID:              data[0],
		RandomID:            data[1],
		CreataedAtTimestamp: timeStamp,
	}
	return info, info.valid()
}

// sessionIDParams is the parameters needed to generate a session ID. The additional user agent is to ensure
// the new session ID is unique per user agent.
type sessionIDParams struct {
	UserID             string
	RandomID           string
	CreatedAtTimestamp int64
	UserAgent          string
}

// generateSessionID generates a session ID based on the paramters.
//
// These parameters are retrieved from the user's request:
// - UserID: The user's ID.
// - RandomID: A random id when creating the session.
// - CreatedAtTimestamp: The timestamp when the session was created.
// - UserAgent: The user's user agent from the request.
//
// This mean a session id is unique per user_id and user_agent. These parameters
// are enough for basic authorization but not strong enough for security.
func generateSessionID(gen sessionIDParams) uuid.UUID {
	timeStampStr := strconv.FormatInt(gen.CreatedAtTimestamp, 10)

	data := []byte(gen.UserID + ":" + gen.RandomID + ":" + timeStampStr + ":" + gen.UserAgent)
	return uuid.NewSHA1(uuid.NameSpaceOID, data)
}
