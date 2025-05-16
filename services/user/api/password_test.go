package api

import (
	"errors"
	"testing"

	"github.com/studio-asd/go-example/services/user"
)

func TestEncryptPassword(t *testing.T) {
	t.Parallel()

	if !testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name     string
		password string
		salt     string
		err      error
	}{
		{
			name:     "password too long",
			password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			salt:     randSalt(),
			err:      user.ErrPasswordTooLong,
		},
		{
			name:     "salt empty",
			password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			salt:     "",
			err:      user.ErrPasswordSaltEmpty,
		},
		{
			name:     "password too short",
			password: "aaa",
			salt:     randSalt(),
			err:      user.ErrPasswordTooShort,
		},
		{
			name:     "valid password",
			password: "aaaabbbb",
			salt:     randSalt(),
			err:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := encryptUserPassword(test.password, test.salt)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
		})
	}
}
