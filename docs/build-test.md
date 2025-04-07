# Build & Test

This document will outline on how we build the `go-example` program and how we tests it.

## Testing

As testing is essential for every programs, we also want to ensure the program is being tested carefully in different-different layers. In general, this is how we tests the program:

1. Unit Test

	The `unit test` tests a single unit of function to check whether the function is working as intended based on the input parameter for the function. We usually parallelize the tests as much as possible to ensure we have a fast feedback.

2. Integration Test

	The `integration test` tests integrations of several components in a program to ensure that the behavior of a feature is working as intended from the `white box` perspective. This means the test will check the validity of the data in the database if the feature is touching the database. We usually parallelize the tests as much as possible to ensure we have a fast feedback. Not all integration tests can be parallelized depending on the dependencies and the complexities to parallize the tets.

3. End to End Test

	The `end to end test` tests the whole program via HTTP `api` that exposed by the program. The `end to end` test is a `black box` test and we tests everything from the `api` perspective. This means we will not check the database records directly, but we will try to retrieve the data after execution to ensure the result is as expected. It is hard to parallelize the `end to end test`, so most of the things will run in sequence.

### Short, Long, Longest

The approach of the test will also affects the testint time needed by the program, so we split our tests into three different categories:

1. Short

	The short build machine will tests the program using `go test -v -race -short` flags. This means we conciously `skip` all the itegration tests inside the program.

2. Long

	The long build machine will tests the program using `go test -v -race` flags. This means we will run all the integration tests only.

3. Longest

	The longest build machine will tests the `testing/e2e` folder where the end to end tests is exists. This means only the e2e test will be running on this machine.
