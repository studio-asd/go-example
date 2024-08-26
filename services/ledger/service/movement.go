package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/albertwidi/pkg/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/internal/currency"
	ledgerpg "github.com/albertwidi/go-example/ledger/internal/postgres"
	"github.com/albertwidi/go-example/services/ledger"
)

type Movement struct {
	ID        string
	CreatedAt time.Time
}

type MovementEntry struct {
	FromAccountID string
	ToAccountID   string
	Amount        decimal.Decimal
}

// movementEntriesToLedgerEntries converts the initial movement entries to ledger entries, check and
// summarize them into a correct entries.
func movementEntriesToLedgerEntries(movementID string, balances map[string]ledgerpg.GetAccountsBalanceRow, entries ...MovementEntry) (ledger.MovementLedgerEntries, error) {
	le := ledger.MovementLedgerEntries{
		// Create the entries with length of (entries*2) as we need the entry of DEBIT and CREDIT for each movement.
		LedgerEntries:   make([]ledger.LedgerEntry, len(entries)*2),
		AccountsSummary: make(map[string]ledger.AccountMovementSummary),
	}

	createdAt := time.Now()
	for idx, entry := range entries {
		// Check whether we have the correct currencies from and to account as we don't want to mix the currencies in the transfer.
		currFrom, err := currency.Currencies.GetByID(balances[entry.FromAccountID].CurrencyID)
		if err != nil {
			return ledger.MovementLedgerEntries{}, err
		}
		currTo, err := currency.Currencies.GetByID(balances[entry.ToAccountID].CurrencyID)
		if err != nil {
			return ledger.MovementLedgerEntries{}, err
		}
		if err := checkEligibleForMovement(checkEligible{
			FromAccountID: entry.FromAccountID,
			ToAccountID:   entry.ToAccountID,
			FromCurrency:  currFrom,
			ToCurrency:    currTo,
		}); err != nil {
			return ledger.MovementLedgerEntries{}, fmt.Errorf("%w: please check entry at index [%d]", err, idx)
		}

		// Normalize the amount based on the currency, because the exponent might be more than expected.
		amount := currFrom.NormalizeDecimal(entry.Amount)
		// As we have two entries in the ledger in every movement entry, the starting index will be always idx*2.
		arrIdx := idx * 2
		// sequence is always stats from 1, this is why we add the idx(which began with 0) with 1.
		sequence := idx + 1

		// Create the DEBIT record.
		debitAmount := amount.Mul(decimal.NewFromInt(-1))
		// The ledger_id is a UUIDV5 with namespace_oid and format of: movement_id:from_account_id:sequence.
		debitLedgerID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(movementID+":"+entry.FromAccountID+":"+strconv.Itoa(sequence))).String()
		le.LedgerEntries[arrIdx] = ledger.LedgerEntry{
			LedgerID:         debitLedgerID,
			MovementID:       movementID,
			AccountID:        entry.FromAccountID,
			Amount:           debitAmount,
			MovementSequence: sequence,
			CreatedAt:        createdAt,
			Timestamp:        createdAt.Unix(),
		}

		// Create the CREDIT record.
		// The ledger_id is a UUIDV5 with namespace_oid and format of: movement_id:from_account:sequence.
		creditLedgerID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(movementID+":"+entry.ToAccountID+":"+strconv.Itoa(sequence))).String()
		le.LedgerEntries[arrIdx+1] = ledger.LedgerEntry{
			LedgerID:         creditLedgerID,
			MovementID:       movementID,
			AccountID:        entry.ToAccountID,
			Amount:           amount,
			MovementSequence: sequence,
			CreatedAt:        createdAt,
			Timestamp:        createdAt.Unix(),
		}
		// Add the accounts to the list of accounts.
		if _, ok := le.AccountsSummary[entry.FromAccountID]; !ok {
			le.Accounts = append(le.Accounts, entry.FromAccountID)
		}
		if _, ok := le.AccountsSummary[entry.ToAccountID]; !ok {
			le.Accounts = append(le.Accounts, entry.ToAccountID)
		}

		// Add each account to the summary.
		// Sender account.
		var fromSummary ledger.AccountMovementSummary
		if summary, ok := le.AccountsSummary[entry.FromAccountID]; ok {
			fromSummary = summary
		} else {
			fromSummary = ledger.AccountMovementSummary{
				LastLedgerID:  balances[entry.FromAccountID].LastLedgerID,
				EndingBalance: balances[entry.FromAccountID].Balance,
			}
		}
		fromSummary.NextLedgerID = debitLedgerID
		fromSummary.BalanceChanges = fromSummary.BalanceChanges.Add(debitAmount)
		fromSummary.EndingBalance = fromSummary.EndingBalance.Add(debitAmount)
		// Check whether the ending balance is negative, we cannot allow negative balance for most the accounts.
		if fromSummary.EndingBalance.IsNegative() && !balances[entry.FromAccountID].AllowNegative {
			return ledger.MovementLedgerEntries{}, ledger.ErrInsufficientBalance
		}
		le.AccountsSummary[entry.FromAccountID] = fromSummary

		// Receiver account.
		var toSummary ledger.AccountMovementSummary
		if summary, ok := le.AccountsSummary[entry.ToAccountID]; ok {
			toSummary = summary
		} else {
			toSummary = ledger.AccountMovementSummary{
				LastLedgerID:  balances[entry.ToAccountID].LastLedgerID,
				EndingBalance: balances[entry.ToAccountID].Balance,
			}
		}
		toSummary.NextLedgerID = creditLedgerID
		toSummary.BalanceChanges = toSummary.BalanceChanges.Add(amount)
		toSummary.EndingBalance = toSummary.EndingBalance.Add(amount)
		// Check whether the ending balance is negative, we cannot allow negative balance for most the accounts.
		if toSummary.EndingBalance.IsNegative() && !balances[entry.ToAccountID].AllowNegative {
			return ledger.MovementLedgerEntries{}, ledger.ErrInsufficientBalance
		}
		le.AccountsSummary[entry.ToAccountID] = toSummary
	}
	return le, nil
}

type CreateTransaction struct {
	UniqueID string
	Entries  []MovementEntry
}

func (c CreateTransaction) validate() error {
	if c.UniqueID == "" {
		return ledger.ErrUniqueIDEmpty
	}
	if len(c.Entries) == 0 {
		return ledger.ErrEmptyEntries
	}
	return nil
}

// accounts retrurns the list of accounts from movement entries.
func (c CreateTransaction) accounts() []string {
	// Mapped out all the accounts inside the movement.
	accountsInEntries := make(map[string]bool)
	var accounts []string
	for _, e := range c.Entries {
		if _, ok := accountsInEntries[e.FromAccountID]; !ok {
			accounts = append(accounts, e.FromAccountID)
			accountsInEntries[e.FromAccountID] = true
		}
		if _, ok := accountsInEntries[e.ToAccountID]; !ok {
			accounts = append(accounts, e.ToAccountID)
			accountsInEntries[e.ToAccountID] = true
		}
	}
	return accounts
}

// Transact allows client or other apis in the same codebase to execute queries within one transaction(this also means it needs to be in the same database).
// Because everything is being done inside one transaction, the API provide strong consistency between modules and features. But with that being said, with this
// we also expose the posibilities of long-transactions and possibly locking the user balance row for quite some time. The user of the APIs need to understand this
// and ensure only codes that need strong consistency should exists within the function parameter.
//
// The function will automatically call Commit if erorr is nil and Rollback if error from the function parameter is not nil.
func (l *Ledger) Transact(ctx context.Context, tx CreateTransaction, fn func(ctx context.Context, pg *postgres.Postgres) error) error {
	if err := tx.validate(); err != nil {
		return err
	}
	balances, err := l.q.GetAccountsBalanceMappedByAccID(ctx, tx.accounts()...)
	if err != nil {
		return err
	}
	le, err := movementEntriesToLedgerEntries("", balances, tx.Entries...)
	if err != nil {
		return err
	}
	if err := l.q.WithTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *ledgerpg.Queries) error {
		if err := q.Move(ctx, le); err != nil {
			return err
		}
		errC := make(chan error, 1)
		// Call do in a subsequent concurrent call because we don't know how long Do will block.
		go func() {
			errC <- q.Do(ctx, fn)
		}()
		// Wait until the parent context is done or the do function returns with an error.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errC:
			return err
		}
	}); err != nil {
		return err
	}
	return nil
}

type checkEligible struct {
	FromAccountID     string
	ToAccountID       string
	FromAccountStatus int32
	ToAccountStatus   int32
	FromCurrency      *currency.Currency
	ToCurrency        *currency.Currency
}

// checkEligibleForMovement checks whether the movement is allowed.
func checkEligibleForMovement(ce checkEligible) error {
	if ce.FromAccountID == "" || ce.ToAccountID == "" {
		return ledger.ErrAccountSourceOrDestinationEmpty
	}
	// Prevent from transfering to self.
	if ce.FromAccountID == ce.ToAccountID {
		return ledger.ErrCannotMoveToSelf
	}
	// Prevent to transfering different currencies between account.
	if ce.FromCurrency.ID != ce.ToCurrency.ID {
		return ledger.ErrMismatchCurrencies
	}
	return nil
}
