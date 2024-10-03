package await

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		do      DoFunc
		err     error
	}{
		{
			name: "no timeout",
			do: func(ctx context.Context) error {
				return nil
			},
			err: errTimeoutZero,
		},
		{
			name:    "with timeout, ok",
			timeout: time.Second,
			do: func(ctx context.Context) error {
				return nil
			},
			err: nil,
		},
		{
			name:    "with timeout, timeout/cancelled",
			timeout: time.Millisecond * 300,
			do: func(ctx context.Context) error {
				time.Sleep(time.Second)
				return nil
			},
			err: context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Do(context.Background(), test.timeout, test.do)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
		})
	}
}
