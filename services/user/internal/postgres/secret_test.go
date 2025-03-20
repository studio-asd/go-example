package postgres

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/studio-asd/pkg/postgres"
)

// TestCreateNewSecret runs in parallel with the other tests but internally it runs sequentially.
// This is because in the test we need a deterministic order of secret_id.
func TestCreateNewSecret(t *testing.T) {
	t.Parallel()
	createdAt := time.Now()

	tests := []struct {
		name   string
		secret CreateNewSecret
		expect GetUserSecretValueRow
		err    error
	}{
		{
			name: "create new secret",
			secret: CreateNewSecret{
				ExternalID: "one",
				UserID:     1,
				Key:        "user_password",
				Value:      "a password",
				Type:       1,
				CreatedAt:  createdAt,
			},
			expect: GetUserSecretValueRow{
				SecretID:             1,
				ExternalID:           "one",
				UserID:               1,
				SecretKey:            "user_password",
				SecretType:           1,
				CurrentSecretVersion: 1,
				CreatedAt:            createdAt,
				UpdatedAt:            sql.NullTime{},
				SecretValue:          "a password",
			},
			err: nil,
		},
		{
			name: "same secret, different value",
			secret: CreateNewSecret{
				ExternalID: "one",
				UserID:     1,
				Key:        "user_password",
				Value:      "a duplicate password",
				Type:       1,
				CreatedAt:  createdAt,
			},
			err: postgres.ErrUniqueViolation,
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
			err := tq.CreateNewSecret(t.Context(), test.secret)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
			if err != nil {
				return
			}
			got, err := tq.GetUserSecretValue(t.Context(), GetUserSecretValueParams{
				UserID:     test.expect.UserID,
				SecretKey:  test.expect.SecretKey,
				SecretType: test.expect.SecretType,
			})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expect, got); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
