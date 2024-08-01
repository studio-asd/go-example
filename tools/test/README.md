# Test

We use the `main.go` inside this folder to test go codes inside this repository.

Basically the test do:

1. Setup the service dependencies, for example `PostgreSQL`.
1. Applying the database schema.
1. Executes `go test -v -race ./...`
1. Tearing down the service dependencies.
