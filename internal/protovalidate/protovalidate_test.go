package protovalidate

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/albertwidi/go-example/internal/errors"
	testdatav1 "github.com/albertwidi/go-example/proto/api/testdata/v1"
	"github.com/google/go-cmp/cmp"
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
			name: "ip",
			message: &testdatav1.TestRequest{
				TestRequired: "required",
				TestEmail:    "email@gmail.com",
				TestIp:       "lalala",
			},
			kind: errors.KindBadRequest,
			fields: errors.Fields{
				"protovalidate.constraint_id", "validate.ip",
				"protovalidate.field_path", "test_ip",
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
