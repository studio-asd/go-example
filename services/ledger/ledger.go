package ledger

import (
	"time"

	"github.com/shopspring/decimal"
)

type LedgerEntry struct {
	LedgerID         string
	MovementID       string
	AccountID        string
	MovementSequence int
	CurrencyID       int32
	Amount           decimal.Decimal
	PreviousLedgerID string
	ClientID         string
	CreatedAt        time.Time
	Timestamp        int64
}

type MovementSummary struct {
	LedgerID      string
	MovementID    string
	FromAccountID string
	ToAccountID   string
	Amount        decimal.Decimal
}

// AccountMovementSummary is the summary of balance changes inside a movement. This means every entries will be summarized so we will
// have both the final balance/state of an account and the total balance changes for an account. This is useful for two reasons:
//  1. We know the total balance changes for an account, so we can compute the ending balance in the database layer later
//     when lock happens. This will happen if the 'last_ledger_id' is not the same with the one in the summary.
//  2. We don't have to re-compute the ending balance and just use the EndingBalance if the 'last_ledger_id' is the same.
//     This means no transaction are executed for the particular account in the of processing this transaction.
type AccountMovementSummary struct {
	// BalanceChanges is the total of balance change of an account. The amount of changes can be either positive or negative depends
	// on the sum of the changes.
	BalanceChanges decimal.Decimal
	// NextLedgerID is the next ledger id for the account balance to be recorded. The next ledger id will be set
	// as the last ledger id inside the database.
	NextLedgerID string
	// LastLedgerID is the latest ledger_id recorded when we are retrieving the account balance. This value will be compared
	// inside the data layer to check whether there are new records being recorded mid-flight. If there are no new records,
	// then we don't have to recalculate the balance as the number will be the same(even though we still need to lock the row).
	LastLedgerID string
	// EndingBalance is the ending balance of an account, calculated from the latest balance of 'latest_ledger_id'. This balance
	// should not be used when the 'last_ledger_id' != last ledger id when retrieving the balance in a lock.
	EndingBalance decimal.Decimal
}

// MovementLedgerEntries is the entries of ledger and balances for a single movement. In a single movement, it is possible to
// have multiple entries of ledger and balance changes. There are three types of movement possible in the current setup:
//
// 1. From one account to another account.
// 2. From one account ot many accounts.
// 3. From many accounts to many accounts.
type MovementLedgerEntries struct {
	// MovementID is a unique id issued by the ledger service. This id is the main identifier when moving money from/to accounts.
	MovementID string
	// IdempotencyKey ensure the transaction is idempotent to the same key. This means the exact key for a transaction
	// can't happen twice. The client should use this key to ensure movement uniqueness.
	IdempotencyKey  string
	LedgerEntries   []LedgerEntry
	MovementSummary []MovementSummary
	// AccountsSummary is the summary of balance changes per account basis. The amount is summarized
	// so its easier for us to check whether the account balance is sufficient or not and to udpate
	// them in bulk.
	AccountsSummary map[string]AccountMovementSummary
	// Accounts is the list of accounts which balance is affected byt the movements. While we already understand
	// the list of accounts in the account summary, but having it in array is also beneficial to retrieve account based
	// informations.
	Accounts  []string
	CreatedAt time.Time
}
