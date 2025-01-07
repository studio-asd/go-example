package await

import (
	"context"
	"errors"
	"strings"
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

func TestWithTrace(t *testing.T) {
	opts := &options{}
	WithTrace("test")(opts)
	if opts.traceSpanName == "" {
		t.Fatal("trace span name is empty")
	}
	if !strings.HasSuffix(opts.traceSpanName, ".await") {
		t.Fatal("expecting .await suffix at the end of trace name")
	}
}
