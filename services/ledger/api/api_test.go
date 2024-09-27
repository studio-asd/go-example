package api

import (
	"context"
	"testing"

	"github.com/albertwidi/pkg/postgres"
	"github.com/google/uuid"
)

func TestTransact(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	tq, err := testHelper.ForkPostgresSchema(context.Background(), testQueries, "public")
	if err != nil {
		t.Fatal(err)
	}
	tl := New(tq.Queries())

	newTableQuery := "CREATE TABLE IF NOT EXISTS trasact_test(id int PRIMARY KEY);"
	err = tq.Queries().Do(context.Background(), func(ctx context.Context, pg *postgres.Postgres) error {
		_, err := pg.Exec(ctx, newTableQuery)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("transact_success", func(t *testing.T) {
		t.Parallel()

		fn := func(ctx context.Context, pg *postgres.Postgres) error {
			insertQuery := "INSERT INTO transact_test VALUES(1);"
			_, err := pg.Exec(ctx, insertQuery)
			if err != nil {
				return err
			}
			return nil
		}

		tl.Transact(context.Background(), CreateTransaction{
			UniqueID: uuid.NewString(),
			Entries:  []MovementEntry{},
		}, fn)
	})

	t.Run("transact_failed", func(t *testing.T) {
		t.Parallel()
	})
}
