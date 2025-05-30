// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package postgres

import (
	"database/sql"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type SecurityPermission struct {
	PermissionID    int64
	PermissionUuid  uuid.UUID
	PermissionName  string
	PermissionType  string
	PermissionKey   string
	PermissionValue string
	CreatedAt       time.Time
	UpdatedAt       sql.NullTime
}

type SecurityRole struct {
	RoleID    int64
	RoleUuid  uuid.UUID
	RoleName  string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type SecurityRolePermission struct {
	RoleID       int64
	PermissionID int64
	CreatedAt    time.Time
}

type User struct {
	UserID        int64
	UserUuid      uuid.UUID
	SecurityRoles []string
	CreatedAt     time.Time
	UpdatedAt     sql.NullTime
}

type UserPii struct {
	UserID         int64
	Email          string
	PhoneNumber    sql.NullString
	IdentityNumber sql.NullString
	IdentityType   sql.NullInt32
	CreatedAt      time.Time
	UpdatedAt      sql.NullTime
}

type UserSecret struct {
	SecretID             int64
	SecretUuid           uuid.UUID
	UserID               int64
	SecretKey            string
	SecretType           int32
	CurrentSecretVersion int64
	CreatedAt            time.Time
	UpdatedAt            sql.NullTime
}

type UserSecretVersion struct {
	SecretID      int64
	SecretVersion int64
	SecretValue   string
	SecretSalt    sql.NullString
	CreatedAt     time.Time
}

type UserSession struct {
	SessionID            uuid.UUID
	PreviousSesisionID   uuid.NullUUID
	SessionType          int32
	UserID               sql.NullInt64
	RandomID             string
	CreatedFromIp        netip.Addr
	CreatedFromLoc       sql.NullString
	CreatedFromUserAgent string
	SessionMetadata      []byte
	CreatedAt            time.Time
	ExpiredAt            time.Time
}
