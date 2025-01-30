# Ledger & Wallet

> This is not a Production ready software and not intended to be used outside of the go-example project. This project is intended to give an example of how to do programming in Go.

The go-example project provides a project sample for two different domains, `ledger` and `wallet`. The `wallet` domain is specifically designed in the scope of `e-wallet` application.

Please note that the `ledger` and `wallet` does not maintains the strict order of the user requests. While the strict order of transactions are important in some applications(for example in stock/cryptocurrency exchange), most of `e-wallet` system does not need the ordered guarantee. Instead of order guarantee, the system need the transactions to be `idempotent`.

## Ledger

### Ledger Transaction & Lock

Money inside one account in the ledger will be monotonicaly increased and decreased. The ledger will ensure a lock is held when a transaction happen and no other transaction can change the account. Because we are using PostgreSQL as our database, it only rational that we also utilize the database to do the lock. Do you remember that we have `ledger_accounts_balance` where we stores the latest `balance` of an account? Everytime we are doing a transaction, we will lock the account row using `SELECT FOR UPDATE` within a [read committed]() transaction. Using the read committed isolation is enough for this use case because we are locking the `row` by using `FOR UPDATE`, so other sessions will not be able to change the data.

By locking for each row, each account can concurrently use their ledger without get affected by another account. This means as long as the accounts are separated for the end users the


## Wallet

Wallet domain provide a wallet and stores the money in a currency inside it. Under the hood, wallet use `ledger` to store its balance. While `ledger` maintains the balance and bookeeping of the wallet transactions, the wallet domain maintains its abstraction to serve user-facing features.

For example, we cannot do this inside the `ledger`:

1. Set a `status` for a given ledger.

    The `status` for a wallet is a pretty common business use-case as user might want to create a `wallet_type` fo different-different purposes, and one of the wallet might get deactivated/deleted by the user.

2. Set a different `wallet_type` for each ledger.

    By default, `ledger` is only a place to store a historical data of an account. So it doesn't understand any concept of `type` inside of it, everything is the same. So the `wallet` domain need to maintain its own abstraction of `wallet_type` inside its database. For business use cases, wallet usually have different-different kind of types. For example we can have `main` wallet `savings` wallet and other kind of wallet based on the business and user needs.

### Wallet Transaction Types

There are several transaction types supported by the `wallet` system.

1. [Deposit](#deposit-transaction)
2. [Withdrawal](#withdrawal-transaction)
3. [Transfer](#transfer-transaction)
4. [Payment](#payment-transaction)
5. [Chargeback](#chargebeback-transaction)

#### Deposit Transaction

Deposit transaction is a way to insert money from outside of the system into the wallet ecosystem. The amount of digital money should reflect the money that has been transferred into the bank account owned by the digital money product. Or, if you are in the cyrptocurrency industry, the money should reflect the money that stored inside your cryptocurrency wallet.

> The cryptocurrency industry still need the off-chain solution because they need to off-ramp the digital currency to fiat or real money. As we are living in the world of off-chains, off-ramp is unavoidable.

```text
             |------|
             | Bank |
             |------|-------------------------------------------|
        |----| $$$$ |-------------|                             |
        |    |------|             |                             |
        |                     |------------|               |----------|
        |                     | Acc Wallet |               | Acc User |
        |                     |------------|               |----------|
        |                     |   $10.000  |               |   $100   |
        |                     |------------|    + $100     |----------|
        |              Credit |   $10.100  | <------------ |    $0    | Debit
        |                     |------------|               |----------|
        |
        |
------------------------------------- Separator ------------------------------------------
        |
        |
        |  Notification
        |  of Transfer
        v
 |-------------- |
 | Wallet System |
 |---------------|---------------------------------------------|
 |     $$$       |----------------|                            |
 |---------------|                |                            |
                          |----------------|             |-------------|
                          | Deposit Wallet |             | User Wallet |
                          |----------------|             |-------------|
                          |  - $10.000     |             |     $0      |
                          |----------------|   + $100    |-------------|
                    Debit |  - $10.100     | ----------> |    $100     | Credit
                          |----------------|             |-------------|
```

While the above model is a simplified model of how the transfer happens between accounts in the bank, in general the concept is still the same. The double entry accounting is the foundation to record money movement, thus money is flowing from a valid souce into a a valid destination.

Maybe you have a question on why the `deposit wallet` records negative balance? The short answer is because of [scaling issues](#scaling-the-wallet), we directly transfer the amount of money from the system's `deposit wallet` to the `user's wallet`. But while doing so, we are maintaining the exact opposite of our money inside the bank, which is fine as long as the data is tally.

The deposit flow is crucial because now we are able to create the digital money out from nowehre, thus ensuring the backing one to one(1:1) assets is really important. If this happen, then people can spend more than they should and somebody else(the company) should fill the gap in their book.

#### Withdrawal Transaction

To be added

#### Transfer Transaction

To be added

#### Payment Transaction

To be added

#### Chargebeback Transaction

Chargeback is a transaction that created by the system to charge the user with the amount of money that they should not be able to spent/receive(but it happen anyway). The chargeback is needed because there are some edge-cases because of failure in the dependencies or internal system that causing the business owner to lose money.

The effect of chargeback is pretty much similar to `reversal` in the ledger on where we want to create a reversed version of a transaction.

For example:

There might be a case where a payment gateway/bank give us a notification of `deposit`. Upon the notification, we increase the amount of wallet of our user accordingly. The user then receive the money in their wallet and use the money. But suddenly, the payment gateway/banks says that the deposit is invalid. So we don't really get the money in our bank. If this happen, we are actually giving the user "free" money for them, twice even. First, their money is not deducted in the bank. And the second they are able to spend the money in the wallet.

```
1. |-------------------
   | 
```

While the idea of chargeback is simple, the intention and communication to the end user must be clear so they know that they are spending money that not belong to them and they are being charged because of that. So there are several things that we need to consider:

1. A chargeback must be a separate transaction.

    As the chargeback need to be communicated clearly to ther user, a charback need to be a separated transaction for the user. A chargeback also need a link/referrence to another transaction that it wants to chargeback/reverse.

1. Chargeback wallet.

    A chargeback wallet might need to be introduced as a separate wallet as we don't want:

      1. Make the `main` wallet to go below zero(0).

          By default the `main` wallet should not goes below 0, as this will impose a risk of us allowing user to overspend.

      2. Complicate the flow of the main wallet.

          If the chargeback wallet is not separated, we need to do everything in the main wallet. This will make `main` wallet as a multi-purpose wallet and it will complicates the logic on the `main` wallet. The flow of money should be straightforward to minimize confusion.

      3. Ensure chargeback flow is clear to the end user.

With an additional chargeback wallet, its clear that the user need to make the chargeback wallet zero(0) first before they can spend their money. To be clear the flow of blocking of transactions and paying back the chargeback is defined as below:

1.  Transactions Blocking

    The chargeback wallet need to be always zero(0) before the user can spend their money inside the `main` wallet. These transactions are essentially blocked:

    1. Withdrawal.

        User should not be able to withdraw their money because withdrawal is usually triggered from the `main` wallet. User need to pay for the chargeback first.

    2. Payment.

        User should not be able to make any sorts of payment, so all payment transactions will get rejected. User need to pay the chargeback first.

    3. Transfer.

        Transfer between user should not be able to be executed as well. The user need to pay the chargeback first.

2. Chargeback Payment

    To be added

### Wallet Types

### Wallet Transaction & Lock

As we already know, `wallet` uses `ledger` to store its balance. This means the order of transaction and locks is guaranteed inside one ledger account only, and not across all ledgers owned by an account. And because the `wallet` uses `ledger` under the hood, it doesn't guarantee the order of the transactions on some edge-cases. For example, we have two different type of wallet: `main` and `chargeback` wallet. The `main` wallet can only be used to transact if the `chargeback` wallet is zero(0) in value. And there might be some cases where there are a race condition of a `chargeback` is being triggered at the same time when a user pays for something else. This means, the `chargeback` is not being prioritized and money already flowing out from the user's account to pay for something.

```text
|----
|----
```

In this case, should we make all transactions for `wallet` ordered for a reason like this? In our case, we don't think we need a strict ordering for all transactions in the `wallet` system. This is because most of the time, transactions that deduct user money is triggered by the end user and not fully automatic. Even though there is automatic recurring payment, it won't be at a random time thus race condition with chargeback will be a less likely condition. Thus we don't think it is worth it to introduce strict ordering in the `wallet` system.

### Scaling The Wallet

In the [previous](#wallet-transaction--lock) section we learned on how `wallet` utilize `ledger` under the hood to ensure the monotonicity of the balance. While a single wallet lock contention is manageable because each user has their own wallet, how about the wallet that shared across all users. For example, `deposit` and `withdrawal` wallet are shared across all users and a transaction to that wallet will fully lock the wallet until the transaction is completed. To understand the problem better, lets take a look again in how the transaction for each use case is being executed:

- Transfer from `user` to `user`
- Payment from `user` to `merchant`
- Deposit from `deposit` to `user`
- Withdrawal from `user` to `withdrawal`
