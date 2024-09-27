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

```text
|--------|   |
|  API   |   |  Client Facing
|--------|   |
|  Data  |   |  Internal Package Facing
|--------|   v
```

Folder wise, we always structure them like this:

```text
-| services
      -| service_a
            |- api
                |- api.go
                |- api_test.go
            |- internal
                |- postgres
                      |- postgres.go
                      |- postgres_test.go
```

1. Services is the parent folder for all services.
2. Inside the services, we can create a service folder. For example, `service_a`.
3. Inside of the `service_a`, we expose all functions to the internal program via `api`.
4. The `service_a` should not expose its internal packages that private to the package. Thus `internal` folder will be used.
5. The data layer usually located inside the `internal` folder of a service to prevent direct usage by other packages.

### Monolith Services Interaction



### API Layer

API layer is used to expose application programming interface(API) to the client. The client can be another package or
an end user by using `grpc-gateway`.

The business logic is also placed inside the API layer as we don't want to create more layers(for now) that might make
us harder to continue build and test the program.

**Proto For API Interface**

You might asks on why we use `proto` as the interface to pass the data to the `api` layer. We are doing this because
we want to ensure we are providing exactly the same interface when we interact internally and also externally. Some of
our internal APIs/functions might also need to hit the exact same API that we expose to public, this consistency makes
it easier to do both things consistently.

### Data/Database Layer

The data/database layer is a layer that interacts directly with the storage system. The package is intended to be used
internally within the parent package and should not used directly by other packages.

In this layer, we expect less to none business logic involved as this will make separation of concern to be broken
between `api` and `data` layer.

## Error Handling
