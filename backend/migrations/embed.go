// Package migrations embeds per-engine SQL migration sets. They are equivalent
// in effect but use each engine's own syntax. Minimum supported versions:
// Postgres 13+, MySQL 8.0+.
package migrations

import "embed"

//go:embed postgres/*.sql
var PostgresFS embed.FS

//go:embed mysql/*.sql
var MysqlFS embed.FS
