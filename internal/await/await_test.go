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
		do      DoFunc[any, any]
		err     error
	}{
		{
			name: "no timeout",
			do: func(ctx context.Context, params any) (any, error) {
				return nil, nil
			},
			err: errTimeoutZero,
		},
		{
			name:    "with timeout, ok",
			timeout: time.Second,
			do: func(ctx context.Context, params any) (any, error) {
				return nil, nil
			},
			err: nil,
		},
		{
			name:    "with timeout, timeout/cancelled",
			timeout: time.Millisecond * 300,
			do: func(ctx context.Context, params any) (any, error) {
				time.Sleep(time.Second)
				return nil, nil
			},
			err: context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Do(context.Background(), test.timeout, nil, test.do)
			if !errors.Is(err, test.err) {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}
		})
	}
}
