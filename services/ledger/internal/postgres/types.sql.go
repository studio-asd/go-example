// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package postgres

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
)

func (e *AccountStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AccountStatus(s)
	case string:
		*e = AccountStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for AccountStatus: %T", src)
	}
	return nil
}

type NullAccountStatus struct {
	AccountStatus AccountStatus
	Valid         bool // Valid is true if AccountStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAccountStatus) Scan(value interface{}) error {
	if value == nil {
		ns.AccountStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AccountStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAccountStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.AccountStatus), nil
}

type Account struct {
	AccountID       string
	ParentAccountID string
	AccountStatus   AccountStatus
	CurrencyID      int32
	CreatedAt       time.Time
	UpdatedAt       sql.NullTime
}

type AccountsBalance struct {
	AccountID      string
	CurrencyID     int32
	AllowNegative  bool
	Balance        decimal.Decimal
	LastMovementID string
	LastLedgerID   string
	CreatedAt      time.Time
	UpdatedAt      sql.NullTime
}

type AccountsBalanceHistory struct {
	MovementID         string
	AccountID          string
	Balance            decimal.Decimal
	PreviousBalance    decimal.Decimal
	PreviousMovementID string
	CreatedAt          time.Time
}

type AccountsLedger struct {
	LedgerID         string
	MovementID       string
	AccountID        string
	MovementSequence int32
	CurrencyID       int32
	Amount           decimal.Decimal
	PreviousLedgerID string
	CreatedAt        time.Time
	Timestamp        int64
	ClientID         sql.NullString
}

type Movement struct {
	MovementID     string
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      sql.NullTime
}
