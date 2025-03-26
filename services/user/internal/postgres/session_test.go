package postgres

import "testing"

// TestGetUserSession retrieve the user session information including the user information.
func TestGetUserSession(t *testing.T) {
	t.Parallel()

	// Fork the schema so we don't mix the data between tests.
	th, err := testHelper.ForkPostgresSchema(t.Context(), testQueries.Postgres(), "public")
	if err != nil {
		t.Fatal(err)
	}
	_ = New(th.Postgres())

	// Prepare some user.
}
