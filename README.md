# Go Example

In this Go example, we will create a simple service called `ledger`. The `ledger` is designed to manage money transfer from one user to another in atomic fashion using database transaction.

The goal of this example is to show an example of running Go program and how easy to build one with the help of some publicly available packages and helper scripts.

## Disclaimer

The `ledger` service is not a production ready service/system, and it only designed for an **example** and transfer only purpose. Although the concept is similar with some of ledger systems out there, usually a specific system is designed specifically to solve a set of problems. So I don't recommend to use this design as is for your production system.

## System Requirements

1. Go programming language v1.22+
2. Docker client/desktop.
3. Docker compose (this should be covered within the newest docker client).
4. PostgreSQL (should be covered by Docker as we will be using `docker compose`)

## Used Packages/Libraries

1. [Sqlc](https://sqlc.dev/)

   We will leverate `sqlc` to generate the queries needed for interacting with the database.

   By using `sqlc`, we automatically validate our queries against our schema. The code generation will fail if our query is invalid. Furthermore, as we write `sql` file, we will be able to use [sqlfluff](https://sqlfluff.com/) to lint our queries.

2. [Squirrel](https://github.com/Masterminds/squirrel)

   We will use `squirrel` to dynamically build some of the queries. This cannot be done by `sqlc`.

3. [Decimal](https://github.com/shopspring/decimal)

   We will use `decimal` to maintain abritrary precission decimal value.

4. [Pkg](https://github.com/studio-asd/pkg)

   We will use `pkg` and its helper package to make the development of the service easier.

## Program/Package Design

This project use a simple package design to separate each service into separate package. For example, we have `./ledger` as the main folder of `ledger` service. Inside the package, we have `./service` or `service` package to represent the service entrypoint. We have another package inside `ledger` called `postgres` to represent the database layer. The `./ledger` package is not used to the entrypoint of the service because it is used to share types between the `service` and `postgres`.

```text
             import  |-------|
             |-----> |Service|
             |       |-------|
|------|     |           ^
|Ledger| -----           | import
|------|     |           |
             |       |--------|
             |-----> |Postgres|
             import  |--------|
```

They layer is made thin on-purpose to make life easier while building a pogram. A layer should not be introduced unless we have a **very** good reason to add them and what problem they solve.

But, this design is not without problems. There are several problems that occurs during development.

1. There are `type`s that need to be shared between `service` and `postgres` layer.

  To solve this issue, we use the top level package, in this case `ledger` to store all `exported` types. So, it consistent that packages/client need to import the top level package to use the type for argument, etc.

2. Postgres package exposed to other packages.

## Service Requirements

The `ledger` service should be able to

1. `Move` money from one account to another. There are three use cases of moving money to comply with [double entry accounting](https://en.wikipedia.org/wiki/Double-entry_bookkeeping).

   - From one account to another account.
   - From on account to another accounts.
   - From multiple accounts to multiple accounts.

2. The service should be able to reject/ignore `idempotent` transaction requested to the service.

3. The service should keep the consistency of records inside the database and ensure no double-spending is possible.

4. The service should provide a set of APIs via `HTTP` interface for other systems to interact.
