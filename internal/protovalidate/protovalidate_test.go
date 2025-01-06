package protovalidate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"

	"github.com/studio-asd/go-example/internal/errors"
	testdatav1 "github.com/studio-asd/go-example/proto/testdata/protovalidate/v1"
)

func TestValidate(t *testing.T) {
	validator, err := New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		message proto.Message
		kind    errors.Kind
		fields  errors.Fields
	}{
		{
			name:    "required",
			message: &testdatav1.TestRequest{},
			kind:    errors.KindBadRequest,
			fields: errors.Fields{
				"protovalidate.constraint_id", "required",
				"protovalidate.field_path", "test_required",
			},
		},
		{
			name: "email",
			message: &testdatav1.TestRequest{
				TestRequired: "required",
				TestEmail:    "not_an_email",
			},
			kind: errors.KindBadRequest,
			fields: errors.Fields{
				"protovalidate.constraint_id", "validate.email",
				"protovalidate.field_path", "test_email",
			},
		},
		{
			name: "another_email",
			message: &testdatav1.TestRequest{
				TestRequired: "required",
				TestEmail:    "an@gmail.com",
			},
			kind: errors.KindBadRequest,
			fields: errors.Fields{
				"protovalidate.constraint_id", "validate.email",
				"protovalidate.field_path", "test_another_email",
			},
		},
		{
			name: "ip",
			message: &testdatav1.TestRequest{
				TestRequired:     "required",
				TestEmail:        "email@gmail.com",
				TestAnotherEmail: "another_email@gmail.com",
				TestIp:           "lalala",
			},
			kind: errors.KindBadRequest,
			fields: errors.Fields{
				"protovalidate.constraint_id", "validate.ip",
				"protovalidate.field_path", "test_ip",
			},
		},
		{
			name: "repeated_string",
			message: &testdatav1.TestRequest{
				TestRequired:     "required",
				TestEmail:        "email@gmail.com",
				TestAnotherEmail: "another_email@gmail.com",
				TestIp:           "127.0.0.1",
				RepeatedString:   []string{"this", "is", "a", "string"},
			},
			kind: errors.KindBadRequest,
			fields: errors.Fields{
				"protovalidate.constraint_id", "repeated.max_items",
				"protovalidate.field_path", "repeated_string",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.Validate(test.message)
			if err != nil {
				errs := err.(*errors.Errors)
				if diff := cmp.Diff(test.fields, errs.Fields()); diff != "" {
					t.Fatalf("(-want/+got)\n%s", diff)
				}
				if errs.Kind() != test.kind {
					t.Fatalf("expecting kind %s but got %s", test.kind, errs.Kind())
				}
			}
		})
	}
}
