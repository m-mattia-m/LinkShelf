package migrations

import "embed"

// FS Embed the migrations directory in the binary file
//
//go:embed *.sql
var FS embed.FS
