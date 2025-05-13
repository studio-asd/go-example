package api

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeSessionToken(t *testing.T) {
	t.Parallel()
	if !testing.Short() {
		t.Skip()
	}

	createdAt := time.Now()

	tests := []struct {
		name         string
		genTokenFunc func(t *testing.T) string
		expect       sessionTokenInfo
		validErr     error
	}{
		{
			name: "valid token",
			genTokenFunc: func(t *testing.T) string {
				token, err := encodeSessionToken(sessionTokenInfo{
					UserID:              "1",
					RandomID:            "1234",
					CreataedAtTimestamp: createdAt.UnixMilli(),
				})
				if err != nil {
					t.Fatal(err)
				}
				return token
			},
			expect: sessionTokenInfo{
				UserID:              "1",
				RandomID:            "1234",
				CreataedAtTimestamp: createdAt.UnixMilli(),
			},
			validErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			token := test.genTokenFunc(t)
			info, err := decodeSessionToken(token)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expect, info); diff != "" {
				t.Fatalf("(-want/+got): %s", diff)
			}
			if err := info.valid(); err != test.validErr {
				t.Fatalf("expecting error %v but got %v", test.validErr, err)
			}
		})
	}
}
