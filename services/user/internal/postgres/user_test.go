package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/studio-asd/pkg/postgres"

	userv1 "github.com/studio-asd/go-example/proto/types/user/v1"
)

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	createdAt := time.Now()
	tests := []struct {
		name                  string
		register              RegisterUserWithPassword
		expectUserTable       User
		expectUserPIITable    UserPii
		expectUserSecretTable GetUserSecretByExternalIDRow
		err                   error
	}{
		{
			name: "a new user",
			register: RegisterUserWithPassword{
				UUID:               "one",
				Email:              "testing@email.com",
				Password:           "a password",
				PasswordSecretKey:  "user_password",
				PasswordSecretType: int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
				CreatedAt:          createdAt,
			},
			expectUserTable: User{
				UserID:     1,
				ExternalID: "one",
				CreatedAt:  createdAt,
			},
			expectUserPIITable: UserPii{
				UserID:    1,
				Email:     "testing@email.com",
				CreatedAt: createdAt,
			},
			expectUserSecretTable: GetUserSecretByExternalIDRow{
				SecretID:             1,
				ExternalID:           "one",
				UserID:               1,
				SecretKey:            "user_password",
				SecretType:           int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
				SecretValue:          "a password",
				CurrentSecretVersion: 1,
				CreatedAt:            createdAt,
			},
			err: nil,
		},
		// This test should fail because we are using the same email as the first test.
		{
			name: "new user same email",
			register: RegisterUserWithPassword{
				UUID:               "two",
				Email:              "testing@email.com",
				Password:           "a password",
				PasswordSecretKey:  "one",
				PasswordSecretType: int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
				CreatedAt:          createdAt,
			},
			err: postgres.ErrUniqueViolation,
		},
		// This test should fail because the external id is the same as the first test.
		{
			name: "new user same external id",
			register: RegisterUserWithPassword{
				UUID:               "one",
				Email:              "testing_2@email.com",
				Password:           "a password",
				PasswordSecretKey:  "user_password",
				PasswordSecretType: int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
				CreatedAt:          createdAt,
			},
			err: postgres.ErrUniqueViolation,
		},
		{
			name: "new user different everything",
			register: RegisterUserWithPassword{
				UUID:               "three",
				Email:              "testing_3@email.com",
				Password:           "a password",
				PasswordSecretKey:  "user_password",
				PasswordSecretType: int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
				CreatedAt:          createdAt,
			},
			expectUserTable: User{
				UserID:     4,
				ExternalID: "three",
				CreatedAt:  createdAt,
			},
			expectUserPIITable: UserPii{
				UserID:    4,
				Email:     "testing_3@email.com",
				CreatedAt: createdAt,
			},
			expectUserSecretTable: GetUserSecretByExternalIDRow{
				SecretID:             2,
				ExternalID:           "three",
				UserID:               4,
				SecretKey:            "user_password",
				SecretType:           int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
				SecretValue:          "a password",
				CurrentSecretVersion: 1,
				CreatedAt:            createdAt,
			},
			err: nil,
		},
	}

	// Fork the schema so we don't mix the data between tests.
	th, err := testHelper.ForkPostgresSchema(t.Context(), testQueries.Postgres(), "public")
	if err != nil {
		t.Fatal(err)
	}
	tq := New(th.Postgres())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := tq.RegisterUserWithPassword(t.Context(), test.register)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
			if err != nil {
				return
			}
			user, err := tq.GetUserByExternalID(t.Context(), test.register.UUID)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectUserTable, user); diff != "" {
				t.Fatalf("user_table (-want/+got)\n%s", diff)
			}
			userPII, err := tq.GetUserPII(t.Context(), user.UserID)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectUserPIITable, userPII); diff != "" {
				t.Fatalf("user_pii_table (-want/+got)\n%s", diff)
			}
			userSecret, err := tq.GetUserSecretByExternalID(t.Context(), test.register.UUID)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expectUserSecretTable, userSecret); diff != "" {
				t.Fatalf("user_secret_table (-want/+got)\n%s", diff)
			}
		})
	}
}
