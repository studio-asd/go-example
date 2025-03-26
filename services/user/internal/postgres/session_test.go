package postgres

import (
	"database/sql"
	"net/netip"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	userv1 "github.com/studio-asd/go-example/proto/types/user/v1"
)

// TestGetUserSession retrieve the user session information including the user information.
func TestGetUserSession(t *testing.T) {
	t.Parallel()
	createdAt := time.Now()

	// Fork the schema so we don't mix the data between tests.
	th, err := testHelper.ForkPostgresSchema(t.Context(), testQueries.Postgres(), "public")
	if err != nil {
		t.Fatal(err)
	}
	tq := New(th.Postgres())

	// Prepare some users.
	usersRegister := []RegisterUser{
		{
			UUID:             "one",
			Email:            "one@email.com",
			Password:         "one",
			PasswordSecretID: "one_password",
			CreatedAt:        createdAt,
		},
		{
			UUID:             "two",
			Email:            "two@email.com",
			Password:         "two",
			PasswordSecretID: "two_password",
			CreatedAt:        createdAt,
		},
	}

	for _, user := range usersRegister {
		if err := tq.RegisterUser(t.Context(), user); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name          string
		session       CreateUserSessionParams
		expectSession GetUserSessionRow
	}{
		{
			name: "guest session",
			session: CreateUserSessionParams{
				SessionID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				SessionType:          int32(userv1.UserSessionType_USER_SESSION_TYPE_GUEST),
				RandomID:             "random_1",
				CreatedFromIp:        netip.MustParseAddr("127.0.0.1"),
				CreatedFromUserAgent: "testing",
				SessionMetadata:      []byte(`{"key": "value"}`),
				CreatedAt:            createdAt,
				ExpiredAt:            createdAt,
			},
			expectSession: GetUserSessionRow{
				SessionID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				PreviousSesisionID:   uuid.NullUUID{},
				UserID:               sql.NullInt64{},
				Email:                sql.NullString{},
				SessionType:          int32(userv1.UserSessionType_USER_SESSION_TYPE_GUEST),
				RandomID:             "random_1",
				CreatedFromIp:        netip.MustParseAddr("127.0.0.1"),
				CreatedFromUserAgent: "testing",
				SessionMetadata:      []byte(`{"key": "value"}`),
				CreatedAt:            createdAt,
				ExpiredAt:            createdAt,
			},
		},
		{
			name: "authenticated session one",
			session: CreateUserSessionParams{
				SessionID:   uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				SessionType: int32(userv1.UserSessionType_USER_SESSION_TYPE_AUTHENTICATED),
				UserID: sql.NullInt64{
					Int64: 1,
					Valid: true,
				},
				RandomID:             "random_2",
				CreatedFromIp:        netip.MustParseAddr("127.0.0.1"),
				CreatedFromUserAgent: "testing",
				SessionMetadata:      []byte(`{"key": "value"}`),
				CreatedAt:            createdAt,
				ExpiredAt:            createdAt,
			},
			expectSession: GetUserSessionRow{
				SessionID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				UserID: sql.NullInt64{
					Int64: 1,
					Valid: true,
				},
				Email: sql.NullString{
					String: "one@email.com",
					Valid:  true,
				},
				SessionType:          int32(userv1.UserSessionType_USER_SESSION_TYPE_AUTHENTICATED),
				RandomID:             "random_2",
				CreatedFromIp:        netip.MustParseAddr("127.0.0.1"),
				CreatedFromUserAgent: "testing",
				SessionMetadata:      []byte(`{"key": "value"}`),
				CreatedAt:            createdAt,
				ExpiredAt:            createdAt,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := tq.CreateUserSession(t.Context(), test.session); err != nil {
				t.Fatal(err)
			}
			session, err := tq.GetUserSession(t.Context(), test.session.SessionID)
			if err != nil {
				t.Fatal(err)
			}

			equateComp := cmpopts.EquateComparable(netip.Addr{})
			if diff := cmp.Diff(test.expectSession, session, equateComp); diff != "" {
				t.Fatalf("user_session (-want/+got)\n%s", diff)
			}
		})
	}
}
