package api

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	schema "github.com/studio-asd/go-example/database/schemas/go-example"
	"github.com/studio-asd/go-example/internal/testing/pghelper"
	ledgerpg "github.com/studio-asd/go-example/services/ledger/internal/postgres"
)

var (
	// All variables below this only available if '-short' is not used, this means we will do integration test.
	testAPI     *API
	testQueries *ledgerpg.Queries
	testHelper  *pghelper.Helper
)

func TestMain(m *testing.M) {
	flag.Parse()
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {
	defer func() {
		if err != nil {
			code = 1
		}
	}()

	if !testing.Short() {
		dbName := "go_example"
		// Use a different database name if we are not in the global test mode.
		if !pghelper.SkipPrepare(true) {
			dbName = "ledger_api"
		}
		testHelper, err = pghelper.New(context.Background(), pghelper.Config{
			DatabaseName:   dbName,
			EmbeddedSchema: schema.EmbeddedSchema,
		})
		if err != nil {
			return
		}
		defer func() {
			closeErr := testHelper.Close()
			if closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
		testQueries = ledgerpg.New(testHelper.Postgres())
		testAPI = New(testHelper.Postgres())
	}
	code = m.Run()
	return
}
