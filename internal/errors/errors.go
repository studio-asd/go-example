package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindBadRequest
	KindUnauthorized
	KindInternalError
)

func (k Kind) HTTPCode() int {
	switch k {
	case KindBadRequest:
		return http.StatusBadRequest
	case KindInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (k Kind) String() string {
	switch k {
	case KindBadRequest:
		return "Bad Request"
	case KindInternalError:
		return "Internal Error"
	default:
		return "Internal Error"
	}
}

type Errors struct {
	err    error
	kind   Kind
	fields Fields
	// constraintID is protovalidate constraint_id.
	constraintID string
}

func (e *Errors) Error() string {
	return e.err.Error()
}

func (e *Errors) Kind() Kind {
	return e.kind
}

func (e *Errors) Fields() Fields {
	return e.fields
}

// New creates a completely new error.
func New(s string, v ...any) *Errors {
	e := errors.New(s)
	return Wrap(e, v...)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

// Wrap an error.
func Wrap(err error, v ...any) *Errors {
	// Don't return a wrapped error if the error is nil, we should return nil to ensure the usual
	// error checking flow working as intended.
	if err == nil {
		return nil
	}
	var e *Errors

	// Check whether the errors is already an *Errors. Use the previous *Errors if possible and append
	// or override the values.
	var internalErrs *Errors
	if As(err, &internalErrs) {
		e = internalErrs
	} else {
		e = &Errors{err: err}
	}

	for idx := range v {
		switch t := v[idx].(type) {
		case error:
			e.err = errors.Join(e.err, t)
		case Kind:
			if e.kind == KindUnknown && t != KindUnknown {
				e.kind = t
			}
		case Fields:
			if len(e.fields) > 0 {
				e.fields = append(e.fields, t)
			} else {
				e.fields = t
			}
		}
	}
	return e
}

// Join wraps multiple error into one error. This functionality is added in Go 1.20.
func Join(err error, v ...error) error {
	if len(v) == 0 {
		return err
	}

	// If we have the internal error type here, we should wrap the internal error with the incoming errors.
	var errs *Errors
	if errors.As(err, &errs) {
		v = append([]error{errs.err}, v...)
		errs.err = errors.Join(v...)
		return errs
	}
	v = append([]error{err}, v...)
	return errors.Join(v...)
}

// Is shadows the errors.Is function to check whether the internal error type is the same
// with the one we want to compare.
func (e *Errors) Is(err error) bool {
	return errors.Is(e.err, err)
}

// As shadows the errors.As function to check whether the internal error type is the same
// with the one we want to compare.
func (e *Errors) As(target any) bool {
	return errors.As(e.err, target)
}

type Fields []any

// NewFields is for safely creating error fields becauase the error fields format
// is a key value to add more context to the error.
func NewFields(kv ...any) (f Fields) {
	if kv == nil {
		return nil
	}

	kvlen := len(kv)
	if kvlen%2 == 0 {
		f = kv
		return
	}

	// Ensure that the Fields is never 'odd'. If they 'key' is not available then we
	// should replace the 'key' with 'unknown?'.
	newKV := make([]interface{}, kvlen+1)
	for i := 0; i < kvlen; i++ {
		if i == kvlen-1 {
			newKV[i] = "unknown?"
			// We will always know this is safe to do because we previously
			// set the array capacity to kv length + 1.
			newKV[i+1] = kv[i]
			break
		}
		newKV[i] = kv[i]
	}

	f = newKV
	return
}

// ToSlogAttributes safely converts fields([]any) to slog.Attr. The function will
// return an empty []slog.Attr if the fields is empty.
//
// Please note that []Fields are expected to form a []Field{"string", "value"} where
// the first key is always a 'string' and 'any' for the value. If the key is not a
// 'string', the function will convert the key to 'string' to be compatible with the
// slog.Attr standard.
func (f Fields) ToSlogAttributes() []slog.Attr {
	kvlen := len(f)
	// We should not convert the fields because it doesn't fulfill the
	// fields k/v criteria. Possibly, the fields is created without
	// using the NewFields function.
	if kvlen%2 != 0 {
		f = NewFields(f...)
		kvlen = len(f)
	}

	attrs := make([]slog.Attr, kvlen/2)
	for i := 0; i < kvlen; i += 2 {
		var slogKey string
		key, isString := f[i].(string)
		// If the key is not string then we will convert the key using fmt.Sprinf to a string.
		// This will allocates but only if the key is not string, so we think the trade-off
		// is worth-it.
		if !isString {
			// We use f[i] here because the value of 'key' will be empty if the field is not
			// a string.
			slogKey = fmt.Sprintf("%v", f[i])
		} else {
			slogKey = key
		}

		// The counter for attribute array need to be divided by 2 because we always increase
		// i by 2.
		attrs[i/2] = slog.Attr{
			Key:   slogKey,
			Value: slog.AnyValue(f[i+1]),
		}
		if i+2 == kvlen {
			break
		}
	}
	return attrs
}
