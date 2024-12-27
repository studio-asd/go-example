# Ledger & Wallet

The go-example project provides a project sample for two different domains, `ledger` and `wallet`.

## Ledger

## Wallet

Wallet domain provide a wallet and stores the money in a currency inside it. Under the hood, wallet use `ledger` to store its balance. While `ledger` maintains the balance and bookeeping of the wallet transactions,
the wallet domain maintains its abstraction to serve user-facing features.

For example, we cannot do this inside the `ledger`:

1. Set a `status` for a given ledger.

  The `status` for a wallet is a pretty common business use-case as user might want to create a `wallet_type` for different-different purposes, and one of the wallet might get deactivated/deleted
  by the user.

2. Set a different `wallet_type` for each ledger.

  By default, `ledger` is only a place to store a historical data of an account. So it doesn't understand any concept of `type` inside of it, everything is the same. So the `wallet` domain need
  to maintain its own abstraction of `wallet_type` inside its database.
