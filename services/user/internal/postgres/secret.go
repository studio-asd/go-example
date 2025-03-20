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
	Type       int32
	CreatedAt  time.Time
}

// CreateNewSecret creartes a new secret for the user with version of one(1).
func (q *Queries) CreateNewSecret(ctx context.Context, new CreateNewSecret) error {
	// A new secret will always have version one(1).
	secretVersion := 1
	fn := func(ctx context.Context, q *Queries) error {
		secretID, err := q.CreateUserSecret(ctx, CreateUserSecretParams{
			ExternalID:           new.ExternalID,
			UserID:               new.UserID,
			SecretKey:            new.Key,
			SecretType:           new.Type,
			CurrentSecretVersion: int64(secretVersion),
			CreatedAt:            new.CreatedAt,
		})
		if err != nil {
			return err
		}
		if err := q.CreateUserSecretVersion(ctx, CreateUserSecretVersionParams{
			SecretID:      secretID,
			SecretVersion: int64(secretVersion),
			SecretValue:   new.Value,
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
