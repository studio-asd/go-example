// await is a simple wrapper to invoke a task inside a goroutine with a timeout.

package await

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

var (
	errTimeoutZero = errors.New("timeout cannot be zero")
	tracer         trace.Tracer
)

func init() {
	tracer = tracenoop.NewTracerProvider().Tracer("noop")
}

type options struct {
	traceSpanName string
}

type InstrumentationConfig struct {
	Tracer trace.Tracer
}

// SetInstrumentation sets the insturmentation object like tracer and meter.
//
// This function is NOT concurrently safe, so you can set the configuration when the program start to avoid race
// condition.
func SetInstrumentation(config InstrumentationConfig) {
	if config.Tracer != nil {
		tracer = config.Tracer
	}
}

type DoFunc[T, V any] func(ctx context.Context, params T) (V, error)

// Do invokes a function in a goroutine with a certain timeout. Since the function wraps the parent context, it will respect
// the parent context cancellation as well.
func Do[T, V any](ctx context.Context, timeout time.Duration, params T, do DoFunc[T, V], opts ...func(*options) error) (result V, returnedErr error) {
	if timeout == 0 {
		returnedErr = errTimeoutZero
		return
	}
	if do == nil {
		return
	}

	o := &options{}
	for _, fn := range opts {
		returnedErr = fn(o)
		if returnedErr != nil {
			return
		}
	}

	errC := make(chan error, 1)
	resultC := make(chan V, 1)
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		spanCtx = timeoutCtx
		span    trace.Span
	)
	if o.traceSpanName != "" {
		spanCtx, span = tracer.Start(timeoutCtx, o.traceSpanName)
		defer func() {
			if returnedErr != nil {
				span.SetStatus(codes.Error, returnedErr.Error())
			}
			span.End()
		}()
	}

	// Please be aware that goroutines are scheduled after the timeout context are being created. This means it is closer to the timeout
	// than the real timer of timeout context. The client also need to be aware of the scheduling "delay" of goroutines inside the go runtime.
	// If the runtime is busy and the goroutines cannot be scheduled as soon as possible, it will also decrease the run duration of the
	// DoFunc.
	go func() {
		var err error
		var result V
		result, err = do(spanCtx, params)
		if err != nil {
			errC <- err
			return
		}
		resultC <- result
	}()

	select {
	case returnedErr = <-errC:
		return
	case <-timeoutCtx.Done():
		returnedErr = timeoutCtx.Err()
		return
	case result = <-resultC:
		return
	}
}

func WithTrace(spanName string) func(*options) error {
	return func(opts *options) error {
		opts.traceSpanName = spanName + ".await"
		return nil
	}
}
