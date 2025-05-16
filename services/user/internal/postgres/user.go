package postgres

import (
	"context"
	"database/sql"
	"time"

	userv1 "github.com/studio-asd/go-example/proto/types/user/v1"
)

type RegisterUserWithPassword struct {
	UUID               string
	Email              string
	Password           string
	PasswordSecretKey  string
	PasswordSecretType int32
	CreatedAt          time.Time
}

func (q *Queries) RegisterUserWithPassword(ctx context.Context, user RegisterUserWithPassword) (int64, error) {
	var userID int64
	fn := func(ctx context.Context, q *Queries) error {
		var err error
		userID, err = q.CreateUser(ctx, CreateUserParams{
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
			UserID:    userID,
			Key:       user.PasswordSecretKey,
			Type:      int32(userv1.UserSecretType_USER_SECRET_TYPE_PASSWORD),
			Value:     user.Password,
			CreatedAt: user.CreatedAt,
		}); err != nil {
			return err
		}
		return nil
	}
	err := q.WithMetrics(ctx, "registerUserWithPassword", func(ctx context.Context, q *Queries) error {
		return q.ensureInTransact(ctx, sql.LevelDefault, fn)
	})
	return userID, err
}
