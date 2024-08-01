package ledger

import "errors"

var (
	ErrAccountNotFound              = errors.New("account not found")
	ErrMismatchCurrencies           = errors.New("different currencies from source to destination")
	ErrUniqueIDEmpty                = errors.New("unique id is required")
	ErrEmptyEntries                 = errors.New("movement entries is required")
	ErrInsufficientBalance          = errors.New("insufficient balance")
	ErrCannotMoveToSelf             = errors.New("cannot move money to the same account")
	ErrForbiddenAccountTypeTransfer = errors.New("forbidden account type transfer")
)
