// pghelper is a package to provide a test struct/object for postgres database. The package is design to match the function signature
// of intrernal postgres package in this project.

package pghelper

import (
	"context"
	"embed"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/studio-asd/pkg/postgres"
	"github.com/studio-asd/pkg/testing/pgtest"

	"github.com/studio-asd/go-example/internal/env"
)

// PGTEST_SKIP_PREPARE is used to skip the creation of the database and applying all schemas into it. The environment varialbe
// is useful to skip the preparation when a multiple packages tests is being performed.
var skipPrepareTest = os.Getenv("PGTEST_SKIP_PREPARE")

// PGQuery interface defines the type that use this package should implements PGQuery.
type PGQuery interface {
	// Postgres returns the postgres connection object from github.com/studio-asd/pkg/postgres. We need the connection
	// because we will create the query object from the connection.
	Postgres() *postgres.Postgres
}

type Config struct {
	DatabaseName string
	// EmbeddedSchema is used to embed the database schema for test preparation purposes. The embedded schema is used
	// because it is more deterministic rather than guessing about where the schema files are being located when
	// the tests runs.
	EmbeddedSchema embed.FS
	// DontDropOnClose allows/prevents the helper to drop the database when closing all the connections
	// in the test helper.
	//
	// This option is helpful for debugging in case we want to connect directly to the PostgreSQL database.
	DontDropOnClose bool
	// SkipPrepare is used to skip the creation of the database and applying all schemas into it. The environment varialbe
	// is useful to skip the preparation when a multiple packages tests is being performed.
	SkipPrepare bool
}

func (c Config) validate() error {
	if c.DatabaseName == "" {
		return errors.New("database name cannot be empty")
	}
	return nil
}

type Helper struct {
	config       Config
	conn         *postgres.Postgres
	pgtestHelper *pgtest.PGTest

	mu sync.Mutex
	// forks is the list of forked helper throughout the test. We need to track the lis of forked helper as we want
	// to track the resource of helper and close them properly.
	forks []*Helper
	// fork is a mark that the test helper had been forked, thus several expections should be made when
	// doing several operation like closing connections.
	fork   bool
	closed bool
}

func New(ctx context.Context, config Config) (*Helper, error) {
	if !testing.Testing() {
		return nil, errors.New("can only be used in test")
	}
	th := &Helper{
		config:       config,
		pgtestHelper: pgtest.New(),
	}

	var (
		pg  *postgres.Postgres
		err error
		// TEST_PG_DSN can be used to set different DSN for flexible test setup.
		pgDSN = env.GetEnvOrDefault("TEST_PG_DSN", "postgres://postgres:postgres@localhost:5432/"+config.DatabaseName+"?sslmode=disable")
	)
	if !SkipPrepare(config.SkipPrepare) {
		pg, err = prepareTest(ctx, pgDSN, config.EmbeddedSchema)
		if err != nil {
			return nil, err
		}
	} else {
		config, err := postgres.NewConfigFromDSN(pgDSN)
		if err != nil {
			return nil, err
		}
		if err := pgtest.CreateDatabase(ctx, pgDSN, false); err != nil {
			return nil, err
		}
		pg, err = postgres.Connect(ctx, config)
		if err != nil {
			return nil, err
		}
	}
	th.conn = pg
	return th, nil
}

func (th *Helper) Postgres() *postgres.Postgres {
	th.mu.Lock()
	defer th.mu.Unlock()
	return th.conn
}

// Close closes all connections from the test helper.
func (th *Helper) Close() error {
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.closed {
		return nil
	}

	var err error
	if th.conn != nil {
		errClose := th.conn.Close()
		if errClose != nil {
			err = errors.Join(err, errClose)
		}
	}
	// If not a fork, then we should close all the connections in the test helper as it will closes all connections
	// to the forked schemas. But in fork, we should avoid this as we don't want to control this from forked test helper.
	if !th.fork {
		errClose := th.pgtestHelper.Close()
		if errClose != nil {
			errors.Join(err, errClose)
		}
		// Closes all the forked helper, this closes the postgres connection in each helper.
		for _, forkedHelper := range th.forks {
			if err := forkedHelper.Close(); err != nil {
				return err
			}
		}
		if !th.config.DontDropOnClose {
			// Drop the database after test so we will always have a fresh database when we start the test.
			config := th.conn.Config()
			dsn, err := config.DSN()
			if err != nil {
				return err
			}
			err = pgtest.DropDatabase(context.Background(), dsn.URL())
			if err != nil {
				return err
			}
		}
	}
	if err == nil {
		th.closed = true
	}
	return err
}

// ForkPostgresSchema forks the sourceSchema with the underlying connection inside the Queries. The function will return a new connection
// with default search_path into the new schema. The schema name currently is random and cannot be defined by the user.
func (th *Helper) ForkPostgresSchema(ctx context.Context, pg *postgres.Postgres, schemaName string) (*Helper, error) {
	if th.fork {
		return nil, errors.New("cannot fork the schema from a forked test helper, please use the original test helper")
	}
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.closed {
		return nil, errors.New("cannot create a fork from closed test helper")
	}
	pg, err := th.pgtestHelper.ForkSchema(ctx, pg, schemaName)
	if err != nil {
		return nil, err
	}
	newTH := &Helper{
		config:       th.config,
		conn:         pg,
		pgtestHelper: th.pgtestHelper,
		fork:         true,
	}
	// Append the forks to the origin
	th.forks = append(th.forks, newTH)
	return newTH, nil
}

// DefaultSearchPath returns the default PostgreSQL search path. This helper function invoke the pg.DefaultSearchPath
// to do this. This function added to avoid the user/client to go deeper to the postgres object to invoke this function.
func (th *Helper) DefaultSearchPath() string {
	return th.conn.DefaultSearchPath()
}

// CloseFunc is a helper function to close the test helper via testing.T.Cleanup.
func (th *Helper) CloseFunc(t *testing.T) func() {
	return func() {
		if err := th.Close(); err != nil {
			t.Log(err)
		}
	}
}

// prepareTest prepares the designated postgres database by creating the database and applying the schema. The function returns a postgres connection
// to the database that can be used for testing purposes.
func prepareTest(ctx context.Context, pgDSN string, embeddedSchema embed.FS) (*postgres.Postgres, error) {
	// TEST_PG_DSN can be used to set different DSN for flexible test setup.
	if err := pgtest.CreateDatabase(ctx, pgDSN, true); err != nil {
		return nil, err
	}

	if reflect.ValueOf(embeddedSchema).IsZero() {
		return nil, errors.New("embedded schema is empty")
	}

	embeddedFS, err := iofs.New(embeddedSchema, "migrations")
	if err != nil {
		return nil, err
	}
	mg, err := migrate.NewWithSourceInstance("iofs", embeddedFS, pgDSN)
	if err != nil {
		return nil, err
	}
	if err := mg.Up(); err != nil {
		return nil, err
	}
	sourceErr, dbErr := mg.Close()
	if sourceErr != nil || dbErr != nil {
		errs := errors.Join(dbErr, sourceErr)
		return nil, errs
	}
	//
	// Create a new connection with the correct database name.
	config, err := postgres.NewConfigFromDSN(pgDSN)
	if err != nil {
		return nil, err
	}
	// Connect to the PostgreSQL with the configuration.
	testConn, err := postgres.Connect(context.Background(), config)
	if err != nil {
		return nil, err
	}
	return testConn, nil
}

// SkipPrepare returns the value of PGTEST_SKIP_PREPARE to decide whether test preparation need to be skipped or not.
// It is rather useful for other package to understand this as they may want to use a separate database name or
// using other configurations if test preparation is not skipped.
func SkipPrepare(skipPrepareFlag bool) bool {
	if skipPrepareTest == "1" {
		return true
	}
	return skipPrepareFlag
}
