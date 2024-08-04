package service

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	ledgerpg "github.com/albertwidi/go-example/ledger/postgres"
)

var (
	// All variables below this only available if '-short' is not used, this means we will do integration test.
	testLedger  *Ledger
	testQueries *ledgerpg.Queries
	testHelper  *ledgerpg.TestHelper
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
	if !testing.Short() {
		var err error
		testHelper = ledgerpg.NewTestHelper()
		testQueries, err = testHelper.PrepareTest(context.Background())
		if err != nil {
			return 1, err
		}
		defer func() {
			closeErr := testHelper.Close()
			if closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
	}
	code = m.Run()
	return
}
