# Services

Services directory stores all services inside the repository.

## What Is A Service?

A service provides a set of functions or features for a specific domain to be used by the customers. The customers
can be the end users or other packages/services inside the repository.

## Do A Service Act Like A Microservice?

Yes and no, it depends on the context and how the service behaves. Services that share the same database should be able to expose
some functions that wraps some functions inside the same database transaction, otherwise we need to treat the state of the data upon
failures.

> If there are no **really** good reason to separate the database, we should not do it.

### Example Of Using A Single Database Transaction Across Service/Module

You can look the example inside the [ledger api](./ledger/api/api.go). Please look at the `Transact` function.

## Protocol Buffers

We use [protobuf](https://protobuf.dev/) extensively inside of our [API Layer](#API Layer).

### gRPC

### Protovalidate

To validate our `protobuf` message, we use [protovalidate](https://github.com/bufbuild/protovalidate) to validates the fields of our message.

The validation is then tailored to our internal [errors](../internal/errors/README.md) package to wrap and handle errors from `protobuf` message
inside of our application.

For example:

```go
package api

import (
	"github.com/albertwidi/go-example/internal/errors"
	"github.com/albertwidi/go-example/internal/protovalidate"
	serviceapiv1 "github.com/albertwidi/go-examle/proto/api/service_api/v1"
)

var validator *protovalidate.Validator

func init() {
	var err error
	validator, err = protovalidate.New(
		// FailFast will be set to true because we don't want to waste time validating everything if
		// the first field already failing.
		protovalidate.WithFailFast(true),
		// WithMessage will put the message to the memory, so we have them pre-warmed thus leads to faster
		// validation.
		protovalidate.WithMessage(
			&serviceapiv1.FirstRequest{},
			&serviceapiv2.SecondRequest{},
		)
	)
	if err != nil {
		panic(err)
	}
}

type API struct {}

func (a *API) SomeAPI(ctx context.Context, req *serviceapiv1.FirstRequest) (*serviceapiv1.FirstResponse, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}
	// More codes...
}
```

## Layers & Structure

The layer


### API Layer

### Database Layer

## Error Handling
