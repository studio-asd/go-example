package api

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/albertwidi/go-example/internal/testing/pghelper"
	ledgerpg "github.com/albertwidi/go-example/services/ledger/internal/postgres"
)

var (
	// All variables below this only available if '-short' is not used, this means we will do integration test.
	testAPI     *API
	testQueries *ledgerpg.Queries
	testHelper  *pghelper.Helper[*ledgerpg.Queries]
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
		testHelper, err = pghelper.New(context.Background(), "ledger_api", ledgerpg.New)
		if err != nil {
			return
		}
		defer func() {
			closeErr := testHelper.Close()
			if closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
		testQueries = testHelper.Queries()
		testAPI = New(testHelper.Queries())
	}
	code = m.Run()
	return
}
