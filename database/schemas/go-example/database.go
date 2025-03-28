package database

import "embed"

//go:embed schema.sql
var EmbeddedSchema embed.FS

const (
    // DatabaseName is the name of the database.
    DatabaseName = "go_example"
	PostgresDSN = "postgres://postgres:postgres@localhost:5432/go_example"
)
