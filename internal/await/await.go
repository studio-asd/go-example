// await is a simple wrapper to invoke a task inside a goroutine with a timeout.

package await

import (
	"context"
	"errors"
	"time"
)

var (
	errTimeoutZero = errors.New("timeout cannot be zero")
)

type DoFunc[T, V any] func(ctx context.Context, params T) (V, error)

// Do invokes a function in a goroutine with a certain timeout. Since the function wraps the parent context, it will respect
// the parent context cancellation as well.
func Do[T, V any](ctx context.Context, timeout time.Duration, params T, do DoFunc[T, V]) (V, error) {
	var emptyResult V
	if timeout == 0 {
		return emptyResult, errTimeoutZero
	}
	if do == nil {
		return emptyResult, nil
	}

	errC := make(chan error, 1)
	resultC := make(chan V, 1)
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	go func() {
		var err error
		var result V
		result, err = do(timeoutCtx, params)
		if err != nil {
			errC <- err
			return
		}
		resultC <- result
	}()

	select {
	case err := <-errC:
		return emptyResult, err
	case <-timeoutCtx.Done():
		return emptyResult, timeoutCtx.Err()
	case result := <-resultC:
		return result, nil
	}
}

func WithTracer() {

}
