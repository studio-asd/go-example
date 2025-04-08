package testdata

import (
	"embed"
)

//go:embed migrations/*.sql
var EmbeddedSchema embed.FS
