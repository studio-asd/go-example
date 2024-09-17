package errors

import (
	"log/slog"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"

	testdatav1 "github.com/albertwidi/go-example/proto/api/testdata/v1"
)

func TestProtoValidateErr(t *testing.T) {
	validator, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		message     proto.Message
		kind        Kind
		contraintID string
	}{
		{
			name:        "required",
			message:     &testdatav1.TestRequest{},
			kind:        KindBadRequest,
			contraintID: protovalidateViolationRequired,
		},
		{
			name: "email",
			message: &testdatav1.TestRequest{
				TestRequired: "required",
				TestEmail:    "not_an_email",
			},
			kind:        KindBadRequest,
			contraintID: protovalidateViolationEmail,
		},
		{
			name: "ip",
			message: &testdatav1.TestRequest{
				TestRequired: "required",
				TestEmail:    "email@gmail.com",
				TestIp:       "lalala",
			},
			kind:        KindBadRequest,
			contraintID: protovalidateViolationIP,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.Validate(test.message)
			if err != nil {
				errs := Wrap(err)
				if errs.constraintID != test.contraintID {
					t.Fatalf("expecting constraint id %s but got %s", test.contraintID, errs.constraintID)
				}
				if errs.Kind() != test.kind {
					t.Fatalf("expecting kind %s but got %s", test.kind, errs.Kind())
				}
			}
		})
	}
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
