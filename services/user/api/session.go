package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/studio-asd/go-example/services"
)

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
