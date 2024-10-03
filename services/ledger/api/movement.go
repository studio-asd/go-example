package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/albertwidi/go-example/internal/currency"
	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
	"github.com/albertwidi/go-example/services/ledger"
	ledgerpg "github.com/albertwidi/go-example/services/ledger/internal/postgres"
)

// createLedgerEntries converts the initial movement entries to ledger entries, check and
// summarize them into a correct entries.
func createLedgerEntries(balances map[string]ledgerpg.GetAccountsBalanceRow, entries ...*ledgerv1.MovementEntry) (ledger.MovementLedgerEntries, error) {
	movementID := uuid.NewString()
	le := ledger.MovementLedgerEntries{
		MovementID: movementID,
		// Create the entries with length of (entries*2) as we need the entry of DEBIT and CREDIT for each movement.
		LedgerEntries:   make([]ledger.LedgerEntry, len(entries)*2),
		AccountsSummary: make(map[string]ledger.AccountMovementSummary),
	}

	createdAt := time.Now()
	for idx, entry := range entries {
		// Check whether we have the correct currencies from and to account as we don't want to mix the currencies in the transfer.
		currFrom, err := currency.Currencies.GetByID(balances[entry.GetFromAccountId()].CurrencyID)
		if err != nil {
			return ledger.MovementLedgerEntries{}, err
		}
		currTo, err := currency.Currencies.GetByID(balances[entry.GetToAccountId()].CurrencyID)
		if err != nil {
			return ledger.MovementLedgerEntries{}, err
		}
		if err := checkEligibleForMovement(checkEligible{
			FromAccountID: entry.GetFromAccountId(),
			ToAccountID:   entry.GetToAccountId(),
			FromCurrency:  currFrom,
			ToCurrency:    currTo,
		}); err != nil {
			return ledger.MovementLedgerEntries{}, fmt.Errorf("%w: please check entry at index [%d]", err, idx)
		}

		dec, err := decimal.NewFromString(entry.GetAmount())
		if err != nil {
			return ledger.MovementLedgerEntries{}, err
		}
		// Normalize the amount based on the currency, because the exponent might be more than expected.
		amount := currFrom.NormalizeDecimal(dec)
		// As we have two entries in the ledger in every movement entry, the starting index will be always idx*2.
		arrIdx := idx * 2
		// sequence is always stats from 1, this is why we add the idx(which began with 0) with 1.
		sequence := idx + 1

		// Create the DEBIT record.
		debitAmount := amount.Mul(decimal.NewFromInt(-1))
		// The ledger_id is a UUIDV5 with namespace_oid and format of: movement_id:from_account_id:sequence.
		debitLedgerID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(movementID+":"+entry.GetFromAccountId()+":"+strconv.Itoa(sequence))).String()
		le.LedgerEntries[arrIdx] = ledger.LedgerEntry{
			LedgerID:         debitLedgerID,
			MovementID:       movementID,
			AccountID:        entry.GetFromAccountId(),
			Amount:           debitAmount,
			MovementSequence: sequence,
			CreatedAt:        createdAt,
			Timestamp:        createdAt.Unix(),
		}

		// Create the CREDIT record.
		// The ledger_id is a UUIDV5 with namespace_oid and format of: movement_id:from_account:sequence.
		creditLedgerID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(movementID+":"+entry.GetToAccountId()+":"+strconv.Itoa(sequence))).String()
		le.LedgerEntries[arrIdx+1] = ledger.LedgerEntry{
			LedgerID:         creditLedgerID,
			MovementID:       movementID,
			AccountID:        entry.GetToAccountId(),
			Amount:           amount,
			MovementSequence: sequence,
			CreatedAt:        createdAt,
			Timestamp:        createdAt.Unix(),
		}
		// Add the accounts to the list of accounts.
		if _, ok := le.AccountsSummary[entry.GetFromAccountId()]; !ok {
			le.Accounts = append(le.Accounts, entry.GetFromAccountId())
		}
		if _, ok := le.AccountsSummary[entry.GetToAccountId()]; !ok {
			le.Accounts = append(le.Accounts, entry.GetToAccountId())
		}

		// Add each account to the summary.
		// Sender account.
		var fromSummary ledger.AccountMovementSummary
		if summary, ok := le.AccountsSummary[entry.GetFromAccountId()]; ok {
			fromSummary = summary
		} else {
			fromSummary = ledger.AccountMovementSummary{
				LastLedgerID:  balances[entry.GetFromAccountId()].LastLedgerID,
				EndingBalance: balances[entry.GetFromAccountId()].Balance,
			}
		}
		fromSummary.NextLedgerID = debitLedgerID
		fromSummary.BalanceChanges = fromSummary.BalanceChanges.Add(debitAmount)
		fromSummary.EndingBalance = fromSummary.EndingBalance.Add(debitAmount)
		// Check whether the ending balance is negative, we cannot allow negative balance for most the accounts.
		if fromSummary.EndingBalance.IsNegative() && !balances[entry.GetFromAccountId()].AllowNegative {
			return ledger.MovementLedgerEntries{}, ledger.ErrInsufficientBalance
		}
		le.AccountsSummary[entry.GetFromAccountId()] = fromSummary

		// Receiver account.
		var toSummary ledger.AccountMovementSummary
		if summary, ok := le.AccountsSummary[entry.GetToAccountId()]; ok {
			toSummary = summary
		} else {
			toSummary = ledger.AccountMovementSummary{
				LastLedgerID:  balances[entry.GetToAccountId()].LastLedgerID,
				EndingBalance: balances[entry.GetToAccountId()].Balance,
			}
		}
		toSummary.NextLedgerID = creditLedgerID
		toSummary.BalanceChanges = toSummary.BalanceChanges.Add(amount)
		toSummary.EndingBalance = toSummary.EndingBalance.Add(amount)
		// Check whether the ending balance is negative, we cannot allow negative balance for most the accounts.
		if toSummary.EndingBalance.IsNegative() && !balances[entry.GetToAccountId()].AllowNegative {
			return ledger.MovementLedgerEntries{}, ledger.ErrInsufficientBalance
		}
		le.AccountsSummary[entry.GetToAccountId()] = toSummary
	}
	return le, nil
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
