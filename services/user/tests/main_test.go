package user

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	schema "github.com/studio-asd/go-example/database/schemas/user-data"
	"github.com/studio-asd/go-example/internal/testing/pghelper"
	"github.com/studio-asd/go-example/services/user/api"
)

var (
	testCtx    context.Context
	testAPI    *api.API
	testHelper *pghelper.Helper
)

func TestMain(m *testing.M) {
	flag.Parse()
	// Don't invoke the integration test if short flag is used.
	if testing.Short() {
		return
	}

	var cancel context.CancelFunc
	testCtx, cancel = context.WithTimeout(context.Background(), time.Minute*5)
	code, err := run(testCtx, m)
	if err != nil {
		log.Println(err)
	}
	cancel()
	os.Exit(code)
}

func run(ctx context.Context, m *testing.M) (code int, err error) {
	testHelper, err = pghelper.New(ctx, pghelper.Config{
		DatabaseName:   schema.DatabaseName,
		EmbeddedSchema: schema.EmbeddedSchema,
	})
	if err != nil {
		code = 1
		return
	}
	testAPI = api.New(testHelper.Postgres())

	// Close all resources upon exit, and record the error when closing the resources if any.
	defer func() {
		errClose := testHelper.Close()
		if errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	code = m.Run()
	return
}

// setup setups all the records and data needed for the tests to run.
func setup() error {
}
