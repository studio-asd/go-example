package service

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/albertwidi/pkg/postgres"

	ledgerpg "github.com/albertwidi/go-example/ledger/postgres"
)

var (
	// All variables below this only available if '-short' is not used, this means we will do integration test.
	testLedger  *Ledger
	testQueries *ledgerpg.Queries
	testPG      *postgres.Postgres
)

func TestMain(m *testing.M) {
	flag.Parse()
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	if !testing.Short() {
		pg, err := ledgerpg.PrepareTest(context.Background())
		if err != nil {
			return 1, err
		}
		testPG = pg
		testQueries = ledgerpg.New(pg)
		testLedger = New(pg)
	}

	code := m.Run()
	return code, nil
}
