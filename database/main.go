// This tiny Go program is a tiny generator to automatically generate the test and helper function
// for the database layer.

package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/studio-asd/go-example/internal/git"
	"github.com/studio-asd/pkg/postgres"
	"gopkg.in/yaml.v3"
)

//go:embed go_template/*
var goTemplate embed.FS

type Flags struct {
	DBName      string
	DBSchemaDir string
	SQLCConfig  string
}

// SQLCConfig is the sqlc configuration structure to be used in this generator.
type SQLCConfig struct {
	fileName  string
	schemaDir string
	Version   string          `yaml:"version"`
	SQL       []SQLCSQLConfig `yaml:"sql"`
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
	DatabaseName         string
	SQLCVersion          string
	SQLCConfig           string
	SQLCOutputFileName   string
	GoPackageName        string
	SQLPackageName       string
	RelativePathToSchema string
	SchemaName           string
	SchemaDir            string
	DatabaseConn         TemplateDataDatabaseConn
	Date                 string
}

type TemplateDataDatabaseConn struct {
	DatabaseName string
	Host         string
	Port         string
	Username     string
	Password     string
}

// parseFlags parse the strinbrew updategs arguments from(for example) os.Args and returns the Flags struct.
func parseFlags(args []string) (f Flags, err error) {
	// Parse all the flags needed for code generation.
	fset := flag.NewFlagSet("global_flags", flag.ExitOnError)
	fset.StringVar(&f.DBName, "db_name", "", "database name")
	fset.StringVar(&f.DBSchemaDir, "db_schema_dir", "", "database schema directory")
	fset.StringVar(&f.SQLCConfig, "sqlc_config", "sqlc.yaml", "sqlc.yaml configuration")

	err = fset.Parse(args)
	return
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
	if len(args) < 2 {
		return errors.New("please specify the command and directory")
	}

	// Check if we have more than two(2) args which is the command. In general we only use two(2) input after the
	// command, for example main.go gengo [db_name] [dir]. Thus everything after [db_name] [dir] are flags.
	var flagArgs []string
	flagArgs = args[2:]
	flags, err := parseFlags(flagArgs)
	if err != nil {
		return err
	}

	switch args[0] {
	case "gengo":
		dir := "."
		if len(args) > 1 {
			dir = args[1]
		}
		// Change directory to the destination.
		if err := os.Chdir(dir); err != nil {
			return err
		}

		if flags.SQLCConfig == "" {
			return errors.New("sqlc configuration cannot be empty")
		}
		sqlcConfig := SQLCConfig{}

		out, err := os.ReadFile(flags.SQLCConfig)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(out, &sqlcConfig); err != nil {
			return err
		}
		// Set the filename with the base as this command can be invoked from a relative path.
		sqlcConfig.fileName = filepath.Base(flags.SQLCConfig)
		sqlcConfig.schemaDir = flags.DBSchemaDir

		if err := genGoTemplate(sqlcConfig); err != nil {
			return err
		}
	}
	return nil
}

func genGoTemplate(config SQLCConfig) error {
	// Retrieve the version of sqlc via sqlc CLI.
	cmd := exec.Command("sqlc", "version")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	sqlcVersion := string(out)
	sqlcVersion = strings.TrimRight(sqlcVersion, "\n")

	// Parsing the database DSN from SQLC configuration.
	dsn, err := postgres.ParseDSN(config.SQL[0].Database.URI)
	if err != nil {
		return err
	}
	fmt.Printf("Generating template for database %s\n", dsn.DatabaseName)

	goTemplateFile, err := goTemplate.Open("go_template/sqlc_go.tmpl")
	if err != nil {
		return err
	}
	goTemplateContent, err := io.ReadAll(goTemplateFile)
	if err != nil {
		return err
	}

	goTestTemplateFile, err := goTemplate.Open("go_template/sqlc_test_go.tmpl")
	if err != nil {
		return err
	}
	goTestTemplateContent, err := io.ReadAll(goTestTemplateFile)
	if err != nil {
		return err
	}

	dbTemplateFile, err := goTemplate.Open("go_template/db_embed_go.tmpl")
	if err != nil {
		return err
	}
	dbTemplateContent, err := io.ReadAll(dbTemplateFile)
	if err != nil {
		return err
	}

	// Check whether the sqlc files are already generated, we want to replace the generated sqlc.go with our template.
	relativePath := config.SQL[0].Gen.Go.Out

	tmpl := template.New("sqlc_template")
	tmpl, err = tmpl.Parse(string(goTemplateContent))
	if err != nil {
		return err
	}
	tmpl = tmpl.New("db_test")
	tmpl, err = tmpl.Parse(string(goTestTemplateContent))
	if err != nil {
		return err
	}
	tmpl = tmpl.New("db_embed")
	tmpl, err = tmpl.Parse(string(dbTemplateContent))
	if err != nil {
		return err
	}

	_, schemaFile := filepath.Split(config.SQL[0].Schema)

	td := TemplateData{
		DatabaseName:       dsn.DatabaseName,
		SQLCVersion:        sqlcVersion,
		SQLCConfig:         config.fileName,
		SQLCOutputFileName: config.SQL[0].Gen.Go.OutputDBFileName,
		GoPackageName:      config.SQL[0].Gen.Go.Package,
		SQLPackageName:     config.SQL[0].Gen.Go.SQLPackage,
		RelativePathToSchema: filepath.Join(
			genPathToSchemaPath(config.SQL[0].Gen.Go.Out),
			dsn.DatabaseName,
			config.SQL[0].Schema,
		),
		SchemaName: schemaFile,
		SchemaDir:  config.schemaDir,
		DatabaseConn: TemplateDataDatabaseConn{
			Username:     dsn.Username,
			Password:     dsn.Password,
			Host:         dsn.Host,
			Port:         dsn.Port,
			DatabaseName: dsn.DatabaseName,
		},
		Date: time.Now().Format(time.RFC3339),
	}

	buff := bytes.NewBuffer(nil)
	if err := tmpl.ExecuteTemplate(buff, "sqlc_template", td); err != nil {
		return err
	}

	filePath := filepath.Join(relativePath, config.SQL[0].Gen.Go.OutputDBFileName)
	f, err := os.OpenFile(filePath, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()
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

	if err := tmpl.ExecuteTemplate(buff, "db_embed", td); err != nil {
		return err
	}

	// Retrieve the repository root so we can have an absolute path to generate the database.go. We need the absolute path because this
	// program is intended to be running in the database/schema directory. Since we can have multiple sub-schema inside a big schema, we
	// better not guess on where we were at.
	repoRoot, err := git.RepositoryRoot()
	if err != nil {
		return err
	}
	dbEmbedPackagePath := filepath.Join(repoRoot, "database", "schemas", config.schemaDir, "database.go")
	if err := os.WriteFile(dbEmbedPackagePath, buff.Bytes(), 0o777); err != nil {
		return err
	}
	buff.Reset()

	return nil
}

// genPathToSchemaPath returns the relative path from the generated code to the schema. Defining the relative path is quite
// simple as we only need to go backwards from non ".." path and put them as ".." to the repository root and pinpoint them
// to the right schema path.
func genPathToSchemaPath(genPath string) string {
	dirs := strings.Split(genPath, "/")
	goToPrev := ""
	for _, dir := range dirs {
		if dir != ".." {
			goToPrev = path.Join(goToPrev, "..")
		}
	}
	goToPrev = path.Join(goToPrev, "database", "schemas")
	return goToPrev
}
