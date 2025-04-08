# Boostrap

The bootstrap service is responsible for initializing the application and setting up the necessary dependencies. It is the first service to be started when the application is launched.

## How Boostrap Service Works

The bootstrap service works by registering all the bootstrap functions in the application for a certain version of the application. This is why you will see something like this in the [bootstrap.go](bootstrap.go):

```go
// Put the bootstrapper for each version here, although we will re-sort and check the bootstrapper later on, please always
// put the new bootstrapper below the previous one.
b := []bootstrapper{
	&v0Bootstrapper{
		goExampleDBMigrator: goExampleDBMigrator,
		userDBMigrator:      userDBMigrator,
	},
}
```

When the service is started, it will trigger the bootstrap service and check whether the current service version is available in the map of versions. If it is available, then it will trigger the bootstrap function. But it won't blindly trigger the bootstrap function though. We are storing state of the bootstrap versions in the database to understand whether the bootstrap function has been executed successfully or not. This means it will also store the latest version of bootstrap service. If the semver is greater equal than the current service version, then it will ignore the bootstrap function, and vice-versa.

```mermaid
```

## How Bootstrap Is Tested

The bootstrap is tested by executing all the bootstrap functions from `v0...vN` and check wehther there is a failure when it being executed. As we are using semver for the version, the version must be correct and we will sort all the semver by comparing all the version string(because Go's map is unordered).

Currently, this is enough because we are building a small service with as tiny dependencies as possible. For a more complex setup with multiple dependencies, it will require more rigorous testing and other methodology to ensure the state is correct.

## When To Not Use Boostrap

> In general, bootstrap should not be used if we cannot ensure the state of the dependencies in an automatic way. But here's some guide when to not use it.

1. When we are altering a database schema that causes locks on a certain table that frequently used. This can cause downtime for the application as thet table will not be accessible.
1. When we CANNOT create database index concurrently for some reasons. Most of the time, this is not a problem because we can use `CREATE INDEX CONCURRENTLY` to create the index without locking the table.
1. When there are multiple dependencies that need to be changed at the same time and we cannot guarantee the consistency of the data between the dependencies. This means a more complex setup and plan need to be executed to ensure the data is consistent and correct.

## Security

As of now the bootstrap service is running with full user previledge, thus it can do database migrations and altering the schema of the database. While this is not ideal for Security and dangerous in practice, this bootstrap service is only intended for showcase and not to be used in production environment.

In a production environment, its better to build a different `binary` to bootstrap the service so it can use a different user with different priviledge from the application.
