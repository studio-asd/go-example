# Ledger & Wallet, The Tech

> This is not a Production ready software and not intended to be used outside of the go-example project. This project is intended to give an example of how to do programming in Go. This document will explain the product point of view and how we leverage technology to solve some problems. For product document, please look at [here](ledger-and-wallet-product.md).

The go-example project provides a project sample for two different domains, `ledger` and `wallet`. The `wallet` domain is specifically designed in the scope of `e-wallet` application. The name of `ledger` is coming from [accounting ledger](https://en.wikipedia.org/wiki/Ledger).

While there are some applications that maintains the strict order of transactions(like stock/cryptocurrency exchange), this project doesn't guarantee strict ordering as it doesn't have to. The [idempotency]((https://en.wikipedia.org/wiki/Idempotence)) of the transaction is far more important for the transaction in this project.

## Ledger

### Database Tables & Relation

**Accounts Table:**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| account_id | varchar | No | Yes |
| parent_account_id | varchar | No | No |
| account_status | account_status | No | No |
| currency_id | int | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |

**Movements Table:**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| movement_id | varchar | No | Yes |
| idempotency_key | varchar | No | No* |
| movement_status | movement_status | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |
| reversed_at | timestamptz | Yes | No |
| reversal_movement_id | varchar | Yes | No |

*Note: idempotency_key has a UNIQUE constraint

**Accounts Balance Table:**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| account_id | varchar | No | Yes |
| parent_account_id | varchar | Yes | No |
| currency_id | int | No | No |
| allow_negative | boolean | No | No |
| balance | numeric | No | No |
| last_movement_id | varchar | No | No |
| last_ledger_id | varchar | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |

**Accounts Balance History Table:**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| history_id | bigint | No | Yes |
| movement_id | varchar | No | No |
| ledger_id | varchar | No | No |
| account_id | varchar | No | No |
| balance | numeric | No | No |
| previous_balance | numeric | No | No |
| previous_movement_id | varchar | No | No |
| previous_ledger_id | varchar | No | No |
| created_at | timestamptz | No | No |

**Accounts Ledger Table:**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| internal_id | bigint | No | Yes |
| ledger_id | varchar | No | No* |
| movement_id | varchar | No | No |
| account_id | varchar | No | No |
| movement_sequence | int | No | No |
| currency_id | int | No | No |
| amount | numeric | No | No |
| previous_ledger_id | varchar | No | No |
| created_at | timestamptz | No | No |
| client_id | varchar | Yes | No |
| reversal_of | varchar | Yes | No |

*Note: ledger_id has a UNIQUE constraint

**Reversed Movements Table:**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| movement_id | varchar | No | No |
| reversal_movement_id | varchar | No | No |
| reversal_reason | text | No | No |
| created_at | timestamptz | No | No |

## Wallet

### Database Tables & Relation

**Wallet Users Table:**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| user_id | varchar | No | No |
| user_type | varchar | No | No |
| user_status | int | No | No |
| intermediary_wallet_id | varchar | No | No |
| chargeback_wallet_id | varchar | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |

**Accounts Table:**
| Column Name | Data Type | Nullable | Primary Key | Description |
|------------|-----------|----------|-------------|-------------|
| account_id | varchar | No | Yes | |
| parent_account_id | varchar | No | No | |
| account_status | account_status | No | No | |
| currency_id | int | No | No | |
| created_at | timestamptz | No | No | |
| updated_at | timestamptz | Yes | No | |

*Table Description: Used to store all user accounts*

**Movements Table:**
| Column Name | Data Type | Nullable | Primary Key | Description |
|------------|-----------|----------|-------------|-------------|
| movement_id | varchar | No | Yes | UUID_v7 |
| idempotency_key | varchar | No | No* | |
| movement_status | movement_status | No | No | |
| created_at | timestamptz | No | No | |
| updated_at | timestamptz | Yes | No | |
| reversed_at | timestamptz | Yes | No | A marker for reversal. If the reversed_at is not null then the movement is reversed |
| reversal_movement_id | varchar | Yes | No | The id of movement where the reversal is being performed |

*Note: idempotency_key has a UNIQUE constraint*

*Table Description: Used to store all movement records*

**Accounts Balance Table:**
| Column Name | Data Type | Nullable | Primary Key | Description |
|------------|-----------|----------|-------------|-------------|
| account_id | varchar | No | Yes | |
| parent_account_id | varchar | Yes | No | |
| currency_id | int | No | No | |
| allow_negative | boolean | No | No | Allows some accounts to have negative balance. For example, for the funding account we might allow the account to have negative balance |
| balance | numeric | No | No | |
| last_movement_id | varchar | No | No | |
| last_ledger_id | varchar | No | No | |
| created_at | timestamptz | No | No | |
| updated_at | timestamptz | Yes | No | |

*Table Description: Used to store the latest state of user's balance. This table will be used for user balance fast retrieval and for locking the user balance for movement*

**Accounts Balance History Table:**
| Column Name | Data Type | Nullable | Primary Key | Description |
|------------|-----------|----------|-------------|-------------|
| history_id | bigint | No | Yes | |
| movement_id | varchar | No | No | |
| ledger_id | varchar | No | No | The id of where the balance is being summarized, the SUM of the balance in the accounts_ledger should be the same if we summarize everything up to this ledger_id |
| account_id | varchar | No | No | |
| balance | numeric | No | No | |
| previous_balance | numeric | No | No | |
| previous_movement_id | varchar | No | No | |
| previous_ledger_id | varchar | No | No | |
| created_at | timestamptz | No | No | |

*Table Description: A historical balance changes for an account based on movement_id. The historical balance is per movement_id and not per ledger_id because we pre-calculates everything inside the system. And because we are building the ledger optimistically, the historical amount can changed by the time we calculate because we are not locking the balance up-front(as this will be expensive). So it makes more sense to create the history based on movement_id because we will do that in bulk rather than ledger_id. This table can be used for various things like retrieving opening and ending balance of an account at a given time.*

**Accounts Ledger Table:**
| Column Name | Data Type | Nullable | Primary Key | Description |
|------------|-----------|----------|-------------|-------------|
| internal_id | bigint | No | Yes | |
| ledger_id | varchar | No | No* | The ledger id is the secondary unique key of the accounts_ledger. Even though we have unique constraint, but the client can always refer themselves to this unique id when it comes to reconciliation |
| movement_id | varchar | No | No | |
| account_id | varchar | No | No | |
| movement_sequence | int | No | No | The ordered sequence of movement inside a movement_id |
| currency_id | int | No | No | |
| amount | numeric | No | No | |
| previous_ledger_id | varchar | No | No | Will be used to track the sequence of the ledger entries of a user |
| created_at | timestamptz | No | No | |
| client_id | varchar | Yes | No | An identifier that the client can use in case they want to link their ids to per-ledger-row. With this, there are many cases they can use with the ledger |
| reversal_of | varchar | Yes | No | A ledger_id that being reversed by this ledger_id |

*Note: ledger_id has a UNIQUE constraint*
*Table Description: Used to store all ledger changes for a specific account. A single transaction can possibly affecting multiple accounts in the ledger. Row in this table is IMMUTABLE and should NOT be updated.*

**Wallet Users Table:**
| Column Name | Data Type | Nullable | Primary Key | Description |
|------------|-----------|----------|-------------|-------------|
| user_id | varchar | No | No | |
| user_type | varchar | No | No | |
| user_status | int | No | No | |
| intermediary_wallet_id | varchar | No | No | An intermediary wallet for state transition for all wallets inside a given user. The intermediary wallet is useful for state transition because its lock contention is per user-basis before we are transitioning the funds to a global system wallet like withdrawal |
| chargeback_wallet_id | varchar | No | No | A wallet to chargeback the user for any case of reversal. For example when payment reversal happens and the user doesn't have enough
