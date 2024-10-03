package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWrap(t *testing.T) {
	t.Run("error_is", func(t *testing.T) {
		var testErrType = errors.New("a standard error")

		tests := []struct {
			name   string
			err    error
			args   []any
			expect error
		}{
			{
				name:   "standard error",
				err:    testErrType,
				args:   nil,
				expect: testErrType,
			},
			{
				name:   "wrapped error with error args",
				err:    New("a wrapped error"),
				args:   []any{testErrType},
				expect: testErrType,
			},
			{
				name:   "wrapped error with join error(custom, standard)",
				err:    Join(New("a wrapped error"), testErrType),
				args:   nil,
				expect: testErrType,
			},
			{
				name:   "wrapped error with join fmt_error(custom, standard)",
				err:    Join(New("a wrapped error"), fmt.Errorf("error %w", testErrType)),
				args:   nil,
				expect: testErrType,
			},
			// This does not work as we are wrapping the standard error type to the custom one. We need to insert the
			// standard error into the custom error chain(somehow).
			// {
			// 	name:   "wrapped error with join error(standard, custom)",
			// 	err:    Join(testErrType, New("a wrapped error")),
			// 	args:   nil,
			// 	expect: testErrType,
			// },
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := Wrap(test.err, test.args...)
				if !Is(err, test.expect) {
					t.Fatalf("expecting error %v but got %v", test.expect, err)
				}
			})
		}
	})
}

func TestNewFields(t *testing.T) {
	tests := []struct {
		name   string
		kvs    []interface{}
		expect []interface{}
	}{
		{
			name: "simple KV",
			kvs: []interface{}{
				"a", "b",
			},
			expect: []interface{}{
				"a", "b",
			},
		},
		{
			name: "not a KV",
			kvs: []interface{}{
				"a", "b", "c",
			},
			expect: []interface{}{
				"a", "b",
				"unknown?", "c",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewFields(test.kvs...)

			// When diffing, we will force the 'got' type to any because 'got' type is errors.Fields.
			if diff := cmp.Diff(test.expect, []any(got)); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}
		})
	}
}

func TestFieldsToSlogAttributes(t *testing.T) {
	tests := []struct {
		name   string
		kvs    []any
		expect []slog.Attr
	}{
		{
			name: "simple kvs",
			kvs:  []any{"key", "value", "key2", 10, "key3", 10.4, "key4", true},
			expect: []slog.Attr{
				{
					Key:   "key",
					Value: slog.AnyValue("value"),
				},
				{
					Key:   "key2",
					Value: slog.AnyValue(10),
				},
				{
					Key:   "key3",
					Value: slog.AnyValue(10.4),
				},
				{
					Key:   "key4",
					Value: slog.AnyValue(true),
				},
			},
		},
		{
			name: "not valid kvs",
			kvs:  []any{"key", "value", "value2"},
			expect: []slog.Attr{
				{
					Key:   "key",
					Value: slog.AnyValue("value"),
				},
				{
					Key:   "unknown?",
					Value: slog.AnyValue("value2"),
				},
			},
		},
		{
			name: "key not string",
			kvs:  []any{"key", "value", 10, "value2"},
			expect: []slog.Attr{
				{
					Key:   "key",
					Value: slog.AnyValue("value"),
				},
				{
					Key:   "10",
					Value: slog.AnyValue("value2"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := Fields(test.kvs)
			attrs := f.ToSlogAttributes()

			for idx, e := range test.expect {
				if !e.Equal(attrs[idx]) {
					t.Fatalf("slog attributes are not equal for attr with key %s. Expect %v but got %v", e.Key, e, attrs[idx])
				}
			}
		})
	}
}
