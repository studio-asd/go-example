package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/albertwidi/pkg/postgres"
)

type Config struct {
	// PostgresURI defines the PostgreSQL connection. The configuration will define to where
	// the we will connect the PostgreSQL. We will rewrite all host, port, username and password
	// for the underlying configurations like sqlc to match with the configuration URI.
	PostgresURI string
}

var conf Config

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	flag.Parse()
	flag.StringVar(&conf.PostgresURI, "postgres.uri", "postgres://postgres:postgres@localhost:5432/", "postgres://postgres:postgres@localhost:5432/")

	dsn, err := postgres.ParseDSN(conf.PostgresURI)
	if err != nil {
		return err
	}
	fmt.Println(dsn.DatabaseName)
	return nil
}
