package user

import (
	"errors"
	"testing"

	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	"github.com/studio-asd/go-example/services/user/api"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	th, err := testHelper.ForkPostgresSchema(t.Context(), testHelper.Postgres(), "public")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(th.Postgres().Config().DBName)
	ta := api.New(th.Postgres())

	tests := []struct {
		name string
		req  *userv1.RegisterUserRequest
		err  error
	}{
		{
			name: "normal registration",
			req: &userv1.RegisterUserRequest{
				Email:    "something@gmail.com",
				Password: "somethingtotest",
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ta.Register(t.Context(), tt.req); !errors.Is(err, tt.err) {
				t.Fatalf("expecting error %v but got %v", tt.err, err)
			}
		})
	}
}
