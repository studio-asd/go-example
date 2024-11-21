package protovalidate

import (
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"

	"github.com/albertwidi/go-example/internal/errors"
)

// DefaultConstraintsMap defines the default map for constraint_id from protovalidate. The map is based
// on the internal standard so we can map them further to our internal errors and erorrs.Kind.
var DefaultConstraintsMap = map[string]errors.Kind{
	"required":           errors.KindBadRequest,
	"validate.email":     errors.KindBadRequest,
	"validate.ip":        errors.KindBadRequest,
	"repeated.max_items": errors.KindBadRequest,
	"repeated.min_items": errors.KindBadRequest,
}

type Validator struct {
	validate       *protovalidate.Validator
	constraintsMap map[string]errors.Kind
}

// New creates a thin wrapper of protovalidate.Validator. The wrapper overrides the Validate function to
// ensure it can give a rich error context.
func New(opts ...protovalidate.ValidatorOption) (*Validator, error) {
	validator, err := protovalidate.New(opts...)
	if err != nil {
		return nil, err
	}
	return &Validator{
		validate:       validator,
		constraintsMap: DefaultConstraintsMap,
	}, nil
}

// SetConstraintsMapping sets the constraints map to produce the desired errors.Kind based on the mapping.
// Please NOTE that changing the mapping is not concurrently safe, so you need to set the value upfront.
func (v *Validator) SetConstraintsMapping(m map[string]errors.Kind) {
	v.constraintsMap = m
}

// Validate returns a custom error from protovalidate. The error returned will no-longer be protovalidate.ValidationError
// as we will return errors.Error as our custom internal error. The custom error gives rich context for the error that
// hopefully helps the user to understand the error better.
func (v *Validator) Validate(message proto.Message) error {
	var validateErr *protovalidate.ValidationError
	err := v.validate.Validate(message)
	if err == nil {
		return nil
	}
	if !errors.As(err, &validateErr) {
		return err
	}
	if len(validateErr.Violations) < 0 {
		return err
	}
	// We expect that failfast is being used, as we will only retrieve the first error.
	violation := validateErr.Violations[0]
	kind, ok := v.constraintsMap[violation.ConstraintId]
	if !ok {
		kind = errors.KindUnknown
	}
	// The errorMessage to the user will be "violation_message for vioation_field_path". The field path will be quite verbose
	// for the user, but we think it should be fine.
	errorMessage := violation.GetMessage() + " for " + violation.GetFieldPath()
	return errors.New(
		errorMessage,
		kind,
		errors.Fields{
			"protovalidate.constraint_id", violation.GetConstraintId(),
			"protovalidate.field_path", violation.GetFieldPath(),
		},
	)
}

func WithFailFast(failFast bool) protovalidate.ValidatorOption {
	return protovalidate.WithFailFast(failFast)
}

func WithMessages(messages ...proto.Message) protovalidate.ValidatorOption {
	return protovalidate.WithMessages(messages...)
}
