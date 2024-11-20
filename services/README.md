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
                |- internal_types.go
```

1. Services is the parent folder for all services.
2. Inside the services, we can create a service folder. For example, `service_a`.
3. Inside of the `service_a`, we expose all functions to the internal program via `api`.
4. The `service_a` should not expose its internal packages that private to the package. Thus `internal` folder will be used.
5. The data layer and internal types usually located inside the `internal` folder of a service to prevent direct usage by other packages.

### Internal Package

Go programming language has a concept called [internal package](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit?tab=t.0) release in
Go [1.4](https://go.dev/doc/go1.4#internalpackages). The internal package allowed the user to structure program with clean boundaries and disallow other
packages outside of the owner's subtree to import the package.

As we are building a program that contains several domains/services, it is important for each domain/services to not directly importing "internal" packages
such as `postgres` directly. As the data layer of a domain/service is only belong to that domain and should not be accessible directly from other services.
Additionally, we can also share a data structure that only belongs to that domain and exposes types that can be used externally outside of the internal package.

For example, this is possible:

```text
|- service_a
      |- api
          |- api.go -------------------|
      |- internal                      |
          |- postgres  <---------------| can import
                |- postgres.go
```

While this is not:

```text
|- service_a
      |- api
          |- api.go
      |- internal
          |- postgres <------------|
                |- postgres.go     |
|- service_b                       | cannot import
      |- api                       |
          |- api.go ---------------|
```

### Services Interaction

1. Direct

  A `direct` communication between services happen when a service calls a function/api of another service. Even though the behavior
  can be different from one APIs to another(as some APIs might do things asynchronously), but a direct function/api call still counted
  as a direct communication.

2. Indirect

  An `indirect` communication between services might happen via another service. For example `service_a` communicates with `service_b` that
  internally use `service_c` to do some stuffs. This might happen and is allowed. But please be mindful to as this creates more complex
  dependencies between services.

**Database Transaction Block**

A `service` able to open an API that directly interacts with `database transaction` scope. This allowed other services to ensure
data consistency between two services.

Usually, a service will open an API like this to ensure `transaction` can be used when the API/function is being called.

```go
type API struct {
	q *servicePg.Queries
}

func (a *API) FuncWithTxInside(ctx context.Context, req Request, fn func(context.Context, *postgres.Postgres) error) (Response, error) {
	// If the additional function scope is not nil, then we should invoke the function inside a time-bounded
	// goroutine as we don't know how much time the function will spent. So we need to ensure the function runs
	// inside the Transact SLA.
	if fn != nil {
		timeoutSLA := time.Second * 3
		// We use await internal package here to invoke a new time-bounded goroutine.
		err = await.Do(ctx, timeoutSLA, func(ctx context.Context) error {
			return a.queries.WithTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *servicePg.Queries) error {
				// Do something first within the *Queries.
				// ...

				// Do is a special function that exposes *postgres.Postgres. This means the function on the other side can do this:
				//
				// func DoSomethingInTxScope(ctx context.Context, pg *postgres.Postgres) error {
				//   	q := servicePg.New(pg)
				//		// Do something with the query object.
				// }
				if err := q.Do(ctx, fn); err != nil {
					return err
				}
				return nil
			})
		})
	} else {
		// Do something directly.
		// ...
	}
		return Response{}, nil
}
```

It might be more straightforward if we explain it with a picture:

```text
    |---------|
    |service_a|
    |---------|
         |
 call, passing "fn"
   as foreign_fn
         |
         v
|-----------------|
|    service_b    |
|-----------------|
|    tx_session   |
|-----------------|
| |-------------| |
| |  local_fn   | |
| | foreign_fn  | |
| |-------------| |
|-----------------|
```

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
