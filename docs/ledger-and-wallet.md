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
  to maintain its own abstraction of `wallet_type` inside its database. For business use cases, wallet usually have different-different kind of types. For example we can have `main` wallet,
  `savings` wallet and other kind of wallet based on the business and user needs.

### Chargeback

Chargeback is a transaction that created by the system to charge the user with the amount of money that they should not be able to spent/receive(but it happen anyway). The chargeback
is needed because there are some edge-cases because of failure in the dependencies or internal system that causing the business owner to lose money.

While the idea of chargeback is simple, the intention and communication to the end user must be clear so they know that they are spending money that not belong
to them and they are being charged because of that. So there are several things that we need to consider:

1. A charge must be a separate transaction.

    As the chargeback need to be communicated clearly to ther user, a charback need to be a separated transaction for the user. A chargeback
    also need a link/referrence to another transaction.
