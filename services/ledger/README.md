# Ledger

The ledger service provides APIs to create account and moves money from one account to another.

## Accounting Concept

The `ledger` service is created based on the accounting concept of [ledger](https://en.wikipedia.org/wiki/Ledger). It is basically
a book or collection of transactions.

## Database & Table Design

We use `PostgreSQL` to store the `ledger` data.

### Movement
