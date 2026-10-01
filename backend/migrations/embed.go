// Package migrations embeds the Postgres (13+) and MySQL (8.0+) migrations.
package migrations

import "embed"

//go:embed postgres/*.sql
var PostgresFS embed.FS

//go:embed mysql/*.sql
var MysqlFS embed.FS
