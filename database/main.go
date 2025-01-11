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
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/studio-asd/pkg/postgres"
	"gopkg.in/yaml.v3"
)

//go:embed go_template/*
var goTemplate embed.FS

type Flags struct {
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
	Date               string
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
	fset.StringVar(&f.SQLCConfig, "sqlc_config", "sqlc.yaml", "sqlc.yaml configuration")

	if err = fset.Parse(args); err != nil {
		return
	}
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

	// Check whether the sqlc files are already generated, we want to replace the generated sqlc.go with our template.
	relativePath := config.SQL[0].Gen.Go.Out
	filePath := filepath.Join(relativePath, config.SQL[0].Gen.Go.OutputDBFileName)
	f, err := os.OpenFile(filePath, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

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
		Date: time.Now().Format(time.RFC3339),
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
