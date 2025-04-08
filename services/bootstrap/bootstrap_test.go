package bootstrap

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/studio-asd/pkg/postgres"

	"github.com/studio-asd/go-example/internal/testing/pghelper"
)

var (
	goExampleDB *postgres.Postgres
	userDB      *postgres.Postgres
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	goExampleHelper, err := pghelper.New(context.Background(), pghelper.Config{
		DatabaseName: "go_example",
		SkipPrepare:  true,
	})
	if err != nil {
		return 1, err
	}
	goExampleDB = goExampleHelper.Postgres()

	userHelper, err := pghelper.New(context.Background(), pghelper.Config{
		DatabaseName: "user_data",
		SkipPrepare:  true,
	})
	if err != nil {
		return 1, err
	}
	userDB = userHelper.Postgres()
	return m.Run(), nil
}

func TestBootstrapFromZero(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	b, err := New(Params{
		GoExampleDB: goExampleDB,
		UserDB:      userDB,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Execute(t.Context(), ExecuteParams{All: true}); err != nil {
		t.Fatal(err)
	}
}
