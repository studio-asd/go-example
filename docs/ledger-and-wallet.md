# Ledger & Wallet

> This is not a Production ready software and not intended to be used outside of the go-example project. This project is intended to give an example of how to do programming in Go.

The go-example project provides a project sample for two different domains, `ledger` and `wallet`. The `wallet` domain is specifically designed in the scope of `e-wallet` application. The name of `ledger` is coming from [accounting ledger](https://en.wikipedia.org/wiki/Ledger).

Please note that the `ledger` and `wallet` does not maintains the strict order of the user requests. While the strict order of transactions are important in some applications(for example in stock/cryptocurrency exchange), most of `e-wallet` system does not need the ordered guarantee. Instead of order guarantee, the system need the transactions to be `idempotent`.

We recommend the reader to learn the basic accounting so the content can be more relateable. Let's discuss a bit about it to save time.

1. [Ledger](https://en.wikipedia.org/wiki/Ledger)

    Ledger keeps the records of all accounts transactions. Each transaction in the ledger recorded as a **pair** of either debit or credit. This method is called [double entry bookeeping or double entry accounting](https://en.wikipedia.org/wiki/Double-entry_bookkeeping).

2. [Double Entry Bookeeping](https://en.wikipedia.org/wiki/Double-entry_bookkeeping)

    As being mentioned in above section, the double entry bookeeping is a method to record two-sided accounting to maintain financial information. This method is fundamental as every debit must have a credit record alongside of it. The purpose of double entry bookkeeping is to detect error or fraud in financial records.

    For example, we have a case of a money transfer of $100 from `user_a` to `user_b`. Then the record will looked like this

    |user|debit|credit|
    |-|-|-|
    |user_a|$100|-|
    |user_b|-|$100|

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

All of these transactions are [idempotent](https://en.wikipedia.org/wiki/Idempotence) by default, so the system will not process transactions with the same identifier twice.

#### Deposit Transaction

Deposit transaction is a way to insert money from outside of the system into the wallet ecosystem. The amount of digital money should reflect the money that has been transferred into the bank account owned by the digital money product. Or, if you are in the cyrptocurrency industry, the money should reflect the money that stored inside your cryptocurrency wallet. The deposit transaction uses a special wallet called `deposit wallet`. Only this wallet can "create" money inside the ecosystem and give them to the user.

> The cryptocurrency industry still need the off-chain solution because they need to off-ramp the digital currency to fiat or real money. As we are living in the world dominated by off-chain ecosystem, off-ramp is unavoidable.

```text
             |------|
             | Bank |
             |------|-------------------------------------------|
        |----| $$$$ |-------------|                             |
        |    |------|             |                             |
        |                     |------------|               |----------|
        |                     | Acc Wallet |               | Acc User |
        |                     |------------|               |----------|
        |                     |   $10,000  |               |   $100   |
        |                     |------------|    + $100     |----------|
        |              Credit |   $10,100  | <------------ |    $0    | Debit
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
                          |  - $10,000     |             |     $0      |
                          |----------------|   + $100    |-------------|
                    Debit |  - $10,100     | ----------> |    $100     | Credit
                          |----------------|             |-------------|
```

While the above model is a simplified model of how the transfer happens between accounts in the bank, in general the concept is still the same. The double entry accounting is the foundation to record money movement, thus money is flowing from a valid souce into a a valid destination.

Maybe you have a question on why the `deposit wallet` records negative balance? The short answer is because the user's money is now becomes the [liability](https://en.wikipedia.org/wiki/Liability_(financial_accounting)#:~:text=In%20financial%20accounting%2C%20a%20liability,obligation%20arising%20from%20past%20events.) of the wallet platform, the platform is now responsible for the user's money. So, even though we are recording a surplus of balance in the bank(from the users), we are liable to give the money back in the platform.

The deposit flow is crucial because now we are able to create the digital money out from nowehre, thus ensuring the backing one to one(1:1) liable assets is really important. The mapping should not only for the total of assets in the bank, but the mapping of each user assets.

#### Withdrawal Transaction

Withdrawal transaction is a way to withdraw money from the wallet ecosystem into other ecosystem that receives the same currency. It can be Banks, another wallet or other type of ecosystem that recognize the system's assets as interchangeable assets. All user need to withdraw their money via `withdrawal wallet`, the money inside the `withdrawal wallet` cannot be transferred and will be there forever. However, to ensure all transactions to the `withdrawal wallet` is a valid and successful transactions, when withdrawal happens user will automatically transfer their funds into the user's `escrow wallet` first. The reason is, because withdrawal transaction usually involves third party connections as the money destination. As we should treat all thrid party connections as "uncertain"(because it can fail), we need a place where we are able to put the money temporarily before it marked as a success and fully transferred to the `withdrawal wallet`. To understand more on why the `escrow wallet` is needed, please check the [scalability](#scaling-the-wallet) section.

```text
|---------------|
| Wallet System |
|---------------|
|      $$$      |---|
|---------------|   |
                    |
                    |
  |--------------------------------------------------------------|          |---------------------------------|
  | User                                                         |          | System                          |
  |                                                              |          |                                 |
  |        |-------------|              |---------------|        |          |    |-------------------|        |
  |        | Main Wallet |              | Escrow Wallet |        |          |    | Withdrawal Wallet |        |
  |        |-------------|              |---------------|        |          |    |-------------------|        |
  |        |    $1,000   |              |       $0      |        |          |    |      $10,000      |        |
  |        |-------------|    + 500     |---------------|        |          |    |-------------------|        |
  |  Debit |     $500    | -----------> |      $500     | Credit |          |    |      $10,000      |        |
  |        |-------------|              |---------------|        |  + 500   |    |-------------------|        |
  |                               Debit |       $0      | ---------------------> |      $10,500      | Credit |
  |                                     |---------------|        |          |    |-------------------|        |
  |                                        |                     |          |                                 |
  |----------------------------------------|---------------------|          |---------------------------------|
                                           |                ^
                                           | + 500          |
                                           v                |   Success
                                   |--------------------|   | Notification
                                   | Third Party System |---|
                                   |--------------------|
                                   |        $$$         |
                                   |--------------------|
```

As you can see above, the money will only be transferred to the `withdrawal` wallet when there are "success notification" from the third party system that receives the money. If the transfer to the third party system results in failure, then the money will be transferred back to the `main wallet` fomr the `escrow wallet`.

#### Transfer Transaction

Transfer transaction is a way to move money from a user wallet to another wallet of user. The `transfer` can **only** be triggered by the user/client that authorized to access the wallet.

```text
|---------------|
| Wallet System |
|---------------|------------------------|
|      $$$      |---|                    |
|---------------|   |                    |
                    |                    |
                    |                    |
               |--------|            |--------|
               | User A |            | User B |
               |--------|            |--------|
               |  $500  |            |  $300  |
               |--------|    +200    |--------|
         Debit |  $300  | ---------> |  $500  | Credit
               |--------|            |--------|
```

The `transfer` transaction is very straightforward as it only moving money to another account, thus a double entry accounting is happened here.

#### Payment Transaction

A `payment` transaction is a way to categorized a movement of money as a `payment`. While this is simply a movement from user to another user(merchant), categorizing the transaction is important to not mix the transaction types. The `payment` transaction also supports additional metadata that supported to ensure there are enough information for the end user, for example additional fee because of tax, service, etc.

```text
|---------------|
| Wallet System |
|---------------|-----------------------------|
|      $$$      |---|                         |
|---------------|   |                         |
                    |                         |
                    |                         |
               |--------|            |-------------------|
               | User A |            | User D (Merchant) |
               |--------|            |-------------------|
               |  $500  |            |       $1,000      |
               |--------|    +200    |-------------------|
         Debit |  $300  | ---------> |       $1,200      | Credit
               |--------|    |       |-------------------|
                             |
                             |
                     |----------------|
                     | $150 - Price   |
                     |  $15 - Tax     |
                     |  $35 - Service |
                     |----------------|
```

As you can see above, even though the total amount of transaction is $200, but the price of the item is not $200. There are additional items that being added to the price that incur the cost.

But, usually the `payment` transaction is not this simple though. As the provider of the payments, the wallet ecosystem also need some incentive so it able to continue to operate. Thus the system will retrieve some percentage of the item price as the revenue of the wallet ecosystem.

```text
|---------------|
| Wallet System |-------------------------------------------------------------|
|---------------|-----------------------------|                               |
|      $$$      |---|                         |                               |
|---------------|   |                         |                               |
                    |                         |                               |
                    |                         |                               |
               |--------|            |-------------------|            |--------------|
               | User A |            | User D (Merchant) |            | Payment Fees |
               |--------|            |-------------------|            |--------------|
               |  $500  |            |       $1,000      |            |   $10,000    |
               |--------|    +200    |-------------------|            |--------------|
         Debit |  $300  | ---------> |       $1,200      | Credit     |   $10,000    |
               |--------|            |-------------------|            |--------------|
         Debit | $292.5 | ------------------------------------------> |   $10,007.5  | Credit
               |--------|                    + $7.5                   |--------------|
```

There we go, there is an additional account called `payment fees` involved in the transaction and the user will pay 5% of $150(item price) as the payment fees($7.5).

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

There are two wallet criteria

#### System Wallet

1. Deposit wallet

    Deposit wallet is where the money is coming from when the "real money" is coming in to the bank's account of the e-wallet platform. The deposit wallet then disburse the same amount of money to the end user.

2. Withdrawal wallet

    Withdrawal wallet is where the money goes to when the end user wants to withdraw their money to other ecosystem such like banks, other wallet ecosystem, etc.

3. Withdrawal fees wallet

    Withdrawal fees wallet is where the `fee` money from withdrawal transaction is being kept by the system.

4. Payment fees wallet

    Payment fees wallet is where the `fee` money from payment transaction is being kept by the system.

5. Consolidated fees wallet

    Consolidated fees wallet is where all `fee` will be consilidated automatically by the system from all fees wallet. The consolidation of the fees is needed because there are several fees wallet exists within the system. So it will be easier for the system operator to withdraw all the fees money at once.

6. Intermediary wallet

#### User Wallet 

1. Uer's main wallet
2. User's escrow wallet
3. Merchant wallet

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

### Fee Management

### Reconciliation And Monitoring

#### Reconciling Deposit Transaction

The reconciliation between "real money" and the digital money inside the wallet is something that need to be performed in daily or even in a more frequent basis. Because only by reconciling the data one by one we are really sure all `deposit` transactions are tally.