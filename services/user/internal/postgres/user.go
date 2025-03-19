package postgres

import (
	"context"
	"database/sql"
	"time"
)

type RegisterUser struct {
	UUID      string
	Email     string
	Password  string
	CreatedAt time.Time
}

func (q *Queries) RegisterUser(ctx context.Context, user RegisterUser) error {
	fn := func(ctx context.Context, q *Queries) error {
		_, err := q.CreateUser(ctx, CreateUserParams{
			ExternalID: user.UUID,
			UserEmail:  user.Email,
			CreatedAt:  user.CreatedAt,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return q.WithMetrics(ctx, "registerUser", func(ctx context.Context, q *Queries) error {
		return q.ensureInTransact(ctx, sql.LevelDefault, fn)
	})
}
