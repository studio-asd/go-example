// Code is generated by helper script. DO NOT EDIT.
//
// sqlc_version    : v1.28.0
// sqlc_config     : sqlc.yaml
// sqlc_sql_package: pgx/v5
// database        : user_data
// generated_time  : 2025-03-26T23:55:27+07:00

package postgres

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/studio-asd/go-example/internal/testing/pghelper"
	schema "github.com/studio-asd/go-example/database/schemas/user_data"
)

var (
	testCtx context.Context
	testQueries *Queries
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
	   DatabaseName: schema.DatabaseName,
	   EmbeddedSchema: schema.EmbeddedSchema,
	})
	if err != nil {
		code = 1
		return
	}
	testQueries = New(testHelper.Postgres())

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
