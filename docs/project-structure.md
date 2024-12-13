# Go Example Project Structure

Welcome to the first section of the go-example doc. In this section we will talk about what is the thought process behind this repository project structure.

Please note that this project never intend to standarize the Go project structure. In my opinion, the project structure should be idiomatic for the intention of the project,
and there is no silver bullet for the project structure.

## Background

The go-example project is intended to show the perspective of Go developer on designing software for [industrial programming](https://peter.bourgon.org/go-for-industrial-programming/).
In the Peter Bourgon's blog, he describes the programming in industrial context as:

1. In a startup or corporate environment.
1. Within a team where engineers come and go.
1. On code that outlives any single engineer.
1. Serving highly mutable business requirements.

This topic is brought up because of it is where the writer have been spending most of his career at and accumulates some of the knowledge on how to deal with the various projects and team sizes.

Every startup, project and team starts small at first, and then becoming bigger over time. So we want to focus on something really matters first, delivering the project/software so customers can use it. And because of that, the **monolithic** software is written for the example to ensure the software is straightforward to build, test and deliver.

## The Structure

### Domain Separation

As being said earlier, the project is about a monolithic server that serves HTTP APIs for the customers/client to use. While being monolithic, the project is structured in a way that every obvious domain is separated with each other.

What does it mean by separated domain? A user facing software have features that can be used by the customers, and some software usually have more than one feature. Or, something that being seen as one feature from the user perspective possibly can be supported by other features to run.

```text
|-----------------------|
| E-Commerce            |
|-----------------------|
| |-------------------| |
| |      Payment      | |
| |-------------------| |
|                       |
| |-------------------| |
| |       Order       | |
| |-------------------| |
|                       |
| |-------------------| |
| |     Logistic      | |
| |-------------------| |
|                       |
| |-------------------| |
| |      Wallet       | |
| |-------------------| |
|                       |
| |-------------------| |
| |        ...        | |
| |-------------------| |
|                       |
|-----------------------|
```

In the example above, for `e-commerce` there are a lot of features they need to build to ensure the items bought by the customers can be delivered safely to the customer's house. In this case, the cusomer's need to choose the item they want to buy, choose the logistic and destination for the shipment and pay the order.

```text
|---------------------|
|  Buy Flow           |
|---------------------|
| |------------|      |
| | Put Orders |      |
| |------------|      |
|       |             |
|       v             |
| |-----------------| |
| | Choose Logistic | |
| |-----------------| |
|       |             |
|       v             |
| |---------|         |
| | Invoice |         |
| |---------|         |
|       |             |
|       v             |
| |---------|         |
| | Payment |         |
| |---------|         |
|---------------------|
```

All of these are separate feature `domain` that need to be maintained by the `e-commerce`. And each of them is called a `domain` because they are doing a completely different things from each domain perspective. If you have ever read the single responsibility principle([SRP](https://en.wikipedia.org/wiki/Single-responsibility_principle)), this is the same concept the we applied here.

1. Order domain is responsible to process customer's order and turn them into a payable invoice.
2. Logistic domain is reponsible to provide an information about shipping and how the customer's item can be delivered into the destination.
3. Payment domain is responsible to provide a payment options and execution for the customer's so they able to pay the invoice of their orders.

From the `buy flow` above, we learn that there are ordered sequence of events executed by different-different domains. This means the domain need to communicate with each other for them to be able to reach the end of the event from the example(payment).

```text
  E-Commerce System                                     External Parties
|-----------------------------------------------|
| |-------|  1. retrieve_logistics |----------| |    |--------------------|
| | Order |----------------------> | Logistic |----->| Logistics Provider |
| |-------|                        |----------| |    |--------------------|
|    | 2. generate |---------|                  |
|    |-----------> | Invoice |<--|              |
|                  |---------|   |  3. pay      |
|                                |              |
|                             |---------|       |    |---------------------|
|                             | Payment |----------->| Banks/Other Payment |
|                             |---------|       |    |     Channels        |
|-----------------------------------------------|    |---------------------|
```

Ok, now how all of those explanations related to what we have in this project? And how those things are reflected? As mentioned above, as we care about each domain responsibility we are trying to build a project structure to ensure those responsibility are still separated but able to communicate with each others. For this reason, each domain are located inside the `services` folder. As you can see in the folder, there are several domain being created there.

```text
|- services
      |- ledger
           |- api
		   |- internal
	  |- wallet
	       |- api
		   |- internal
```

The `ledger` and `wallet` domain are separated domain with different functionalities and responsibility. And because these two domains are exists within one program, they can talk to each other via function calls. But when doing this, each domain need to ensure they are not leaking the implementation detail or the internals of each domain.

Let's create a real example of what exists inside the `go-example` program.

```text
|-------------------------------------------|
| |----------------------------------|      |
| |  |--------|        |----------|  |      |
| |  | Ledger |------> | Postgres |  |      |
| |  |--------|        |----------|  |      |
| |      ^  Ledger Domain   ^        |      |
| |------|------------------|--------|      |
|        |                  |               |
|        | call             |               |
|        |                  X Not Allowed   |
|    |--------|             |               |
|    | Wallet | ------------|               |
|    |--------|                             |
|-------------------------------------------|
```

In this example, `ledger` as a domain has a PostgreSQL database depdeency, and its up to the `ledger` domain on how to use the database to fullfil it needs. Because the PostgreSQL is the dependency of `ledger`, only `ledger` is authorized to write to its own database. Other domain like `wallet` should not be authorized to write to the same database as it will violates the access and responsibility of `ledger` domain.

Some of you might be familiar with the concept of micro service and all of these domain separation and single responsibility is aligned with the concept of micro service. A monolithic software is indeed crafted with micro service inside the software itself, it just they are not communicating via network, they use function as their APIs.

So, how does the `domain` itself protected its own internal implementation inside the `services` structure? If you look again at the structure:

```text
|- services
      |- ledger
           |- api
               |- internal <- We have this.
```

You will find there is an `internal` folder inside each domain/service. Because in Go packages are represented by directory, we can create an internal directory/package in our structure to protect the implementation from another domain/service/package. This capability is added in Go [1.4](https://go.dev/doc/go1.4#internalpackages). By using internal directory, the Go toolchain will not allow another domain/service/package to import the internal implementation of other domain. For example:

This is possible:

```text
|- ledger
      |- api
          |- api.go --------------|
      |- internal                 | can import
          |- postgres  <----------|
                |- postgres.go
```

While this is not:

```text
|- ledger
      |- api
          |- api.go
      |- internal
          |- postgres <------------|
                |- postgres.go     |
|- wallet                          | cannot import
      |- api                       |
          |- api.go ---------------|
```

In summary, we only allow the domain/service/package to communicate to each other via their `api` and not directly into the internal implementation.

```text
|- ledger
      |- api
          |- api.go <--------------|
      |- internal                  |
          |- postgres              |
                |- postgres.go     | import
|- wallet                          |
      |- api                       |
          |- api.go ---------------|
```

### Managing Consistency And Data State Between Domain

When talking about domain separation we already learned about how different domain communicate with each other to produce the wanted end result for the users.

When talking about domain to domain communication, consistency of the data is not the only thing we want to take care about. We also want to ensure the latency is also considered.

```text
|------------------------------------------------------------------------------------|
|           |------------------|                      |------------------|           |
|           |   Wallet Domain  |                      |  Ledger Domain   |           |
|           |                  |                      |                  |           |
|   request |    |--------|    |        call          |    |--------|    |           |
|  ------------->| wallet |------------------------------->| Ledger |    |           |
|           |    |--------|    |                      |    |--------|    |           |
|           |        |         |                      |        |         |           |
|           |        | store   |                      |        | store   |           |
|           |        v         |                      |        v         |           |
|           |  |------------|  |                      |  |------------|  |           |
|           |  | PostgreSQL |  |                      |  | PostgreSQL |  |           |
|           |  |------------|  |                      |  |------------|  |           |
|           |------------------|                      |------------------|           |
|       <--------------------------> <-----------> <------------------------->       |
|        Time spent in wallet domain    Network    Time spent in ledger domain       |
|                                    <--------------------------------------->       |
|                                             Wallet domain waiting                  |
|       <-------------------------------------------------------------------->       |
|                                   Total time spent                                 |
|------------------------------------------------------------------------------------|
```

Taking latency to the consideration is imporatnt as we want to ensure the "request" is being handled as soon as possible thus we can give the feedback faster to the end user. With higher latency, there is a subsequent problem that need to be handled. As you can see in the above example, the "total time spent" depends on the "time spent in wallet domain" + "network" + "time spent in ledger domain".

```go
package api

import (
	"github.com/albertwidi/pkg/postgres"

	
	walletpg "github.com/albertwidi/go-example/services/wallet/internal/postgres"
	ledgerapi "github.com/albertwidi/go-example/services/ledger/api"
)

// API for wallet pacakge.
type API struct{
	q         *walletpg.Query
	ledgerAPI *ledgerapi.API
}

// New returns the wallet API by injecting the dependency to the function.
func New(pg *postgres.Postgres, ledgerAPI *ledgerapi.API) *API {
	return &API{
		q: walletpg.New(pg),
		ledgerAPI: ledgerAPI,
	}
}

func (a *API) CreateWallet(ctx context.Context) error {
	fn := func(pg *postgres.Postgres) error {
		q := walletpg.New(pg)
		if err := q.CreateWallet(ctx); err != nil {
			return err
		}
		return nil
	}
	// When creating a wallet, the internal wallet api will also create an account for the ledger. The CreateAccount api
	// allowed a foreign function to be passed and it will invoke the function inside the same database transaction to
	// create the ledger account.
	if err := a.ledgerAPI.CreateAccount(ctx, fn); err != nil {
		return err
	}
	return nil
}
```

As you can see above, the `wallet/api` package imports both `postgres` and `ledger/api` as dependencies. Both of them are needed to create a new `API` struct of wallet's domain, this mean to use `wallet/api` now you explicitly saying `postgres` and `ledger` is needed.

If you look closely in above example, you will notice that we are tyring to use the APIs of `ledger` package while also using the same underlying infrastructure other domain uses.

```go
func (a *API) CreateWallet(ctx context.Context) error {
	fn := func(pg *postgres.Postgres) error {
		q := walletpg.New(pg)
		if err := q.CreateWallet(ctx); err != nil {
			return err
		}
		return nil
	}
	// When creating a wallet, the internal wallet api will also create an account for the ledger. The CreateAccount api
	// allowed a foreign function to be passed and it will invoke the function inside the same database transaction to
	// create the ledger account.
	if err := a.ledgerAPI.CreateAccount(ctx, fn); err != nil { // <- THIS ONE.
		return err
	}
	return nil
}
```

When creating a wallet account, the ledger account also need to be created as `wallet` uses `ledger` infrastructure to record its balance. But, `wallet` need to ensure the data consistency.

### Internal Pacakge
