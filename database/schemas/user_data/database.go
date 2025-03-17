package database

import "embed"

//go:embed schema.sql
var EmbeddedSchema embed.FS

const (
    // DatabaseName is the name of the database.
    DatabaseName = "user_data"
	PostgresDSN = "postgres://postgres:postgres@localhost:5432/user_data"
)
