// Package migrations embeds per-engine SQL migration sets. The two sets are
// equivalent in effect but not line-for-line identical - each uses its
// engine's own identifier quoting and DDL syntax (e.g. Postgres's
// ON CONFLICT vs MySQL's INSERT IGNORE). Minimum supported versions:
// Postgres 13+, MySQL 8.0+.
package migrations

import "embed"

//go:embed postgres/*.sql
var PostgresFS embed.FS

//go:embed mysql/*.sql
var MysqlFS embed.FS
