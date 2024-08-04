// This tiny Go program is a tiny generator to automatically generate the test and helper function
// for the database layer.

package main

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"

	"github.com/albertwidi/pkg/postgres"
)

//go:embed sqlc.yaml
var embeddedFiles embed.FS

type Flags struct {
	All        bool
	DBName     string
	SQLCConfig string
}

// SQLCConfig is the sqlc configuration structure to be used in this generator.
type SQLCConfig struct {
	fileName string
	Version  string          `yaml:"version"`
	SQL      []SQLCSQLConfig `yaml:"sql"`
}

type SQLCSQLConfig struct {
	Schema  string `yaml:"schema"`
	Queries string `yaml:"queries"`
	Engine  string `yaml:"engine"`
	Gen     struct {
		Go SQLCGenGo `yaml:"go"`
	} `yaml:"gen"`
	Database struct {
		URI string `yaml:"uri"`
	} `yaml:"database"`
}

type SQLCGenGo struct {
	Package              string `yaml:"package"`
	Out                  string `yaml:"out"`
	SQLPackage           string `yaml:"sql_package"`
	OutputDBFileName     string `yaml:"output_db_file_name"`
	OutputModelsFileName string `yaml:"output_models_file_name"`
}

type TemplateData struct {
	DatabaseName       string
	SQLCVersion        string
	SQLCConfig         string
	SQLCOutputFileName string
	GoPackageName      string
	SQLPackageName     string
	PathToSchema       string
	DatabaseConn       TemplateDataDatabaseConn
}

type TemplateDataDatabaseConn struct {
	DatabaseName string
	Host         string
	Port         string
	Username     string
	Password     string
}

func main() {
	if len(os.Args) < 2 {
		panic("missing command in the arguments")
	}
	err := run(os.Args[1:])
	if err != nil {
		panic(err)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return errors.New("command is empty")
	}

	// Check if we have more than one(1) args which is the command. In general we only use one(1) input after the
	// command, for example main.go gengo [db_name]. Thus everything after [db_name] are flags.
	var flagArgs []string
	if len(args) > 1 {
		flagArgs = args[1:]
	}
	flags, err := parseFlags(flagArgs)
	if err != nil {
		return err
	}

	composeLogs := bytes.NewBuffer(nil)
	composeUpFunc := func(ctx context.Context) error {
		execComposeUp := exec.CommandContext(ctx, "docker", "compose", "up", "-d")
		execComposeUp.Stdout = os.Stdout
		execComposeUp.Stderr = os.Stderr

		err := execComposeUp.Run()
		if err == nil {
			execComposeLogs := exec.CommandContext(ctx, "docker", "compose", "logs", "-f", "--no-color")
			execComposeLogs.Stdout = composeLogs
			execComposeLogs.Stderr = composeLogs
			go execComposeLogs.Run()
		}
		return err
	}
	composeDownFunc := func(ctx context.Context) error {
		execComposeDown := exec.CommandContext(ctx, "docker", "compose", "down", "--remove-orphans")
		execComposeDown.Stdout = os.Stdout
		execComposeDown.Stderr = os.Stderr
		return execComposeDown.Run()
	}

	switch args[0] {
	case "gengo":
		var dbName string

		if len(args) > 1 {
			dbName = args[1]
		}
		databases, err := DatabaseList(dbName, flags, ".")
		if err != nil {
			return err
		}

		err = executeCommand(
			context.Background(),
			func(ctx context.Context) error {
				if err := composeUpFunc(ctx); err != nil {
					return err
				}
				// Wait for the postgres to fully up.
				time.Sleep(time.Second * 2)
				return nil
			},
			func(ctx context.Context) error {
				currentDir, err := os.Getwd()
				if err != nil {
					return err
				}
				for _, db := range databases {
					// Create the database first, as we will need to connect to the database when invoking sqlc
					// to generate go codes.
					fmt.Printf("Creating Database %s\n", db)
					err = execQuery(
						"postgres://postgres:postgres@localhost:5432?sslmode=disable",
						fmt.Sprintf("CREATE DATABASE %s;", db),
					)
					if err != nil {
						return fmt.Errorf("failed to create database: %w", err)
					}

					// Applying schema to the database, we need to peek into the sqlc configuration for the schema name.
					fmt.Println("Applying schema...")
					sqlcConfig := SQLCConfig{
						fileName: flags.SQLCConfig,
					}
					out, err := os.ReadFile(filepath.Join(db, flags.SQLCConfig))
					if err != nil {
						return err
					}
					if err := yaml.Unmarshal(out, &sqlcConfig); err != nil {
						return err
					}
					out, err = os.ReadFile(filepath.Join(db, sqlcConfig.SQL[0].Schema))
					if err != nil {
						return err
					}
					err = execQuery(
						fmt.Sprintf("postgres://postgres:postgres@localhost:5432/%s?sslmode=disable", db),
						string(out),
					)
					if err != nil {
						return err
					}

					if err := os.Chdir(filepath.Join(currentDir, db)); err != nil {
						return fmt.Errorf("failed to change directory to the database dir: %w", err)
					}
					sqlcExec := exec.Command("sqlc", "generate", "-f", flags.SQLCConfig)
					sqlcExec.Stdout = os.Stdout
					sqlcExec.Stderr = os.Stderr
					if err := sqlcExec.Run(); err != nil {
						return fmt.Errorf("failed to execute sqlc: %w", err)
					}
					if err := genTemplate(db, sqlcConfig); err != nil {
						return err
					}
				}
				return nil
			},
			func(ctx context.Context) error {
				var err error
				if errDown := composeDownFunc(ctx); errDown != nil {
					err = errors.Join(err, errDown)
				}
				return err
			},
		)
		if err != nil {
			fmt.Println(composeLogs.String())
			return err
		}

	case "copyconf":
		var dbName string

		if len(args) > 1 {
			dbName = args[1]
		}
		databases, err := DatabaseList(dbName, flags, ".")
		if err != nil {
			return err
		}
		out, err := embeddedFiles.ReadFile("sqlc.yaml")
		if err != nil {
			return err
		}

		for _, db := range databases {
			sqlcConfig := bytes.ReplaceAll(out, []byte("database_name"), []byte(db))
			if err := os.WriteFile(filepath.Join(db, "sqlc.yaml"), sqlcConfig, 0o666); err != nil {
				return err
			}
		}
	}
	return err
}

// executeCommand is a helper function to execute setup, the command and teardown function.
func executeCommand(ctx context.Context, setup func(ctx context.Context) error, execute func(context.Context) error, teardown func(context.Context) error) error {
	defer func() {
		if teardown != nil {
			teardown(ctx)
		}
	}()
	if setup != nil {
		if err := setup(ctx); err != nil {
			return err
		}
	}
	return execute(ctx)
}

// parseFlags parse the strinbrew updategs arguments from(for example) os.Args and returns the Flags struct.
func parseFlags(args []string) (f Flags, err error) {
	// Parse all the flags needed for code generation.
	fset := flag.NewFlagSet("global_flags", flag.ExitOnError)
	fset.BoolVar(&f.All, "all", false, "all flag to decide whether we want to generates for all or not")
	fset.StringVar(&f.DBName, "db_name", "", "database name")
	fset.StringVar(&f.SQLCConfig, "sqlc_config", "sqlc.yaml", "sqlc.yaml configuration")

	if err = fset.Parse(args); err != nil {
		return
	}
	return
}

func genTemplate(dbName string, config SQLCConfig) error {
	fmt.Printf("Generating template for database %s\n", dbName)
	// Retrieve the version of sqlc via sqlc CLI.
	cmd := exec.Command("sqlc", "version")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	sqlcVersion := string(out)
	sqlcVersion = strings.TrimRight(sqlcVersion, "\n")

	// Check whether the sqlc files are already generated, we want to replace the generated sqlc.go with our template.
	relativePath := config.SQL[0].Gen.Go.Out
	filePath := filepath.Join(relativePath, config.SQL[0].Gen.Go.OutputDBFileName)
	f, err := os.OpenFile(filePath, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.New("sqlc_template")
	tmpl, err = tmpl.Parse(sqlcTemplate)
	if err != nil {
		return err
	}
	tmpl = tmpl.New("db_test")
	tmpl, err = tmpl.Parse(dbTest)
	if err != nil {
		return err
	}
	tmpl = tmpl.New("test_helper")
	tmpl, err = tmpl.Parse(testHelper)
	if err != nil {
		return err
	}

	dsn, err := postgres.ParseDSN(config.SQL[0].Database.URI)
	if err != nil {
		return err
	}

	td := TemplateData{
		DatabaseName:       dbName,
		SQLCVersion:        sqlcVersion,
		SQLCConfig:         config.fileName,
		SQLCOutputFileName: config.SQL[0].Gen.Go.OutputDBFileName,
		GoPackageName:      config.SQL[0].Gen.Go.Package,
		SQLPackageName:     config.SQL[0].Gen.Go.SQLPackage,
		PathToSchema:       filepath.Join("database", dbName, config.SQL[0].Schema),
		DatabaseConn: TemplateDataDatabaseConn{
			Username:     dsn.Username,
			Password:     dsn.Password,
			Host:         dsn.Host,
			Port:         dsn.Port,
			DatabaseName: dsn.DatabaseName,
		},
	}

	buff := bytes.NewBuffer(nil)
	if err := tmpl.ExecuteTemplate(buff, "sqlc_template", td); err != nil {
		return err
	}
	_, err = f.Write(buff.Bytes())
	if err != nil {
		return err
	}
	buff.Reset()

	if err := tmpl.ExecuteTemplate(buff, "db_test", td); err != nil {
		return err
	}
	sqlcTestPath := filepath.Join(relativePath, "sqlc_test.go")
	if err := os.WriteFile(sqlcTestPath, buff.Bytes(), 0o666); err != nil {
		return err
	}
	buff.Reset()

	if err := tmpl.ExecuteTemplate(buff, "test_helper", td); err != nil {
		return err
	}
	testHelperPath := filepath.Join(relativePath, "test_helper.go")
	if err := os.WriteFile(testHelperPath, buff.Bytes(), 0o666); err != nil {
		return err
	}

	return nil
}

func execQuery(dsn, query string, args ...any) error {
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := sqldb.Ping(); err != nil {
		return err
	}
	_, err = sqldb.Exec(query, args...)
	return err
}

// isInTest detects whether the code is invoked inside a test environment/wrapper.
func isInTest() bool {
	if os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("TOOLS_TEST") != "" {
		return true
	}
	return false
}

const sqlcTemplate = `// Code is generated by helper script. DO NOT EDIT.
// This code is generated to replace the SQLC main codes inside {{ .SQLCOutputFileName }}
// SQLC:
//   version    : {{ .SQLCVersion }}
//   config     : {{ .SQLCConfig }}
//   sql_package: {{ .SQLPackageName }}
//   database   : {{ .DatabaseName }}

package {{ .GoPackageName }}

import (
	"context"
	"database/sql"

	"github.com/albertwidi/pkg/postgres"
)

type Queries struct {
	db *postgres.Postgres
}

// New returns a new queries instance of {{ .DatabaseName }} database.
func New(db *postgres.Postgres) *Queries {
	return &Queries{db: db}
}

// WithTransact wraps the queries inside a database transaction. The transaction will be committed if no error returned
// and automatically rolled back when an error occured.
func (q *Queries) WithTransact(ctx context.Context, iso sql.IsolationLevel, fn func(ctx context.Context, q *Queries) error) error {
	return q.db.Transact(ctx, iso, func(ctx context.Context, p *postgres.Postgres) error {
		return fn(ctx, New(p))
	})
}

// Do executes queries inside the function fn and allowed other modules to execute queries inside the same transaction scope.
func (q *Queries) Do(ctx context.Context, fn func(ctx context.Context, pg *postgres.Postgres) error ) error {
	return fn(ctx, q.db)
}
`

const dbTest = `// Code is generated by helper script. DO NOT EDIT.

package {{ .GoPackageName }}

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"
)

var (
	testQueries *Queries
	testCtx context.Context
	testHelper *TestHelper
)

func TestMain(m *testing.M) {
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
	th := NewTestHelper()
	testQueries, err = th.PrepareTest(ctx)
	if err != nil {
		code = 1
		return
	}
	// Close all resources upon exit, and record the error when closing the resources if any.
	defer func() {
		errClose := th.Close()
		if errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	code = m.Run()
	return
}
`

const testHelper = `// Code is generated by the helper script. DO NOT EDIT.

package {{ .GoPackageName }}

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/albertwidi/pkg/postgres"
	testingpkg "github.com/albertwidi/pkg/testing"
	"github.com/albertwidi/pkg/testing/pgtest"
)

type TestHelper struct {
	dbName string
	conn *postgres.Postgres
	pgtestHelper *pgtest.PGTest
}

func NewTestHelper() *TestHelper {
	return &TestHelper{
		dbName: "{{ .DatabaseName }}",
		pgtestHelper: pgtest.New(),
	}
}

// PrepareTest prepares the designated postgres database by creating the database and applying the schema. The function returns a postgres connection
// to the database that can be used for testing purposes.
func (th *TestHelper) PrepareTest(ctx context.Context) (*Queries, error) {
	// Configuration for creating and preparing the database.
	config := postgres.ConnectConfig{
		Driver:   "pgx",
		Username: "{{ .DatabaseConn.Username }}",
		Password: "{{ .DatabaseConn.Password }}",
		Host:     "{{ .DatabaseConn.Host }}",
		Port:     "{{ .DatabaseConn.Port }}",
	}
	pgconn, err := postgres.Connect(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pgtest.CreateDatabase(ctx, pgconn, th.dbName); err != nil {
		return nil, err
	}
	// Close the connection as we no-longer need it. We need it only to create the database.
	if err := pgconn.Close(); err != nil {
		return nil, err
	}

	// Create a new connection with the correct database name.
	config.DBName = th.dbName
	testConn, err := postgres.Connect(context.Background(), config)
	if err != nil {
		return nil, err
	}
	// Read the schema and apply the schema.
	repoRoot, err := testingpkg.RepositoryRoot()
	if err != nil {
		return nil, err
	}
	out, err := os.ReadFile(filepath.Join(repoRoot, "database/ledger/schema.sql"))
	if err != nil {
		return nil, err
	}
	_, err = testConn.Exec(context.Background(), string(out))
	if err != nil {
		return nil, err
	}
	// Assgign the connection for the test helper.
	th.conn = testConn
	return New(testConn), nil
}

// Close closes all connections from the test helper.
func (th *TestHelper) Close() error {
	var err error
	if th.conn != nil {
		errClose := th.conn.Close()
		if errClose != nil {
			err = errors.Join(err, errClose)
		}
	}
	errClose := th.pgtestHelper.Close()
	if errClose != nil {
		errors.Join(err ,errClose)
	}
	return err
}

// ForkPostgresSchema forks the sourceSchema with the underlying connection inside the Queries. The function will return a new connection
// with default search_path into the new schema. The schema name currently is random and cannot be defined by the user.
func (th *TestHelper) ForkPostgresSchema(ctx context.Context, q *Queries, sourceSchema string) (*Queries, error) {
	pg , err:= th.pgtestHelper.ForkSchema(ctx, q.db, sourceSchema)
	if err != nil {
		return nil, err
	}
	return New(pg), nil
}
`
