package postgres

import (
	"context"
	"database/sql"
	"time"

	userv1 "github.com/studio-asd/go-example/proto/types/user/v1"
)

type RegisterUser struct {
	UUID             string
	Email            string
	Password         string
	PasswordSecretID string
	CreatedAt        time.Time
}

func (q *Queries) RegisterUser(ctx context.Context, user RegisterUser) error {
	fn := func(ctx context.Context, q *Queries) error {
		userID, err := q.CreateUser(ctx, CreateUserParams{
			ExternalID: user.UUID,
			CreatedAt:  user.CreatedAt,
		})
		if err != nil {
			return err
		}
		if err := q.CreateUserPII(ctx, CreateUserPIIParams{
			UserID:    userID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		}); err != nil {
			return err
		}
		if err := q.CreateNewSecret(ctx, CreateNewSecret{
			ExternalID: user.PasswordSecretID,
			UserID:     userID,
			Key:        "user_password",
			Type:       int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
			Value:      user.Password,
			CreatedAt:  user.CreatedAt,
		}); err != nil {
			return err
		}
		return nil
	}
	return q.WithMetrics(ctx, "registerUser", func(ctx context.Context, q *Queries) error {
		return q.ensureInTransact(ctx, sql.LevelDefault, fn)
	})
}
