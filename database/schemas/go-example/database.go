package database

import "embed"

//go:embed schema.sql
var EmbeddedSchema embed.FS
