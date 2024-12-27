package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/albertwidi/pkg/postgres"
)

var testPG *postgres.Postgres

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	var err error
	testPG, err = postgres.Connect(context.Background(), postgres.ConnectConfig{
		Driver:   "pgx",
		Username: "postgres",
		Password: "postgres",
	})
	if err != nil {
		return 1, err
	}
	return m.Run(), nil
}

func TestTransactExec(t *testing.T) {
	createTableQuery := `
CREATE TABLE something_1(internal_id INT PRIMARY KEY);
CREATE TABLE something_2(internal_id INT PRIMARY KEY);
`
	_, err := testPG.Exec(context.Background(), createTableQuery)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("already_in_transaction", func(t *testing.T) {
		err := testPG.Transact(context.Background(), sql.LevelReadCommitted, func(ctx context.Context, pg *postgres.Postgres) error {
			err := NewTransactExec(context.Background(), pg, sql.LevelReadCommitted).Do(nil, nil, time.Second)
			return err
		})
		if err == nil {
			t.Fatal("should get already in transaction error")
		}
	})
	t.Run("first_function", func(t *testing.T) {

	})
}
