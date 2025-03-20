package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/studio-asd/pkg/postgres"
)

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	tests := []struct {
		name     string
		register RegisterUser
		err      error
	}{
		{
			name: "a new user",
			register: RegisterUser{
				UUID:             "one",
				Email:            "testing@email.com",
				Password:         "a password",
				PasswordSecretID: "one",
				CreatedAt:        createdAt,
			},
			err: nil,
		},
		// This test should fail because we are using the same email as the first test.
		{
			name: "new user same email",
			register: RegisterUser{
				UUID:             "two",
				Email:            "testing@email.com",
				Password:         "a password",
				PasswordSecretID: "two",
				CreatedAt:        createdAt,
			},
			err: postgres.ErrUniqueViolation,
		},
		// This test should fail because the external id is the same as the first test.
		{
			name: "new user same external id",
			register: RegisterUser{
				UUID:             "one",
				Email:            "testing_2@email.com",
				Password:         "a password",
				PasswordSecretID: "two",
				CreatedAt:        createdAt,
			},
			err: postgres.ErrUniqueViolation,
		},
		{
			name: "new user different everything",
			register: RegisterUser{
				UUID:             "three",
				Email:            "testing_3@email.com",
				Password:         "a password",
				PasswordSecretID: "three",
				CreatedAt:        createdAt,
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
			err := tq.RegisterUser(t.Context(), test.register)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
		})
	}
}
