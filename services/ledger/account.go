package ledger

const (
	// Account types.
	AccountTypeDeposit    = "deposit"
	AccountTypeWithdrawal = "withdrawal"
	AccountTypeUser       = "user"
	// Account statuses.
	AccountStatusActive   = "active"
	AccountStatusInactive = "inactive"
)

type AccountInfo struct {
	AccountID       string
	ParentAccountID string
	AllowNegative   bool
}
