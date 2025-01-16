// Code is generated by helper script. DO NOT EDIT.
//
// sqlc_version    : v1.27.0
// sqlc_config     : sqlc.yaml
// sqlc_sql_package: pgx/v5
// database        : go_example
// generated_time  : 2025-01-16T13:59:43+07:00

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
	db "github.com/studio-asd/go-example/database/schemas/go-example"
)

var (
	testCtx context.Context
	testHelper *pghelper.Helper[*Queries]
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
	dbName := "go_example"
	testHelper, err = pghelper.New(ctx, pghelper.Config{
	   DatabaseName: dbName,
	   EmbeddedSchema: db.EmbeddedSchema,
	}, New)
	if err != nil {
		code = 1
		return
	}
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
