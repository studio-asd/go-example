package postgres

import (
	"context"
	"database/sql"
	"time"
)

type CreateNewSecret struct {
	ExternalID string
	UserID     int64
	Key        string
	Value      string
	Salt       string
	Type       int32
	CreatedAt  time.Time
}

// CreateNewSecret creartes a new secret for the user with version of one(1).
func (q *Queries) CreateNewSecret(ctx context.Context, new CreateNewSecret) error {
	fn := func(ctx context.Context, q *Queries) error {
		secretID, err := q.CreateUserSecret(ctx, CreateUserSecretParams{
			ExternalID: new.ExternalID,
			UserID:     new.UserID,
			SecretKey:  new.Key,
			SecretType: new.Type,
			// A new secret will always have version one(1).
			CurrentSecretVersion: 1,
			CreatedAt:            new.CreatedAt,
		})
		if err != nil {
			return err
		}

		secretSalt := sql.NullString{}
		if new.Salt != "" {
			secretSalt = sql.NullString{String: new.Salt, Valid: true}
		}
		if err := q.CreateUserSecretVersion(ctx, CreateUserSecretVersionParams{
			SecretID: secretID,
			// A new secret will always have version one(1).
			SecretVersion: 1,
			SecretValue:   new.Value,
			SecretSalt:    secretSalt,
			CreatedAt:     new.CreatedAt,
		}); err != nil {
			return err
		}
		return nil
	}
	return q.WithMetrics(ctx, "createNewSecret", func(ctx context.Context, q *Queries) error {
		return q.ensureInTransact(ctx, sql.LevelDefault, fn)
	})
}
