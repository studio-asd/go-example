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
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "github.com/lib/pq"

	"github.com/albertwidi/pkg/postgres"
	"github.com/albertwidi/pkg/testing/pgtest"
)

//go:embed sqlc.yaml
var embeddedFiles embed.FS

type Flags struct {
	All        bool
	Replace    bool
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

	var postgresAvailable bool
	_, err = net.DialTimeout("tcp", "127.0.0.1:5432", time.Second)
	if err == nil {
		postgresAvailable = true
	}

	composeLogs := bytes.NewBuffer(nil)
	composeUpFunc := func(ctx context.Context) error {
		if postgresAvailable {
			return nil
		}
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
		if postgresAvailable {
			return nil
		}
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
		dirs, _, err := SchemaDirs(dbName, flags, ".")
		if err != nil {
			return err
		}
		if len(dirs) == 0 {
			return errors.New("no schema dirs detected")
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
				for _, dir := range dirs {
					// Read the sqlc configuration as we are using the sqlc configuration as the source of truth for database creation
					// and schema location.
					sqlcConfig, err := ReadSQLCConfiguration(dir, flags.SQLCConfig)
					if err != nil {
						return err
					}
					sqlcConfig.fileName = flags.SQLCConfig

					schemaFile := filepath.Join(dir, sqlcConfig.SQL[0].Schema)
					dbURI := sqlcConfig.SQL[0].Database.URI
					// Parse the data source name from the database URI and use that information to construct everything.
					dsn, err := postgres.ParseDSN(dbURI)
					if err != nil {
						return err
					}

					// Create the database first, as we will need to connect to the database when invoking sqlc
					// to generate go codes.
					fmt.Printf("Creating Database %s\n", dsn.DatabaseName)
					if err := pgtest.CreateDatabase(
						context.Background(),
						dbURI,
						dsn.DatabaseName,
						true,
					); err != nil {
						return err
					}

					// Applying schema to the database, we need to peek into the sqlc configuration for the schema name.
					fmt.Println("Applying schema...")
					out, err := os.ReadFile(schemaFile)
					if err != nil {
						return err
					}
					err = execQuery(
						dbURI,
						string(out),
					)
					if err != nil {
						return err
					}

					if err := os.Chdir(filepath.Join(currentDir, dir)); err != nil {
						return fmt.Errorf("failed to change directory to the database dir: %w", err)
					}
					sqlcExec := exec.Command("sqlc", "generate", "-f", flags.SQLCConfig)
					sqlcExec.Stdout = os.Stdout
					sqlcExec.Stderr = os.Stderr
					if err := sqlcExec.Run(); err != nil {
						return fmt.Errorf("failed to execute sqlc: %w", err)
					}
					if err := genTemplate(dsn, sqlcConfig); err != nil {
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
		dirs, _, err := SchemaDirs(dbName, flags, ".")
		if err != nil {
			return err
		}
		out, err := embeddedFiles.ReadFile("sqlc.yaml")
		if err != nil {
			return err
		}
		fmt.Println("DIRS", dirs)

		for _, dir := range dirs {
			// Replace the "database_name" with the directory name or the database name.
			sqlcConfig := bytes.ReplaceAll(out, []byte("database_name"), []byte(dir))

			f, err := os.OpenFile(filepath.Join(dir, "sqlc.yaml"), os.O_RDWR|os.O_CREATE, 0o666)
			if err != nil {
				return err
			}
			wd, _ := os.Getwd()
			fmt.Println(wd)

			// if flags.Replace || err == os.ErrNotExist {
			fmt.Println("WRITING CONFIGURATION")
			_, err = f.Write(sqlcConfig)
			if err != nil {
				return err
			}
			return f.Close()
			// }
			// fmt.Println("SKIPPING CONFIGURATION")
			// return f.Close()
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
	fset.BoolVar(&f.Replace, "replace", false, "replace flag to decide whether we want to replace the existing file with newly generated one")
	fset.StringVar(&f.DBName, "db_name", "", "database name")
	fset.StringVar(&f.SQLCConfig, "sqlc_config", "sqlc.yaml", "sqlc.yaml configuration")

	if err = fset.Parse(args); err != nil {
		return
	}
	return
}

func genTemplate(dsn postgres.DSN, config SQLCConfig) error {
	fmt.Printf("Generating template for database %s\n", dsn.DatabaseName)
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

	td := TemplateData{
		DatabaseName:       dsn.DatabaseName,
		SQLCVersion:        sqlcVersion,
		SQLCConfig:         config.fileName,
		SQLCOutputFileName: config.SQL[0].Gen.Go.OutputDBFileName,
		GoPackageName:      config.SQL[0].Gen.Go.Package,
		SQLPackageName:     config.SQL[0].Gen.Go.SQLPackage,
		PathToSchema:       filepath.Join("database", dsn.DatabaseName, config.SQL[0].Schema),
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
	"fmt"
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

// ensureInTransact ensures the queries are running inside the transaction scope, if the queries is not running inside the a transaction
// the function will trigger WithTransact method. While the function doesn't guarantee the subsequent function to have the same isolation
// level, but this function will return an error if the expectations and the current isolation level is incompatible.
func (q *Queries) ensureInTransact(ctx context.Context, iso sql.IsolationLevel, fn func(ctx context.Context, q *Queries) error) error {
	inTransaction, isoLevel := q.db.InTransaction()
	if !inTransaction {
		return q.WithTransact(ctx, iso, fn)
	}
	// Don't accept different isolation level between transactions as we will be getting different results.
	if iso != isoLevel {
		return fmt.Errorf("different expectations of isolation level. Got %s but expecting %s", isoLevel, iso)
	}
	return fn(ctx, q)
}

// Do executes queries inside the function fn and allowed other modules to execute queries inside the same transaction scope.
func (q *Queries) Do(ctx context.Context, fn func(ctx context.Context, pg *postgres.Postgres) error ) error {
	return fn(ctx, q.db)
}

// Postgres returns the postgres object.
func (q *Queries) Postgres() *postgres.Postgres {
	return q.db
}
`

const dbTest = `// Code is generated by helper script. DO NOT EDIT.

package {{ .GoPackageName }}

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/albertwidi/go-example/internal/testing/pghelper"
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
	dbName := "{{ .DatabaseName }}"
	testHelper, err = pghelper.New(ctx, dbName, New)
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

`
