package service

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	ledgerpg "github.com/albertwidi/go-example/services/ledger/internal/postgres"
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
	defer func() {
		if err != nil {
			code = 1
		}
	}()

	if !testing.Short() {
		testHelper, err = ledgerpg.NewTestHelper(context.Background())
		if err != nil {
			return
		}
		defer func() {
			closeErr := testHelper.Close()
			if closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
		testLedger = New(testQueries)
	}
	code = m.Run()
	return
}
