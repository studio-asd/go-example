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

type DoFunc func(ctx context.Context) error

// Do invokes a function in a goroutine with a certain timeout. Since the function wraps the parent context, it will respect
// the parent context cancellation as well.
func Do(ctx context.Context, timeout time.Duration, do DoFunc) error {
	if timeout == 0 {
		return errTimeoutZero
	}
	if do == nil {
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	errC := make(chan error, 1)
	go func() {
		errC <- do(timeoutCtx)
	}()

	select {
	case err := <-errC:
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

func WithTracer() {

}
